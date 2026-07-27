package service_container

import (
	"os"
	"rag-service-go/internal/domain"
	"rag-service-go/internal/infrastructure/vector_store"

	"github.com/rs/zerolog"
)

type VectorStores struct {
	QdrantStore domain.VectorStore
}

func NewVectorStores(logger zerolog.Logger) VectorStores {
	baseURL := os.Getenv("QDRANT_URL")
	if baseURL == "" {
		logger.Fatal().Msg("QDRANT_URL is not set")
	}

	collection := os.Getenv("QDRANT_COLLECTION")
	if collection == "" {
		logger.Fatal().Msg("QDRANT_COLLECTION is not set")
	}
	
	return VectorStores{
		QdrantStore: vector_store.NewQdrantStore(
			baseURL,
			collection,
			logger,
		),
	}
}
