package graphify

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	ingestThroughputGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "bhe_ingest_throughput",
			Help: "Ingestion throughput in entities per second",
		},
		[]string{"entity_type", "stage"}, // "nodes" or "relationships", "processed" or "written"
	)
)

// RegisterIngestMetrics registers the ingestion throughput gauge with the provided Prometheus registerer.
func RegisterIngestMetrics(registerer prometheus.Registerer) error {
	if err := registerer.Register(ingestThroughputGauge); err != nil {
		return fmt.Errorf("failed to register ingest metrics: %w", err)
	}
	return nil
}
