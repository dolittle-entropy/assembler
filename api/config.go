package api

import (
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"net/http"
)

func NewConfigHandler(config *koanf.Koanf) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		data, err := config.Marshal(yaml.Parser())
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte(err.Error()))
			return
		}

		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write(data)
	}
}
