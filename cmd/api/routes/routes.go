package routes

import (
	"net/http"
	"rag-service-go/cmd/service_container"
)

func InitRoutes(controllers service_container.Controllers) {
	// health checks
	http.Handle("/health", http.HandlerFunc(controllers.HealthCheckController.Execute))
	http.Handle("/health/llm", http.HandlerFunc(controllers.LlmHealthCheckController.Execute))

	/*
		http.Handle("/ingest", http.HandlerFunc(controllers.IngestController.Execute))
	*/

	http.Handle("/llm/generate", http.HandlerFunc(controllers.GenerateController.Execute))
}
