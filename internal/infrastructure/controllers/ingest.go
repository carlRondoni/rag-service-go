package controllers

import (
	"encoding/json"
	"net/http"
	"rag-service-go/internal/application"

	"github.com/rs/zerolog"
)

type IngestController struct {
	handler application.IngestHandler
	logger  zerolog.Logger
}

type IngestRequest struct {
	Text       string `json:"text"`
	DocumentID string `json:"documentId"`
}

func NewIngestController(
	handler application.IngestHandler,
	logger zerolog.Logger,
) IngestController {
	return IngestController{
		handler: handler,
		logger:  logger,
	}
}

func (c IngestController) Execute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.logger.Error().Msg("ingest error: method not allowed")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req IngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.logger.Error().Err(err).Msg("ingest error: invalid json body")
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	if req.Text == "" {
		c.logger.Error().Msg("ingest error: text required")
		http.Error(w, "text required", http.StatusBadRequest)
		return
	}

	if req.DocumentID == "" {
		req.DocumentID = "default-doc"
		c.logger.Info().Msg("Using default-document")
	}

	err := c.handler.Handle(r.Context(), req.Text, req.DocumentID)
	if err != nil {
		c.logger.Error().Err(err).Msg("ingest error: error on handler: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})

		return
	}

	c.logger.Info().Msg("ingest OK")
	w.WriteHeader(http.StatusAccepted)
}
