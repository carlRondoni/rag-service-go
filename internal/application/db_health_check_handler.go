package application

import (
	"context"
	"rag-service-go/internal/domain"
)

type DbHealthHandler struct {
	dbClient domain.VectorStore
}

func NewDbHealthHandler(dbClient domain.VectorStore) DbHealthHandler {
	return DbHealthHandler{
		dbClient: dbClient,
	}
}

func (h DbHealthHandler) Handle(ctx context.Context) error {
	return h.dbClient.Health(ctx)
}
