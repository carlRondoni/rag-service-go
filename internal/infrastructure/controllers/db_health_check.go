package controllers

import (
	"encoding/json"
	"net/http"
	"rag-service-go/internal/application"

	"github.com/rs/zerolog"
)

type DbHealthController struct {
	handler application.DbHealthHandler
	logger  zerolog.Logger
}

func NewDbHealthController(handler application.DbHealthHandler, logger zerolog.Logger) DbHealthController {
	return DbHealthController{
		handler: handler,
		logger:  logger,
	}
}

func (c DbHealthController) Execute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		c.logger.Error().Msg("db health check error: method not allowed")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := c.handler.Handle(r.Context())
	if err != nil {
		c.logger.Error().Err(err).Msg("db health check error: db down")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]any{
			"status": "degraded",
			"db":     "down",
		})
		return
	}

	c.logger.Info().Msg("db health check OK")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"status": "ok",
		"db":     "up",
	})
}
