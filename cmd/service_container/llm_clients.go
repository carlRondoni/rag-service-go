package service_container

import (
	"rag-service-go/internal/infrastructure/llm_clients"
	"os"

	"github.com/rs/zerolog"
)

type LLMClients struct {
	OllamaClient llm_clients.OllamaClient
}

func NewLLMClients(logger zerolog.Logger) LLMClients {
	url := os.Getenv("OLLAMA_URL")
	if url == "" {
		logger.Fatal().Msg("OLLAMA_URL is not set")
	}

	return LLMClients{
		OllamaClient: llm_clients.NewOllamaClient(url, "llama3", logger),
	}
}
