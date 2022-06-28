package input

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"dolittle.io/kokk/resources"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func ListResources(dir string) ([]resources.Resource, error) {
	resourceSlice := []resources.Resource{}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return resourceSlice, err
	}

	for _, file := range files {
		fmt.Println(file.Name())
		data, err := os.ReadFile(path.Join(dir, file.Name()))
		if err != nil {
			return resourceSlice, err
		}
		// var inputMap map[string]interface{}
		var obj interface{}
		err = json.Unmarshal(data, &obj)
		if err != nil {
			return resourceSlice, err
		}

		resource, ok := obj.(*unstructured.Unstructured)
		if !ok {
			// logger.Error().Msg("Received object that was not an *Unstructured")
			return resourceSlice, errors.New("Received object that was not an *Unstructured")
		}

		if err != nil {
			return resourceSlice, err
		}
		resource := resources.Resource{
			Id:      inputMap["metadata"].(map[string]interface{})["selfLink"].(string),
			Content: data,
		}
		resourceSlice = append(resourceSlice, resource)
	}

	return resourceSlice, nil
}
