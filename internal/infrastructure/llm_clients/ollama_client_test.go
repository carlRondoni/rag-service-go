package llm_clients_test

import (
	"context"
	"fmt"
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

func TestGenerate_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := fmt.Fprint(w, `{"response":"hello"}`)
		if err != nil {
			t.Fatal(err)
		}
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)

	res, err := client.Generate(context.Background(), "hi")
	if err != nil {
		t.Fatal(err)
	}

	if res != "hello" {
		t.Fatalf("expected hello got %s", res)
	}
}

func TestGenerate_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)

	_, err := client.Generate(context.Background(), "fail")
	if err == nil {
		t.Fatal("expected error")
	}
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
