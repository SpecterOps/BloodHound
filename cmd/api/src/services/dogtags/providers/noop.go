//go:build noop

package providers

import (
	"context"
)

type NoOpProvider struct{}

func NewProvider(filePath string) (*NoOpProvider, error) {
	return &NoOpProvider{}, nil
}

func (p *NoOpProvider) GetBoolFlag(ctx context.Context, key string, defaultValue bool) bool {
	return defaultValue
}

func (p *NoOpProvider) GetStringFlag(ctx context.Context, key string, defaultValue string) string {
	return defaultValue
}

func (p *NoOpProvider) GetIntFlag(ctx context.Context, key string, defaultValue int64) int64 {
	return defaultValue
}

func (p *NoOpProvider) GetFloatFlag(ctx context.Context, key string, defaultValue float64) float64 {
	return defaultValue
}