package api

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/knadh/koanf"
	"github.com/rs/zerolog"
	"net/http"
	"time"
)

func NewServer(config *koanf.Koanf, logger *zerolog.Logger) (*http.Server, error) {
	handler := apiHandler{
		router: http.NewServeMux(),
		logger: logger,
	}

	if err := config.Unmarshal("server", &handler.config); err != nil {
		return nil, err
	}

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
