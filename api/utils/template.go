package utils

import (
	"html/template"
	"net/http"
)

type TemplateHandler struct {
	template *template.Template
	handler  func(r *http.Request) (any, error)
}

func NewTemplateHandler(path string, handler func(r *http.Request) (any, error)) (http.Handler, error) {
	parsed, err := template.ParseFiles(path)
	if err != nil {
		return nil, err
	}

	return &TemplateHandler{
		template: parsed,
		handler:  handler,
	}, nil
}

func (t *TemplateHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	data, err := t.handler(request)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = writer.Write([]byte(err.Error()))
		return
	}

	writer.WriteHeader(http.StatusOK)
	_ = t.template.Execute(writer, data)
}
