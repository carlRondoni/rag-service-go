package routes

import (
	"net/http"
	"rag-service-go/cmd/service_container"
)

func InitRoutes(controllers service_container.Controllers) {
	// basic routes
	http.Handle("/health", http.HandlerFunc(controllers.HealthCheckController.Execute))

	// RAG routes
	http.Handle("/ingest", http.HandlerFunc(controllers.IngestController.Execute))
	http.Handle("/db/health", http.HandlerFunc(controllers.DbHealthController.Execute))

	// LLM routes
	http.Handle("/llm/generate", http.HandlerFunc(controllers.GenerateController.Execute))
	http.Handle("/llm/health", http.HandlerFunc(controllers.LlmHealthCheckController.Execute))

}
