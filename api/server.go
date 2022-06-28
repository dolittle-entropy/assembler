package api

import (
	"context"
	"dolittle.io/kokk/api/debug"
	"dolittle.io/kokk/api/utils"
	"fmt"
	"github.com/google/uuid"
	"github.com/knadh/koanf"
	"github.com/rs/zerolog"
	"net/http"
	"time"
)

func NewServer(config *koanf.Koanf, input, output debug.Repository, logger *zerolog.Logger) (*http.Server, error) {
	handler := apiHandler{
		router: http.NewServeMux(),
		logger: logger,
	}

	if err := config.Unmarshal("server", &handler.config); err != nil {
		return nil, err
	}

	index, err := utils.NewTemplateHandler("api/index.html", func(r *http.Request) (any, error) {
		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	conf := NewConfigHandler(config)

	ui, err := debug.NewDebugHandler(input, output)
	if err != nil {
		return nil, err
	}

	handler.router.Handle("/", index)
	handler.router.Handle("/config", conf)
	handler.router.Handle("/debug/", ui)

	logger.Info().Int("port", handler.config.Port).Msg("API Server configured")

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", handler.config.Port),
		Handler: &handler,
	}, nil
}

type apiHandler struct {
	config apiHandlerConfig
	router *http.ServeMux
	logger *zerolog.Logger
}

type apiHandlerConfig struct {
	Port int `koanf:"port"`
}

func (a *apiHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	requestID := uuid.NewString()
	start := time.Now()

	a.logger.Info().Str("request_id", requestID).Str("method", request.Method).Str("path", request.URL.Path).Msg("Started handling request")

	idContext := context.WithValue(request.Context(), "request_id", requestID)
	requestWithIdContext := request.WithContext(idContext)
	a.router.ServeHTTP(writer, requestWithIdContext)

	elapsed := time.Since(start)
	a.logger.Info().Str("request_id", requestID).Float64("duration", elapsed.Seconds()).Msg("Finished handling request")
}
