package post

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	postOperationsVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "analysis",
		Name:      "post_processing_ops",
		Help:      "Post-processing operation statistics.",
	}, []string{
		"kind",
		"operation",
	})
)

// RegisterPostProcessingMetrics registers the analysis post-processing metrics.
func RegisterPostProcessingMetrics(registerer prometheus.Registerer) error {
	if err := registerer.Register(postOperationsVec); err != nil {
		return fmt.Errorf("failed to register analysis post-processing metrics: %w", err)
	}

	return nil
}
