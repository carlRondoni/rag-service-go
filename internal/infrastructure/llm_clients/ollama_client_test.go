package llm_clients_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"rag-service-go/internal/infrastructure/llm_clients"
	"testing"

	"github.com/rs/zerolog"
)

func newTestClient(serverURL string) llm_clients.OllamaClient {
	logger := zerolog.Nop()
	return llm_clients.NewOllamaClient(serverURL, "test-model", logger)
}

func TestHealth_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)

	if err := client.Health(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestHealth_500(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)

	err := client.Health(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}
