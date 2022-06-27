package output

import (
	"fmt"
	"github.com/knadh/koanf"
	"github.com/rs/zerolog"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"regexp"
	"strings"
	"time"
)

type KubernetesOutput struct {
	config    kubernetesOutputConfig
	resources []schema.GroupVersionResource
	factory   dynamicinformer.DynamicSharedInformerFactory
	stop      chan struct{}
	logger    *zerolog.Logger
}

type kubernetesOutputConfig struct {
	ResourceTypes []string `koanf:"resources"`
	ResyncSeconds int      `koanf:"resync"`
}

func NewKubernetesOutput(config *koanf.Koanf, dynamicClient dynamic.Interface, discoveryClient *discovery.DiscoveryClient, logger *zerolog.Logger) (*KubernetesOutput, error) {
	output := KubernetesOutput{
		logger: logger,
	}

	if err := output.parseConfig(config); err != nil {
		return nil, err
	}

	if err := output.discoverGroupVersionResources(discoveryClient); err != nil {
		return nil, err
	}

	if err := output.startInformers(dynamicClient); err != nil {
		return nil, err
	}

	return &output, nil
}

func (o *KubernetesOutput) parseConfig(config *koanf.Koanf) error {
	return config.Unmarshal("kubernetes", &o.config)
}

func (o *KubernetesOutput) discoverGroupVersionResources(client *discovery.DiscoveryClient) error {
	lists, err := client.ServerPreferredResources()
	if err != nil {
		return err
	}

	o.logger.Info().Strs("types", o.config.ResourceTypes).Msg("Configured types")

configuredTypes:
	for _, configuredResourceType := range o.config.ResourceTypes {
		for _, list := range lists {
			for _, discoveredResourceType := range list.APIResources {
				if strings.EqualFold(configuredResourceType, discoveredResourceType.Kind) {
					group, version, err := o.getGroupAndVersion(list, discoveredResourceType)
					if err != nil {
						return err
					}

					o.logger.Info().Str("group", group).Str("version", version).Str("kind", discoveredResourceType.Kind).Str("name", discoveredResourceType.Name).Msg("Will monitor Kubernetes Resource type")

					o.resources = append(o.resources, schema.GroupVersionResource{
						Group:    group,
						Version:  version,
						Resource: discoveredResourceType.Name,
					})

					continue configuredTypes
				}
			}
		}

		return fmt.Errorf("the configured type %s is not available on the APIserver", configuredResourceType)
	}

	return nil
}

var groupVersionExpression = regexp.MustCompile(`^(([^/]+)/)?([^/]+)$`)

func (o *KubernetesOutput) getGroupAndVersion(list *v1.APIResourceList, resource v1.APIResource) (string, string, error) {
	group := ""
	version := ""

	matches := groupVersionExpression.FindStringSubmatch(list.GroupVersion)
	if len(matches) == 4 {
		group = matches[2]
		version = matches[3]
	}

	if resource.Group != "" {
		group = resource.Group
	}
	if resource.Version != "" {
		version = resource.Version
	}

	if version == "" {
		return "", "", fmt.Errorf("could not parse GroupVersion")
	}

	return group, version, nil
}

func (o *KubernetesOutput) startInformers(client dynamic.Interface) error {
	o.factory = dynamicinformer.NewDynamicSharedInformerFactory(client, time.Duration(o.config.ResyncSeconds)*time.Second)
	o.stop = make(chan struct{})

	for _, resource := range o.resources {
		a := o.factory.ForResource(resource)
		b, _ := a.Lister().Get("hello")
		b.GetObjectKind()
	}

	o.logger.Debug().Msg("Starting Kubernetes shared informers...")
	o.factory.Start(o.stop)
	o.logger.Debug().Msg("Waiting for Kubernetes shared informers cache to sync...")
	o.factory.WaitForCacheSync(o.stop)
	o.logger.Info().Msg("Kubernetes cache synced")

	return nil
}
