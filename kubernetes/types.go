package kubernetes

import (
	"fmt"
	"github.com/knadh/koanf"
	"github.com/rs/zerolog"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"strings"
)

type kubernetesType struct {
	gkv      schema.GroupVersionKind
	gvr      schema.GroupVersionResource
	resource v1.APIResource
}

type Types struct {
	configuredTypes []string
	discoveredTypes []kubernetesType
	logger          *zerolog.Logger
}

func NewDiscoveredTypes(config *koanf.Koanf, client *discovery.DiscoveryClient, logger *zerolog.Logger) (*Types, error) {
	types := &Types{
		configuredTypes: config.Strings("kubernetes.resources"),
		logger:          logger,
	}

	if err := types.discoverGroupVersionResources(client); err != nil {
		return nil, err
	}

	return types, nil
}

func (t *Types) DiscoveredGVRs() []schema.GroupVersionResource {
	gvrs := make([]schema.GroupVersionResource, 0, len(t.discoveredTypes))
	for _, discoveredType := range t.discoveredTypes {
		gvrs = append(gvrs, discoveredType.gvr)
	}
	return gvrs
}

func (t *Types) IsNamespaced(gvk schema.GroupVersionKind) (bool, error) {
	tpe, err := t.findDiscoveredType(gvk)
	if err != nil {
		return false, err
	}

	return tpe.resource.Namespaced, nil
}

func (t *Types) GetGroupVersionResource(gvk schema.GroupVersionKind) (schema.GroupVersionResource, error) {
	tpe, err := t.findDiscoveredType(gvk)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}

	return tpe.gvr, nil
}

func (t *Types) discoverGroupVersionResources(client *discovery.DiscoveryClient) error {
	lists, err := client.ServerPreferredResources()
	if err != nil {
		return err
	}

	t.logger.Info().Strs("types", t.configuredTypes).Msg("Configured types")

configuredTypes:
	for _, configuredResourceType := range t.configuredTypes {
		for _, list := range lists {
			listGV, err := schema.ParseGroupVersion(list.GroupVersion)
			if err != nil {
				return err
			}

			for _, discoveredResourceType := range list.APIResources {
				if strings.EqualFold(configuredResourceType, discoveredResourceType.Kind) {
					resourceGV := listGV
					if discoveredResourceType.Group != "" {
						resourceGV = schema.GroupVersion{
							Group:   discoveredResourceType.Group,
							Version: resourceGV.Version,
						}
					}
					if discoveredResourceType.Version != "" {
						resourceGV = schema.GroupVersion{
							Group:   resourceGV.Group,
							Version: discoveredResourceType.Version,
						}
					}

					t.logger.Info().Str("group", resourceGV.Group).Str("version", resourceGV.Version).Str("kind", discoveredResourceType.Kind).Str("name", discoveredResourceType.Name).Msg("Will monitor Kubernetes Resource type")

					t.discoveredTypes = append(t.discoveredTypes, kubernetesType{
						gkv:      resourceGV.WithKind(discoveredResourceType.Kind),
						gvr:      resourceGV.WithResource(discoveredResourceType.Name),
						resource: discoveredResourceType,
					})

					continue configuredTypes
				}
			}
		}

		return fmt.Errorf("the configured type %s is not available on the APIserver", configuredResourceType)
	}

	return nil
}

func (t *Types) findDiscoveredType(gvk schema.GroupVersionKind) (*kubernetesType, error) {
	for _, discoveredType := range t.discoveredTypes {
		if discoveredType.gkv.Group != gvk.Group {
			continue
		}
		if discoveredType.gkv.Version != gvk.Version {
			continue
		}
		if discoveredType.gkv.Kind != gvk.Kind {
			continue
		}

		return &discoveredType, nil
	}

	return nil, GroupVersionKindUnknown
}
