//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getAPIBaseURL waits for the real API (the rag-api container by default).
// `make test-integration` rebuilds it, so the test runs against the current code.
func getAPIBaseURL(t *testing.T) string {
	t.Helper()

	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}

	client := &http.Client{Timeout: 1 * time.Second}
	require.Eventually(t, func() bool {
		resp, err := client.Get(apiURL + "/health")
		if err != nil {
			return false
		}
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		return resp.StatusCode == http.StatusOK
	}, 30*time.Second, 500*time.Millisecond, "API must be running at %s (start via make docker-start-build)", apiURL)

	return apiURL
}

func TestRAGPipeline_E2E_Integration(t *testing.T) {
	baseURL := getAPIBaseURL(t)
	client := &http.Client{Timeout: 90 * time.Second}

	// 1. Check Server Health
	t.Run("GET /health", func(t *testing.T) {
		resp, err := client.Get(baseURL + "/health")
		require.NoError(t, err)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// 2. Check Qdrant Health
	t.Run("GET /db/health", func(t *testing.T) {
		resp, err := client.Get(baseURL + "/db/health")
		require.NoError(t, err)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var body map[string]string
		err = json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "ok", body["status"])
		assert.Equal(t, "up", body["db"])
	})

	// 3. Check Ollama Health
	t.Run("GET /llm/health", func(t *testing.T) {
		resp, err := client.Get(baseURL + "/llm/health")
		require.NoError(t, err)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var body map[string]string
		err = json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "ok", body["status"])
		assert.Equal(t, "up", body["llm"])
	})

	// 4. Ingest a specific unique document
	documentID := fmt.Sprintf("doc-e2e-%d", time.Now().UnixNano())
	textToIngest := "Project Antigravity is a specialized autonomous development system developed by Google."

	t.Run("POST /ingest", func(t *testing.T) {
		payload := map[string]string{
			"text":       textToIngest,
			"documentId": documentID,
		}
		b, err := json.Marshal(payload)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/ingest", bytes.NewBuffer(b))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer func() { assert.NoError(t, resp.Body.Close()) }()

		assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	})

	// Brief wait to ensure vector store indexes the point
	time.Sleep(1 * time.Second)

	// 5. Query with RAG and verify retrieval augmented response
	t.Run("POST /generate with RAG Context", func(t *testing.T) {
		payload := map[string]string{
			"prompt": "Who developed Project Antigravity according to the context?",
		}
		b, err := json.Marshal(payload)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/generate", bytes.NewBuffer(b))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer func() { assert.NoError(t, resp.Body.Close()) }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result struct {
			Response string `json:"response"`
		}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.NotEmpty(t, result.Response)

		// The LLM response should mention Google based on the ingested context
		assert.True(t, strings.Contains(strings.ToLower(result.Response), "google"),
			"Expected response to contain 'google' based on context, got: %s", result.Response)
	})
}
