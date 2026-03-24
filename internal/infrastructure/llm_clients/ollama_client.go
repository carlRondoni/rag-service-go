package llm_clients

import (
	"context"
	"fmt"
	"net/http"
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
		defer resp.Body.Close()

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
