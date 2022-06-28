package debug

import (
	"bytes"
	"dolittle.io/kokk/api/utils"
	"encoding/json"
	"net/http"
	"sort"
)

func NewDebugHandler(input, output Repository) (http.Handler, error) {
	handler := http.NewServeMux()

	list, err := utils.NewTemplateHandler("api/debug/list.html", func(r *http.Request) (any, error) {
		resources := output.List()
		ids := make([]string, 0, len(resources))
		for _, resource := range resources {
			ids = append(ids, resource.Id)
		}
		sort.Strings(ids)

		return listData{
			IDs: ids,
		}, nil
	})
	if err != nil {
		return nil, err
	}

	view, err := utils.NewTemplateHandler("api/debug/view.html", func(r *http.Request) (any, error) {
		resourceID := r.URL.Path

		inputContent := ""
		if resource, err := input.Get(resourceID); err == nil {
			var pretty bytes.Buffer
			if err := json.Indent(&pretty, resource.Content, "", "  "); err != nil {
				return nil, err
			}
			inputContent = pretty.String()
		}

		outputContent := ""
		if resource, err := output.Get(resourceID); err == nil {
			var pretty bytes.Buffer
			if err := json.Indent(&pretty, resource.Content, "", "  "); err != nil {
				return nil, err
			}
			outputContent = pretty.String()
		}

		return viewData{
			ID:            resourceID,
			InputContent:  inputContent,
			OutputContent: outputContent,
		}, nil
	})
	if err != nil {
		return nil, err
	}

	handler.Handle("/debug/list", list)
	handler.Handle("/debug/view/", http.StripPrefix("/debug/view/", view))
	handler.Handle("/debug/", http.RedirectHandler("/debug/list", http.StatusTemporaryRedirect))

	return handler, nil
}

type listData struct {
	IDs []string
}

type viewData struct {
	ID            string
	InputContent  string
	OutputContent string
}
