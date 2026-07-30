package llm_clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"rag-service-go/internal/domain"
	"time"

	"github.com/rs/zerolog"
	"github.com/sony/gobreaker/v2"
	"go.opentelemetry.io/otel"
	"golang.org/x/time/rate"
)

type OllamaClient struct {
	baseURL string
	model   string
	http    *http.Client
	breaker *gobreaker.CircuitBreaker[any]
	limiter *rate.Limiter
	logger  zerolog.Logger
}

func NewOllamaClient(baseURL string, model string, logger zerolog.Logger) OllamaClient {
	cbst := gobreaker.Settings{
		Name:        "LLM",
		MaxRequests: 1,
		Interval:    60 * time.Second,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 3
		},
	}

	cb := gobreaker.NewCircuitBreaker[any](cbst)

	return OllamaClient{
		baseURL: baseURL,
		model:   model,
		http: &http.Client{
			Timeout: 2 * time.Minute, // needed for first time doing requests to ollama
		},
		breaker: cb,
		limiter: rate.NewLimiter(2, 5),
		logger:  logger,
	}
}

func (c OllamaClient) Generate(ctx context.Context, prompt string) (string, error) {
	c.logger.Info().Str("prompt", prompt).Msg("generate")

	if err := c.limiter.Wait(ctx); err != nil {
		c.logger.Error().Err(err).Msg("rate limit exceeded")
		return "", err
	}

	tr := otel.Tracer("llm-api")
	ctx, span := tr.Start(ctx, "Generate")
	defer span.End()

	result, err := c.breaker.Execute(func() (any, error) {
		reqBody := map[string]string{
			"model":  c.model,
			"prompt": prompt,
		}

		b, _ := json.Marshal(reqBody)

		req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewBuffer(b))
		if err != nil {
			c.logger.Error().
				Err(err).
				Str("prompt", prompt).
				Msg("failed to request LLM")
			return "", err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.http.Do(req)
		if err != nil {
			c.logger.Error().
				Err(err).
				Str("prompt", prompt).
				Msg("failed to call LLM")
			return "", err
		}
		defer func() {
			err := resp.Body.Close()
			if err != nil {
				c.logger.Error().Err(err).Msg("error closing response body")
			}
		}()

		if resp.StatusCode >= 400 {
			c.logger.Error().
				Err(err).
				Str("prompt", prompt).
				Int("status", resp.StatusCode).
				Msg("failed on call LLM")
			return "", fmt.Errorf("llm generate failed, status: %d", resp.StatusCode)
		}

		var out struct {
			Response string `json:"response"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			c.logger.Error().
				Err(err).
				Str("prompt", prompt).
				Msg("failed to call LLM")
			return nil, err
		}

		return out.Response, nil
	})

	if err != nil {
		c.logger.Error().
			Err(err).
			Str("prompt", prompt).
			Msg("failed to call LLM")
		return "", err
	}

	response, ok := result.(string)
	if !ok {
		c.logger.Error().
			Str("type", fmt.Sprintf("%T", result)).
			Msg("wrong LLM response")
		return "", fmt.Errorf("unexpected response type: %T", result)
	}

	return response, nil
}

func (c OllamaClient) Health(ctx context.Context) error {
	c.logger.Info().Msg("health")

	if err := c.limiter.Wait(ctx); err != nil {
		c.logger.Error().Err(err).Msg("rate limit exceeded")
		return err
	}

	tr := otel.Tracer("llm-api")
	ctx, span := tr.Start(ctx, "Health")
	defer span.End()

	_, err := c.breaker.Execute(func() (any, error) {
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			c.baseURL+"/api/tags",
			nil,
		)
		if err != nil {
			c.logger.Error().
				Err(err).
				Msg("failed on request LLM")
			return nil, err
		}

		resp, err := c.http.Do(req)
		if err != nil {
			c.logger.Error().
				Err(err).
				Msg("failed to call LLM")
			return nil, err
		}
		defer func() {
			err := resp.Body.Close()
			if err != nil {
				c.logger.Error().Err(err).Msg("error closing response body")
			}
		}()

		if resp.StatusCode >= 500 {
			c.logger.Error().
				Int("status", resp.StatusCode).
				Msg("llm unhealthy status")
			return nil, fmt.Errorf("llm unhealthy status: %d", resp.StatusCode)
		}

		return nil, nil
	})

	return err
}

func (o OllamaClient) Embed(
	ctx context.Context,
	texts []string,
) ([]domain.Embedding, error) {

	type inputRequest struct {
		Model string `json:"model"`
		Input string `json:"input"`
	}

	type embeddingResponse struct {
		Embedding []float32 `json:"embedding"`
	}

	var embeddingResult []domain.Embedding

	for _, text := range texts {
		reqBody := inputRequest{
			Model: o.model,
			Input: text,
		}

		b, _ := json.Marshal(reqBody)

		url := fmt.Sprintf("%s/api/embeddings", o.baseURL)

		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(b))
		req.Header.Set("Content-Type", "application/json")

		resp, err := o.http.Do(req)
		if err != nil {
			return nil, err
		}
		defer func() {
			err := resp.Body.Close()
			if err != nil {
				o.logger.Error().Err(err).Msg("error closing response body")
			}
		}()

		var r embeddingResponse
		if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
			return nil, err
		}

		embeddingResult = append(embeddingResult, domain.Embedding{
			Vector: r.Embedding,
		})
	}

	return embeddingResult, nil
}
