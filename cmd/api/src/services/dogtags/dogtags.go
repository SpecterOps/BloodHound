package dogtags

import (
	"context"
	"fmt"

	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags/providers"
)

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

// Service defines the interface for the dogtags service
type Service interface {
	GetAllFlags(ctx context.Context) map[FlagKey]interface{}
}

// Provider is what provider implementations must implement (with string keys)
// External repos can implement this interface to create custom providers
type Provider interface {
	GetBoolFlag(ctx context.Context, key string, defaultValue bool) bool
	GetStringFlag(ctx context.Context, key string, defaultValue string) string
	GetIntFlag(ctx context.Context, key string, defaultValue int64) int64
	GetFloatFlag(ctx context.Context, key string, defaultValue float64) float64
}

// service wraps a provider with the for-loop logic
type service struct {
	provider Provider
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

// NewService creates a new dogtags service with the given provider
func NewService(provider Provider) Service {
	return &service{provider: provider}
}

// NewYAMLProvider creates a YAML-based provider (convenience export)
func NewYAMLProvider(filePath string) (Provider, error) {
	return providers.NewYAMLProvider(filePath)
}
