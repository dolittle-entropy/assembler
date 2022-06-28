package output

import (
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"

	"dolittle.io/kokk/resources"
	"github.com/knadh/koanf"
	"github.com/rs/zerolog"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
)

type KubernetesOutput struct {
	config     kubernetesOutputConfig
	resources  []kubernetesOutputResource
	factory    dynamicinformer.DynamicSharedInformerFactory
	stop       chan struct{}
	repository map[string]resources.Resource
	logger     *zerolog.Logger
}

type kubernetesOutputConfig struct {
	ResourceTypes []string `koanf:"resources"`
	ResyncSeconds int      `koanf:"resync"`
}

type kubernetesOutputResource struct {
	gvr        schema.GroupVersionResource
	namespaced bool
}

func NewKubernetesOutput(config *koanf.Koanf, dynamicClient dynamic.Interface, discoveryClient *discovery.DiscoveryClient, logger *zerolog.Logger) (*KubernetesOutput, error) {
	output := KubernetesOutput{
		repository: make(map[string]resources.Resource),
		logger:     logger,
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

func (o *KubernetesOutput) Get(id string) (*resources.Resource, error) {
	if resource, found := o.repository[id]; found {
		return &resource, nil
	}

	return nil, ResourceNotFound
}

func (o *KubernetesOutput) List() []resources.Resource {
	list := make([]resources.Resource, 0, len(o.repository))
	for _, resource := range o.repository {
		list = append(list, resource)
	}
	return list
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

					o.resources = append(o.resources, kubernetesOutputResource{
						gvr: schema.GroupVersionResource{
							Group:    group,
							Version:  version,
							Resource: discoveredResourceType.Name,
						},
						namespaced: discoveredResourceType.Namespaced,
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
		informer := o.factory.ForResource(resource.gvr).Informer()
		logger := o.logger.With().Str("group", resource.gvr.Group).Str("version", resource.gvr.Version).Str("resource", resource.gvr.Resource).Logger()
		handler := &kubernetesOutputResourceEventHandler{
			resource:   resource,
			repository: o.repository,
			logger:     &logger,
		}
		informer.AddEventHandler(handler)
	}

	o.logger.Debug().Msg("Starting Kubernetes shared informers...")
	o.factory.Start(o.stop)
	o.logger.Debug().Msg("Waiting for Kubernetes shared informers cache to sync...")
	o.factory.WaitForCacheSync(o.stop)
	o.logger.Info().Msg("Kubernetes cache synced")

	return nil
}

type kubernetesOutputResourceEventHandler struct {
	resource   kubernetesOutputResource
	repository map[string]resources.Resource
	logger     *zerolog.Logger
}

func (k *kubernetesOutputResourceEventHandler) OnAdd(obj interface{}) {
	logger := k.logger.With().Str("method", "OnAdd").Logger()

	resource, ok := obj.(*unstructured.Unstructured)
	if !ok {
		logger.Error().Msg("Received object that was not an *Unstructured")
		return
	}

	id := k.getResourceId(resource)

	data, err := json.Marshal(resource.Object)
	if err != nil {
		logger.Error().Err(err).Str("id", id).Msg("Failed to convert object to JSON")
		return
	}

	k.repository[id] = resources.Resource{
		Id:      id,
		Content: data,
	}

	logger.Trace().Str("id", id).Msg("Added resource to repository")
}

func (k *kubernetesOutputResourceEventHandler) OnUpdate(_, newObj interface{}) {
	k.OnAdd(newObj)
}

func (k *kubernetesOutputResourceEventHandler) OnDelete(obj interface{}) {
	logger := k.logger.With().Str("method", "OnDelete").Logger()

	resource, ok := obj.(*unstructured.Unstructured)
	if !ok {
		logger.Error().Msg("Received object that was not an *Unstructured")
		return
	}

	id := k.getResourceId(resource)
	delete(k.repository, id)

	logger.Trace().Str("id", id).Msg("Deleted resource from repository")
}

func (k *kubernetesOutputResourceEventHandler) getResourceId(resource *unstructured.Unstructured) string {
	group := k.resource.gvr.Group
	version := k.resource.gvr.Version
	resourceType := k.resource.gvr.Resource

	if k.resource.namespaced {
		if group == "" {
			return path.Join(version, "namespaces", resource.GetNamespace(), resourceType, resource.GetName())
		}
		return path.Join(group, version, "namespaces", resource.GetNamespace(), resourceType, resource.GetName())
	}
	if group == "" {
		return path.Join(version, resourceType, resource.GetName())
	}
	return path.Join(group, version, resourceType, resource.GetName())
}
