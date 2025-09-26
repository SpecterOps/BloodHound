package dogtags

import (
	"context"
	"fmt"

	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags/providers"
)

// Config for dogtags service
type Config struct {
	FilePath string `yaml:"file_path"`
}

// FlagKey represents a valid dogtag key
type FlagKey string

// All valid dogtag keys - finite list controlled by the system
const (
	CanAppStartup  FlagKey = "can_app_startup"
	MaxConnections FlagKey = "max_connections"
	ApiBaseURL     FlagKey = "api_base_url"
)

// FlagSpec defines the specification for a dogtag
type FlagSpec struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// ValidFlags defines all supported flags with their types
var ValidFlags = map[FlagKey]FlagSpec{
	CanAppStartup:  {Type: "bool", Description: "Controls whether the application can start up"},
	MaxConnections: {Type: "int", Description: "Maximum number of concurrent connections"},
	ApiBaseURL:     {Type: "string", Description: "Base URL for API endpoints"},
}

// ValidateFlag ensures only known flags are used
func ValidateFlag(key FlagKey) error {
	if _, exists := ValidFlags[key]; !exists {
		return fmt.Errorf("unknown dogtag: %s", key)
	}
	return nil
}


// Provider defines the interface for dogtags providers
type Provider interface {
	GetAllFlags(ctx context.Context) map[FlagKey]interface{}
}

// rawProvider is what providers actually implement (with string keys)
type rawProvider interface {
	GetBoolFlag(ctx context.Context, key string, defaultValue bool) bool
	GetStringFlag(ctx context.Context, key string, defaultValue string) string
	GetIntFlag(ctx context.Context, key string, defaultValue int64) int64
	GetFloatFlag(ctx context.Context, key string, defaultValue float64) float64
}

// service wraps a raw provider with the for-loop logic
type service struct {
	provider rawProvider
}

func (s *service) GetAllFlags(ctx context.Context) map[FlagKey]interface{} {
	result := make(map[FlagKey]interface{})

	// Loop through all known flags and get their values from the provider
	for flagKey, spec := range ValidFlags {
		switch spec.Type {
		case "bool":
			result[flagKey] = s.provider.GetBoolFlag(ctx, string(flagKey), false)
		case "string":
			result[flagKey] = s.provider.GetStringFlag(ctx, string(flagKey), "")
		case "int":
			result[flagKey] = s.provider.GetIntFlag(ctx, string(flagKey), 0)
		case "float":
			result[flagKey] = s.provider.GetFloatFlag(ctx, string(flagKey), 0.0)
		}
	}

	return result
}

// NewService creates a new dogtags service
func NewService(config Config) (Provider, error) {
	rawProvider, err := providers.NewProvider(config.FilePath)
	if err != nil {
		return nil, err
	}
	return &service{provider: rawProvider}, nil
}
