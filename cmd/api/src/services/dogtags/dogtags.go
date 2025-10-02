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
	BypassCypherQueryLimits FlagKey = "bypass_cypher_query_limits"
	CypherMutability        FlagKey = "cypher_mutability"
	ZoneAllocation          FlagKey = "zone_allocation"
	LabelAllocation         FlagKey = "label_allocation"
)

// FlagSpec defines the specification for a dogtag
type FlagSpec struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// ValidFlags defines all supported flags with their types
var ValidFlags = map[FlagKey]FlagSpec{
	BypassCypherQueryLimits: {Type: "bool", Description: "Bypass cypher query limits"},
	CypherMutability:        {Type: "bool", Description: "Enable cypher mutability"},
	ZoneAllocation:          {Type: "int", Description: "Maximum zone allocation"},
	LabelAllocation:         {Type: "int", Description: "Maximum label allocation"},
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
