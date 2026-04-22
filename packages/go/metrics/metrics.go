package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

// NewRegistry creates a custom Prometheus registry and registers it to the
// default registerer.
//
// Returns the custom registry for metric registration, or an error if registration fails.
func NewRegistry() (*prometheus.Registry, error) {
	promRegistry := prometheus.NewRegistry()

	// Register custom registry to default registerer so promhttp.Handler() exposes it
	if err := prometheus.DefaultRegisterer.Register(promRegistry); err != nil {
		return nil, fmt.Errorf("failed to register default metrics collector: %w", err)
	}

	return promRegistry, nil
}

// InitializeBHCEMetrics - Placeholder til API middleware and DB metrics are implemented and added here
func InitializeBHCEMetrics() error {
	// register api metrics
	// register db metrics
	return nil
}
