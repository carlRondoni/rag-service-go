package controllers

import (
	"encoding/json"
	"net/http"
	"rag-service-go/internal/application"

	"github.com/rs/zerolog"
)

type LlmHealthCheckController struct {
	handler application.LLMHealthCheckHandler
	logger  zerolog.Logger
}

func NewLlmHealthCheckController(
	handler application.LLMHealthCheckHandler,
	logger zerolog.Logger,
) LlmHealthCheckController {
	return LlmHealthCheckController{
		handler: handler,
		logger:  logger,
	}
}

func (c LlmHealthCheckController) Execute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		c.logger.Error().Msg("llm health check error: method not allowed")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := c.handler.Handle(r.Context())
	if err != nil {
		c.logger.Error().Err(err).Msg("llm health check error: llm down")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]any{
			"status": "degraded",
			"llm":    "down",
		})
		return
	}

	c.logger.Info().Msg("llm health check OK")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"status": "ok",
		"llm":    "up",
	})
}
