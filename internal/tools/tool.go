package tools

import (
	"context"

	"github.com/tmc/langchaingo/llms"
)

type Tool interface {
	Name() string
	Description() string
	Execute(ctx context.Context, input string) (string, error)
	Definition() llms.Tool
}
