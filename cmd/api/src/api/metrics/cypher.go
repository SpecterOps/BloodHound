package metrics

import (
	"errors"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/queries"
)

var (
	CypherQueryTimeouts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cypher_query_errors",
			Help: "A counter for cypher query errors",
		},
		[]string{"error_type"},
	)
)

func RecordCypherQueryTimeout(err error) {
	var errorType string
	switch {
	case strings.Contains(err.Error(), "timeout"):
		errorType = "timeout error"
	case errors.Is(err, api.ErrContentTypeJson), errors.Is(err, api.ErrNoRequestBody), strings.Contains(err.Error(), "could not decode limited payload request into value"):
		errorType = "json decode error"
	case errors.Is(err, queries.ErrCypherQueryTooComplex):
		errorType = "query too complex error"
	case strings.Contains(err.Error(), "query required more memory than allowed"):
		errorType = "query exceeded maximum allowed memory limit"
	case strings.Contains(err.Error(), "failed to parse cypher"):
		errorType = "parse cypher error"
	case strings.Contains(err.Error(), "SQL"):
		errorType = "cypher translation error"
	default:
		errorType = "unknown error"
	}
	CypherQueryTimeouts.WithLabelValues(errorType).Inc()
}
