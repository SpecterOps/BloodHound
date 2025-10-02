//go:build !noop && !replicated

package providers

import (
	"context"
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

type YAMLProvider struct {
	flags    map[string]interface{}
	mu       sync.RWMutex
	filePath string
}

func NewProvider(filePath string) (*YAMLProvider, error) {
	if filePath == "" {
		filePath = "local-harnesses/dogtags.yaml"
	}

	p := &YAMLProvider{
		flags:    make(map[string]interface{}),
		filePath: filePath,
	}

	if err := p.loadFlags(); err != nil {
		return nil, fmt.Errorf("failed to load dogtags: %w", err)
	}

	return p, nil
}

func (p *YAMLProvider) GetBoolFlag(ctx context.Context, key string, defaultValue bool) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if value, exists := p.flags[key]; exists {
		if boolVal, ok := value.(bool); ok {
			return boolVal
		}
	}
	return defaultValue
}

func (p *YAMLProvider) GetStringFlag(ctx context.Context, key string, defaultValue string) string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if value, exists := p.flags[key]; exists {
		if strVal, ok := value.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

func (p *YAMLProvider) GetIntFlag(ctx context.Context, key string, defaultValue int64) int64 {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if value, exists := p.flags[key]; exists {
		switch val := value.(type) {
		case int:
			return int64(val)
		case int64:
			return val
		}
	}
	return defaultValue
}

func (p *YAMLProvider) GetFloatFlag(ctx context.Context, key string, defaultValue float64) float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if value, exists := p.flags[key]; exists {
		if floatVal, ok := value.(float64); ok {
			return floatVal
		}
	}
	return defaultValue
}


func (p *YAMLProvider) loadFlags() error {
	data, err := os.ReadFile(p.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Empty flags if file doesn't exist
		}
		return fmt.Errorf("failed to read flags file %s: %w", p.filePath, err)
	}

	var flags map[string]interface{}
	if err := yaml.Unmarshal(data, &flags); err != nil {
		return fmt.Errorf("failed to parse YAML file %s: %w", p.filePath, err)
	}

	p.mu.Lock()
	p.flags = flags
	p.mu.Unlock()

	return nil
}
