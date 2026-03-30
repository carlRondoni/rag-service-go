package domain

import "context"

type LLMClient interface {
	Generate(ctx context.Context, prompt string) (string, error)
	Health(ctx context.Context) error
}
