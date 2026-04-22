package metrics_test

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/specterops/bloodhound/packages/go/metrics"
)

func TestNewRegistry(t *testing.T) {
	t.Parallel()

	registry, err := metrics.NewRegistry()

	require.NoError(t, err)
	require.NotNil(t, registry)

	// Verify the returned registry is functional and accepts metric registration
	testCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_new_registry_counter_total",
		Help: "A counter used to verify the registry accepts metric registration.",
	})
	assert.NoError(t, registry.Register(testCounter))
}

func TestInitializeBHCEMetrics(t *testing.T) {
	t.Parallel()

	registry := prometheus.NewRegistry()
	err := metrics.InitializeBHCEMetrics(registry)

	assert.NoError(t, err)
}
