package main

import (
	"rag-service-go/cmd/api/routes"
	"rag-service-go/cmd/service_container"
	"log"
	"net/http"
)

func main() {
	container := service_container.NewServiceContainer()

	routes.InitRoutes(container.Controllers)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
