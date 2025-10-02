package changelog

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	actionHit   action = "hit"
	actionMis   action = "miss"
	actionError action = "error"
)

type action string

func (a action) String() string {
	return string(a)
}

// TODO: Add Cache Size to Metrics
var (
	changeLogCacheResult = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "change_log_cache",
		Help: "A counter for ",
	}, []string{"action"}) // hit or miss
)

func RecordChangeLogCacheResult(action action) {
	changeLogCacheResult.With(prometheus.Labels{"action": action.String()}).Inc()
}

func RegisterChangeLogMetrics(reg prometheus.Registerer) error {
	return reg.Register(changeLogCacheResult)
}
