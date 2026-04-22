package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/specterops/bloodhound/packages/go/analysis/post"
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

// InitializeBHCEMetrics registers all BHCE Prometheus metrics to the provided registerer.
func InitializeBHCEMetrics(registerer prometheus.Registerer) error {
	if err := post.InitializePostProcessingMetrics(registerer); err != nil {
		return fmt.Errorf("failed to register post-processing metrics: %w", err)
	}

	return nil
}
