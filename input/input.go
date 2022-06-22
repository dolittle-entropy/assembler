package input

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

type Resource struct {
	Id      string
	Content string
}

func ListResources(dir string) ([]Resource, error) {
	resources := []Resource{}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return resources, err
	}

	for _, file := range files {
		fmt.Println(file.Name())
		data, err := os.ReadFile(path.Join(dir, file.Name()))
		if err != nil {
			return resources, err
		}
		var inputMap map[string]interface{}
		err = json.Unmarshal(data, &inputMap)
		if err != nil {
			return resources, err
		}

		if err != nil {
			return resources, err
		}
		resource := Resource{
			Id:      inputMap["metadata"].(map[string]interface{})["selfLink"].(string),
			Content: string(data),
		}
		resources = append(resources, resource)
	}

	return resources, nil
}
