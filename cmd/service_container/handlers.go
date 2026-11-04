package service_container

import (
	"rag-service-go/internal/application"
	"rag-service-go/internal/infrastructure/chunker"
)

type Handlers struct {
	LLMHealthHandler application.LLMHealthCheckHandler
	GenerateHandler  application.GenerateHandler
	IngestHandler    application.IngestHandler
	DbHealthHandler  application.DbHealthHandler
}

func NewHandlers(
	llmClients LLMClients,
	vectorStores VectorStores,
) Handlers {
	simpleChunker := chunker.NewSimpleChunker(500)

	return Handlers{
		LLMHealthHandler: application.NewLLMHealthCheckHandler(llmClients.OllamaClient),
		GenerateHandler:  application.NewGenerateHandler(llmClients.OllamaClient, vectorStores.QdrantStore),
		IngestHandler:    application.NewIngestHandler(simpleChunker, vectorStores.QdrantStore, llmClients.OllamaClient),
		DbHealthHandler:  application.NewDbHealthHandler(vectorStores.QdrantStore),
	}
}
