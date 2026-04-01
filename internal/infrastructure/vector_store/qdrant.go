package vector_store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"rag-service-go/internal/domain"
)

type QdrantStore struct {
	baseURL    string
	collection string
	client     *http.Client
}

func NewQdrantStore(baseURL, collection string) *QdrantStore {
	return &QdrantStore{
		baseURL:    baseURL,
		collection: collection,
		client:     &http.Client{},
	}
}

func (q *QdrantStore) UpsertChunks(
	ctx context.Context,
	chunks []domain.Chunk,
	embeddings []domain.Embedding,
) error {

	type point struct {
		ID      string                 `json:"id"`
		Vector  []float32              `json:"vector"`
		Payload map[string]interface{} `json:"payload"`
	}

	points := make([]point, 0, len(chunks))

	for i, chunk := range chunks {
		points = append(points, point{
			ID:     chunk.ID,
			Vector: embeddings[i].Vector,
			Payload: map[string]interface{}{
				"text":        chunk.Text,
				"document_id": chunk.DocumentID,
			},
		})
	}

	body := map[string]interface{}{
		"points": points,
	}

	b, _ := json.Marshal(body)

	url := fmt.Sprintf("%s/collections/%s/points", q.baseURL, q.collection)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	resp, err := q.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("qdrant error: %d", resp.StatusCode)
	}

	return nil
}

func (q *QdrantStore) Search(
	ctx context.Context,
	query domain.Embedding,
	topK int,
) ([]domain.SearchResult, error) {

	body := map[string]interface{}{
		"vector": query.Vector,
		"limit":  topK,
	}

	b, _ := json.Marshal(body)

	url := fmt.Sprintf("%s/collections/%s/points/search", q.baseURL, q.collection)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	resp, err := q.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw struct {
		Result []struct {
			ID      string  `json:"id"`
			Score   float32 `json:"score"`
			Payload struct {
				Text       string `json:"text"`
				DocumentID string `json:"document_id"`
			} `json:"payload"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	results := make([]domain.SearchResult, 0, len(raw.Result))

	for _, r := range raw.Result {
		results = append(results, domain.SearchResult{
			Score: r.Score,
			Chunk: domain.Chunk{
				ID:         r.ID,
				DocumentID: r.Payload.DocumentID,
				Text:       r.Payload.Text,
			},
		})
	}

	return results, nil
}
