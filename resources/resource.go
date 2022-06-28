package resources

// Resource defines a resource that Kokk can work with.
type Resource struct {
	Id      string
	Content []byte
}

//type ResourceConverter struct {
//	logger *zerolog.Logger
//}
//
//func NewResourceConverter(logger *zerolog.Logger) ResourceConverter {
//	localLogger := logger.With().Str("type", "ResourceConverter").Logger()
//	return ResourceConverter{
//		logger: &localLogger,
//	}
//}
//
//var (
//	NotUnstructured = errors.New("Received object that was not an *Unstructured")
//)
//
//func (r *ResourceConverter) ToResource(obj interface{}) (Resource, error) {
//	logger := r.logger.With().Str("method", "ToResource").Logger()
//
//	resource := Resource{}
//
//	unstructuredResource, ok := obj.(*unstructured.Unstructured)
//	if !ok {
//		logger.Error().Err(NotUnstructured).Msg("failed to cast obj to unstructured")
//		return resource, NotUnstructured
//	}
//
//	id := k.getResourceId(unstructuredResource)
//
//	data, err := json.Marshal(unstructuredResource.Object)
//	if err != nil {
//		logger.Error().Err(err).Str("id", id).Msg("Failed to convert object to JSON")
//		return
//	}
//
//	k.repository[id] = resources.Resource{
//		Id:      id,
//		Content: data,
//	}
//}
