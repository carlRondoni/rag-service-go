//go:build unit || !integration

package llm_clients_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"rag-service-go/internal/infrastructure/llm_clients"
	"testing"

	"github.com/rs/zerolog"
)

func newTestClient(serverURL string) llm_clients.OllamaClient {
	logger := zerolog.Nop()
	return llm_clients.NewOllamaClient(serverURL, "test-model", "test-embed-model", logger)
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

func TestGenerate_DisablesStreaming(t *testing.T) {
	var body map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		_, _ = fmt.Fprint(w, `{"response":"hello"}`)
	}))
	defer srv.Close()

	if _, err := newTestClient(srv.URL).Generate(context.Background(), "hi"); err != nil {
		t.Fatal(err)
	}

	if body["stream"] != false {
		t.Fatalf("expected stream=false, got %v", body["stream"])
	}
	if body["model"] != "test-model" {
		t.Fatalf("expected generation model, got %v", body["model"])
	}
}

func TestEmbed_UsesEmbedEndpointAndModel(t *testing.T) {
	var path string
	var body map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		_, _ = fmt.Fprint(w, `{"embeddings":[[0.1,0.2]]}`)
	}))
	defer srv.Close()

	embs, err := newTestClient(srv.URL).Embed(context.Background(), []string{"text"})
	if err != nil {
		t.Fatal(err)
	}

	if path != "/api/embed" {
		t.Fatalf("unexpected path %s", path)
	}
	if body["model"] != "test-embed-model" || body["input"] != "text" {
		t.Fatalf("unexpected body %v", body)
	}
	if len(embs) != 1 || len(embs[0].Vector) != 2 {
		t.Fatalf("unexpected embeddings %v", embs)
	}
}

func TestEmbed_EmptyVectorReturnsError(t *testing.T) {
	for _, response := range []string{`{"embeddings":[]}`, `{"embeddings":[[]]}`} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprint(w, response)
		}))

		_, err := newTestClient(srv.URL).Embed(context.Background(), []string{"text"})
		srv.Close()

		if err == nil {
			t.Fatalf("expected error for response %s", response)
		}
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
