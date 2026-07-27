package service_container

import "github.com/rs/zerolog"

type ServiceContainer struct {
	Controllers
	Logger zerolog.Logger
}

func NewServiceContainer() ServiceContainer {
	logger := InitLogs()
	llmClients := NewLLMClients(logger)
	vectorStores := NewVectorStores(logger)
	handlers := NewHandlers(llmClients, vectorStores)
	controllers := NewControllers(handlers, logger)

	return ServiceContainer{
		Controllers: controllers,
		Logger:      logger,
	}
}
