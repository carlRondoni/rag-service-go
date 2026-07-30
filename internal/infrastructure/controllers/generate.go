package controllers

import (
	"encoding/json"
	"rag-service-go/internal/application"
	"net/http"

	"github.com/rs/zerolog"
)

type GenerateController struct {
	handler application.GenerateHandler
	logger  zerolog.Logger
}

type GenerateRequest struct {
	Prompt string `json:"prompt"`
}

func NewGenerateController(
	handler application.GenerateHandler,
	logger zerolog.Logger,
) GenerateController {
	return GenerateController{
		handler: handler,
		logger:  logger,
	}
}

func (c GenerateController) Execute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.logger.Error().Msg("generate error: method not allowed")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.logger.Error().Err(err).Msg("generate error: invalid json body")
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	if req.Prompt == "" {
		c.logger.Error().Msg("generate error: prompt required")
		http.Error(w, "prompt required", http.StatusBadRequest)
		return
	}

	resp, err := c.handler.Handle(r.Context(), req.Prompt)
	if err != nil {
		c.logger.Error().Err(err).Msg("generate error: error on handler: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		err = json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		if err != nil {
			c.logger.Error().Err(err).Msg("generate error: failed to encode response")
			return
		}

		return
	}

	c.logger.Info().Msg("generate OK")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]string{
		"response": resp,
	})
	if err != nil {
		c.logger.Error().Err(err).Msg("generate error: failed to encode response")
		return
	}
}
