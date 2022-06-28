package kubernetes

import (
	"dolittle.io/kokk/resources"
	"encoding/json"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"path"
)

type TypeProvider interface {
	IsNamespaced(gvk schema.GroupVersionKind) (bool, error)
	GetGroupVersionResource(gvk schema.GroupVersionKind) (schema.GroupVersionResource, error)
}

type ResourceConverter struct {
	types TypeProvider
}

func NewResourceConverter(types TypeProvider) *ResourceConverter {
	return &ResourceConverter{
		types: types,
	}
}

func (rc *ResourceConverter) GetIdFor(object *unstructured.Unstructured) (string, error) {
	gvk := object.GroupVersionKind()

	namespaced, err := rc.types.IsNamespaced(gvk)
	if err != nil {
		return "", err
	}
	gvr, err := rc.types.GetGroupVersionResource(gvk)
	if err != nil {
		return "", err
	}

	if namespaced {
		return path.Join(gvr.GroupVersion().String(), "namespaces", object.GetNamespace(), gvr.Resource, object.GetName()), nil
	}

	return path.Join(gvr.GroupVersion().String(), gvr.Resource, object.GetName()), nil
}

func (rc *ResourceConverter) Convert(object *unstructured.Unstructured) (*resources.Resource, error) {
	id, err := rc.GetIdFor(object)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(object.Object)
	if err != nil {
		return nil, err
	}

	return &resources.Resource{
		Id:      id,
		Content: data,
	}, nil
}
