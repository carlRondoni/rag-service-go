package vector_store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"rag-service-go/internal/domain"
	"time"

	"github.com/rs/zerolog"
)

type QdrantStore struct {
	baseURL    string
	collection string
	client     *http.Client
	logger     zerolog.Logger
}

func NewQdrantStore(
	baseURL,
	collection string,
	logger zerolog.Logger,
) *QdrantStore {
	return &QdrantStore{
		baseURL:    baseURL,
		collection: collection,
		client: &http.Client{
			Timeout: 2 * time.Minute,
		},
		logger: logger,
	}
}

func (q *QdrantStore) EnsureCollection(ctx context.Context, vectorSize int) error {
	checkURL := fmt.Sprintf("%s/collections/%s", q.baseURL, q.collection)
	reqCheck, err := http.NewRequestWithContext(ctx, http.MethodGet, checkURL, nil)
	if err != nil {
		return err
	}

	respCheck, err := q.client.Do(reqCheck)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := respCheck.Body.Close(); closeErr != nil {
			q.logger.Error().Err(closeErr).Msg("error closing response body")
		}
	}()

	if respCheck.StatusCode == http.StatusOK {
		return nil
	}

	if respCheck.StatusCode == http.StatusNotFound {
		if vectorSize <= 0 {
			vectorSize = 4096
		}

		createBody := map[string]interface{}{
			"vectors": map[string]interface{}{
				"size":     vectorSize,
				"distance": "Cosine",
			},
		}

		b, err := json.Marshal(createBody)
		if err != nil {
			return err
		}

		reqCreate, err := http.NewRequestWithContext(ctx, http.MethodPut, checkURL, bytes.NewBuffer(b))
		if err != nil {
			return err
		}
		reqCreate.Header.Set("Content-Type", "application/json")

		respCreate, err := q.client.Do(reqCreate)
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := respCreate.Body.Close(); closeErr != nil {
				q.logger.Error().Err(closeErr).Msg("error closing response body")
			}
		}()

		if respCreate.StatusCode >= 300 {
			return fmt.Errorf("failed to create qdrant collection %s, status: %d", q.collection, respCreate.StatusCode)
		}

		q.logger.Info().Str("collection", q.collection).Int("vector_size", vectorSize).Msg("qdrant collection created successfully")
		return nil
	}

	return fmt.Errorf("unexpected status checking qdrant collection: %d", respCheck.StatusCode)
}

func (q *QdrantStore) UpsertChunks(
	ctx context.Context,
	chunks []domain.Chunk,
	embeddings []domain.Embedding,
) error {
	if len(chunks) == 0 {
		return nil
	}

	if len(chunks) != len(embeddings) {
		return fmt.Errorf("chunks length (%d) does not match embeddings length (%d)", len(chunks), len(embeddings))
	}

	if len(embeddings) > 0 && len(embeddings[0].Vector) > 0 {
		if err := q.EnsureCollection(ctx, len(embeddings[0].Vector)); err != nil {
			return fmt.Errorf("failed to ensure collection exists: %w", err)
		}
	}

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

	b, err := json.Marshal(body)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/collections/%s/points", q.baseURL, q.collection)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := q.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			q.logger.Error().Err(err).Msg("error closing response body")
		}
	}()

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
		"vector":       query.Vector,
		"limit":        topK,
		"with_payload": true,
	}

	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/collections/%s/points/search", q.baseURL, q.collection)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := q.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			q.logger.Error().Err(err).Msg("error closing response body")
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		return []domain.SearchResult{}, nil
	}

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("qdrant search error, status: %d", resp.StatusCode)
	}

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

func (q *QdrantStore) Health(ctx context.Context) error {
	url := fmt.Sprintf("%s/collections", q.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := q.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			q.logger.Error().Err(err).Msg("error closing response body")
		}
	}()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("qdrant error: %d", resp.StatusCode)
	}

	return nil
}
