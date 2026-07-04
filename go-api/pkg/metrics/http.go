package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var requestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "collector_http_requests_total",
	Help: "Total HTTP requests handled by the API.",
}, []string{"method", "route", "status"})

var requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "collector_http_request_duration_seconds",
	Help:    "HTTP request duration in seconds.",
	Buckets: prometheus.DefBuckets,
}, []string{"method", "route", "status"})

var requestsInFlight = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "collector_http_requests_in_flight",
	Help: "HTTP requests currently being handled.",
})

func init() {
	prometheus.MustRegister(requestsTotal, requestDuration, requestsInFlight)
}

// Middleware records the RED signals: request rate, errors and duration.
// FullPath returns the Gin route template (for example /products/:id), which
// prevents unbounded labels caused by product IDs or other user input.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		requestsInFlight.Inc()
		defer requestsInFlight.Dec()

		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		status := strconv.Itoa(c.Writer.Status())

		requestsTotal.WithLabelValues(c.Request.Method, route, status).Inc()
		requestDuration.WithLabelValues(c.Request.Method, route, status).Observe(time.Since(startedAt).Seconds())
	}
}

func Handler() http.Handler {
	return promhttp.Handler()
}
