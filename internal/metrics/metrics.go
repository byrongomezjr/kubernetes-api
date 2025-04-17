package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// RequestsTotal is a counter for total HTTP requests
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests by method and path",
		},
		[]string{"method", "path", "status"},
	)

	// RequestDuration is a histogram for HTTP request durations
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// ResponseSize is a histogram for HTTP response sizes
	ResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path"},
	)

	// DatabaseOperationsTotal is a counter for database operations
	DatabaseOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_operations_total",
			Help: "Total number of database operations",
		},
		[]string{"operation", "status"},
	)

	// DatabaseOperationDuration is a histogram for database operation durations
	DatabaseOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_operation_duration_seconds",
			Help:    "Database operation duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)
)

// PrometheusHandler returns the HTTP handler for the Prometheus metrics endpoint
func PrometheusHandler() http.Handler {
	return promhttp.Handler()
}

// MetricsMiddleware is a middleware that records HTTP request metrics
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code and size
		metricsWriter := newMetricsResponseWriter(w)

		// Call next handler
		next.ServeHTTP(metricsWriter, r)

		// Record metrics
		duration := time.Since(start).Seconds()
		RequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
		RequestsTotal.WithLabelValues(r.Method, r.URL.Path, http.StatusText(metricsWriter.statusCode)).Inc()
		ResponseSize.WithLabelValues(r.Method, r.URL.Path).Observe(float64(metricsWriter.size))
	})
}

// metricsResponseWriter is a wrapper for http.ResponseWriter that captures status code and size
type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

// newMetricsResponseWriter creates a new metricsResponseWriter
func newMetricsResponseWriter(w http.ResponseWriter) *metricsResponseWriter {
	return &metricsResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // Default status code
	}
}

// WriteHeader implements http.ResponseWriter
func (w *metricsResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Write implements http.ResponseWriter
func (w *metricsResponseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.size += size
	return size, err
}

// TrackDatabaseOperation tracks a database operation duration
func TrackDatabaseOperation(operation string, f func() error) error {
	start := time.Now()
	err := f()
	duration := time.Since(start).Seconds()

	// Record metrics
	DatabaseOperationDuration.WithLabelValues(operation).Observe(duration)
	if err != nil {
		DatabaseOperationsTotal.WithLabelValues(operation, "error").Inc()
	} else {
		DatabaseOperationsTotal.WithLabelValues(operation, "success").Inc()
	}

	return err
}
