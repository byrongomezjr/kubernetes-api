package api

import (
	"net/http"

	"kubernetes-api/internal/auth"
	"kubernetes-api/internal/metrics"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

// SetupRouter sets up the HTTP router with all endpoints
func SetupRouter() http.Handler {
	r := mux.NewRouter()

	// Add middleware
	r.Use(metrics.MetricsMiddleware)
	r.Use(loggingMiddleware)

	// Public endpoints
	r.HandleFunc("/api/health", healthHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/auth/register", registerHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/auth/login", loginHandler).Methods(http.MethodPost)

	// Metrics endpoint for Prometheus
	r.Handle("/metrics", promhttp.Handler())

	// Authenticated endpoints
	apiV1 := r.PathPrefix("/api/v1").Subrouter()
	apiV1.Use(auth.AuthMiddleware)

	// Items endpoints
	apiV1.HandleFunc("/items", itemsHandler).Methods(http.MethodGet, http.MethodPost)
	apiV1.HandleFunc("/items/{id:[0-9]+}", itemHandler).Methods(http.MethodGet, http.MethodPut, http.MethodDelete)

	// Handle 404
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	return r
}

// loggingMiddleware logs all HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.WithFields(logrus.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"agent":  r.UserAgent(),
		}).Info("HTTP request")
		next.ServeHTTP(w, r)
	})
}

// notFoundHandler handles 404 errors
func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	if _, err := w.Write([]byte(`{"status":"error","error":"Resource not found"}`)); err != nil {
		logrus.WithError(err).Error("Failed to write not found response")
	}
}
