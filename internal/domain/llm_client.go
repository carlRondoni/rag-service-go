package domain

import "context"

type LLMClient interface {
	Health(ctx context.Context) error
}
