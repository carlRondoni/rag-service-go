package controllers

import (
	"net/http"

	"github.com/rs/zerolog"
)

type HealthCheckController struct{
	logger zerolog.Logger
}

func NewHealthCheckController(logger zerolog.Logger) HealthCheckController {
	return HealthCheckController{
		logger: logger,
	}
}

func (c HealthCheckController) Execute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		c.logger.Error().Msg("health check error: method not allowed")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	c.logger.Info().Msg("health check OK")
	w.WriteHeader(http.StatusOK)
}
