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

	t.Run("GET /health", func(t *testing.T) {
		resp, err := client.Get(baseURL + "/health")
		require.NoError(t, err)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

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

	documentID := fmt.Sprintf("doc-e2e-%d", time.Now().UnixNano())
	textToIngest := "Blood Lotus of Kzarr is a gonzo old-school RPG zine where you draw Magic: The Gathering cards to roll mutations, written by a wizard called Thrax the Unwashed."

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

	// let qdrant index the point
	time.Sleep(1 * time.Second)

	t.Run("POST /generate with RAG Context", func(t *testing.T) {
		payload := map[string]string{
			"prompt": "Who wrote Blood Lotus of Kzarr according to the context?",
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

		assert.True(t, strings.Contains(strings.ToLower(result.Response), "thrax"),
			"expected the answer to mention thrax, got: %s", result.Response)
	})
}
