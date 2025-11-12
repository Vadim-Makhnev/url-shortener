package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	URLShortenRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "url_shorten_request_total",
		Help: "Total number of URL shorten requests",
	})

	URLRedirectRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "url_redirect_count_total",
		Help: "Total number of URL redirect requests",
	})

	URLAccessCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "url_access_count_total",
		Help: "Total number of URL accesses",
	}, []string{"short_code"})

	RequestDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duration of HTTP requests",
		Buckets: prometheus.DefBuckets,
	})
)

func InitMetrics() {}
