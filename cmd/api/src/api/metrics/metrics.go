package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

func RegisterApiMetrics(registry *prometheus.Registry) error {
	if err := registry.Register(CypherQueryTimeouts); err != nil {
		return err
	}
	return nil
}
