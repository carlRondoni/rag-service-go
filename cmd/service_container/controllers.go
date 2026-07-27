package service_container

import (
	"rag-service-go/internal/infrastructure/controllers"

	"github.com/rs/zerolog"
)

type Controllers struct {
	HealthCheckController    controllers.HealthCheckController
	LlmHealthCheckController controllers.LlmHealthCheckController
	GenerateController       controllers.GenerateController
	IngestController         controllers.IngestController
	DbHealthController       controllers.DbHealthController
}

func NewControllers(handlers Handlers, logger zerolog.Logger) Controllers {
	return Controllers{
		HealthCheckController:    controllers.NewHealthCheckController(logger),
		LlmHealthCheckController: controllers.NewLlmHealthCheckController(handlers.LLMHealthHandler, logger),
		GenerateController:       controllers.NewGenerateController(handlers.GenerateHandler, logger),
		IngestController:         controllers.NewIngestController(handlers.IngestHandler, logger),
		DbHealthController:       controllers.NewDbHealthController(handlers.DbHealthHandler, logger),
	}
}
