package input

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"dolittle.io/kokk/resources"
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
		var inputMap map[string]interface{}
		err = json.Unmarshal(data, &inputMap)
		if err != nil {
			return resourceSlice, err
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
