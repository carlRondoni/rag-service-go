package service_container

import "rag-service-go/internal/application"

type Handlers struct {
	LLMHealthHandler application.LLMHealthCheckHandler
	GenerateHandler  application.GenerateHandler
	IngestHandler    application.IngestHandler
}

func NewHandlers(llmClients LLMClients) Handlers {
	return Handlers{
		LLMHealthHandler: application.NewLLMHealthCheckHandler(llmClients.OllamaClient),
		GenerateHandler:  application.NewGenerateHandler(llmClients.OllamaClient),
		IngestHandler:    application.NewIngestHandler(),
	}
}
