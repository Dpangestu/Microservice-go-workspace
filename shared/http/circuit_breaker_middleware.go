package http

import (
	"log"
	"net/http"

	"bkc_microservice/shared/circuitbreaker"
)

// CircuitBreakerMiddleware wrap reverse proxy dengan circuit breaker
type CircuitBreakerMiddleware struct {
	handler http.Handler
	cb      *circuitbreaker.CircuitBreaker
}

// NewCircuitBreakerMiddleware membuat middleware baru
func NewCircuitBreakerMiddleware(handler http.Handler, cfg circuitbreaker.Config) *CircuitBreakerMiddleware {
	return &CircuitBreakerMiddleware{
		handler: handler,
		cb:      circuitbreaker.NewCircuitBreaker(cfg),
	}
}

// ServeHTTP handle request dengan circuit breaker
func (m *CircuitBreakerMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check apakah circuit open
	if m.cb.IsOpen() {
		log.Printf("🔴 Circuit breaker OPEN - rejecting request: %s %s", r.Method, r.RequestURI)
		http.Error(w, "service unavailable - circuit breaker open", http.StatusServiceUnavailable)
		return
	}

	// Custom response writer untuk track status
	rw := &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}

	// Forward request
	m.handler.ServeHTTP(rw, r)

	// Record success atau failure based on status code
	if rw.statusCode >= 500 {
		// Server error
		m.cb.RecordFailure()
		log.Printf("⚠️ Request failed with status %d - recording failure", rw.statusCode)
	} else if rw.statusCode == http.StatusServiceUnavailable {
		// Service unavailable
		m.cb.RecordFailure()
		log.Printf("⚠️ Service unavailable - recording failure")
	} else {
		// Success (2xx, 3xx, 4xx)
		m.cb.RecordSuccess()
		if m.cb.GetState() == circuitbreaker.StateClosed {
			log.Printf("✅ Request success - %s", m.cb.String())
		}
	}
}

// responseWriter custom writer untuk capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}
