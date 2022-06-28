package output

import (
	"time"

	"dolittle.io/kokk/resources"
	"github.com/knadh/koanf"
	"github.com/rs/zerolog"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
)

type TypeDiscoverer interface {
	DiscoveredGVRs() []schema.GroupVersionResource
}

type TypeConverter interface {
	GetIdFor(object *unstructured.Unstructured) (string, error)
	Convert(object *unstructured.Unstructured) (*resources.Resource, error)
}

type KubernetesOutput struct {
	resyncSeconds int
	types         TypeDiscoverer
	handler       kubernetesOutputHandler
	stop          chan struct{}
	logger        *zerolog.Logger
}

func NewKubernetesOutput(config *koanf.Koanf, types TypeDiscoverer, converter TypeConverter, client dynamic.Interface, logger *zerolog.Logger) (*KubernetesOutput, error) {
	output := KubernetesOutput{
		resyncSeconds: config.Int("kubernetes.resync"),
		types:         types,
		handler: kubernetesOutputHandler{
			repository: make(map[string]resources.Resource),
			converter:  converter,
			logger:     logger,
		},
		logger: logger,
	}

	if err := output.startInformers(client); err != nil {
		return nil, err
	}

	return &output, nil
}

func (o *KubernetesOutput) Get(id string) (*resources.Resource, error) {
	if resource, found := o.handler.repository[id]; found {
		return &resource, nil
	}

	return nil, ResourceNotFound
}

func (o *KubernetesOutput) List() []resources.Resource {
	list := make([]resources.Resource, 0, len(o.handler.repository))
	for _, resource := range o.handler.repository {
		list = append(list, resource)
	}
	return list
}

func (o *KubernetesOutput) startInformers(client dynamic.Interface) error {
	factory := dynamicinformer.NewDynamicSharedInformerFactory(client, time.Duration(o.resyncSeconds)*time.Second)
	o.stop = make(chan struct{})

	for _, gvr := range o.types.DiscoveredGVRs() {
		informer := factory.ForResource(gvr).Informer()
		informer.AddEventHandler(&o.handler)
	}

	o.logger.Debug().Msg("Starting Kubernetes shared informers...")
	factory.Start(o.stop)
	o.logger.Debug().Msg("Waiting for Kubernetes shared informers cache to sync...")
	factory.WaitForCacheSync(o.stop)
	o.logger.Info().Msg("Kubernetes cache synced")

	return nil
}

type kubernetesOutputHandler struct {
	repository map[string]resources.Resource
	converter  TypeConverter
	logger     *zerolog.Logger
}

func (oh *kubernetesOutputHandler) OnAdd(obj interface{}) {
	logger := oh.logger.With().Str("method", "OnAdd").Logger()

	resource, ok := obj.(*unstructured.Unstructured)
	if !ok {
		logger.Error().Msg("Received object that was not an *Unstructured")
		return
	}

	gvk := resource.GroupVersionKind()
	logger = logger.With().Str("group", gvk.Group).Str("version", gvk.Version).Str("kind", gvk.Kind).Logger()

	converted, err := oh.converter.Convert(resource)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to convert resource")
		return
	}

	oh.repository[converted.Id] = *converted
	logger.Trace().Str("id", converted.Id).Msg("Added resource to repository")
}

func (oh *kubernetesOutputHandler) OnUpdate(_, newObj interface{}) {
	oh.OnAdd(newObj)
}

func (oh *kubernetesOutputHandler) OnDelete(obj interface{}) {
	logger := oh.logger.With().Str("method", "OnDelete").Logger()

	resource, ok := obj.(*unstructured.Unstructured)
	if !ok {
		logger.Error().Msg("Received object that was not an *Unstructured")
		return
	}

	gvk := resource.GroupVersionKind()
	logger = logger.With().Str("group", gvk.Group).Str("version", gvk.Version).Str("kind", gvk.Kind).Logger()

	id, err := oh.converter.GetIdFor(resource)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get id for resource")
		return
	}

	delete(oh.repository, id)
	logger.Trace().Str("id", id).Msg("Removed resource from repository")
}
