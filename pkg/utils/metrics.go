package utils

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

var eventCounter *prometheus.CounterVec
var durationHistogram *prometheus.HistogramVec

func RecordEvent(name string, err error) {
	labels := prometheus.Labels{"name": name, "isError": fmt.Sprintf("%t", err != nil)}
	eventCounter.With(labels).Inc()
}

func RecordEventDuration(name string, code int, start time.Time) {
	duration := time.Since(start)
	labels := prometheus.Labels{"name": name, "code": fmt.Sprintf("%d", code)}
	durationHistogram.With(labels).Observe(float64(duration / time.Microsecond))
}

func init() {
	durationHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "telemetry",
		Subsystem: "hacking",
		Name:      "duration_histogram_microseconds",
		Help:      "record duration of API endpoints in microseconds",
		Buckets:   prometheus.ExponentialBuckets(1, 3, 20),
	}, []string{"name", "code"})
	prometheus.MustRegister(durationHistogram)

	eventCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "telemetry",
		Subsystem: "hacking",
		Name:      "event_counter",
		Help:      "record events happening inside",
	}, []string{"name", "isError"})
	prometheus.MustRegister(eventCounter)
}
