package circuitbreaker

import (
	"fmt"
	"sync"
	"time"
)

// State untuk circuit breaker
type State string

const (
	StateClosed   State = "CLOSED"    // Normal - requests berjalan
	StateOpen     State = "OPEN"      // Error banyak - reject requests
	StateHalfOpen State = "HALF_OPEN" // Testing - coba request lagi
)

// CircuitBreaker implementasi pattern circuit breaker
type CircuitBreaker struct {
	mu                  sync.RWMutex
	state               State
	failures            int
	successes           int
	lastFailureTime     time.Time
	failureThreshold    int
	successThreshold    int
	timeout             time.Duration
	halfOpenMaxRequests int
	halfOpenRequests    int
}

// Config untuk circuit breaker
type Config struct {
	FailureThreshold    int           // jumlah failure sebelum open
	SuccessThreshold    int           // jumlah success sebelum close kembali
	Timeout             time.Duration // berapa lama state open sebelum half-open
	HalfOpenMaxRequests int           // max requests di half-open state
}

// NewCircuitBreaker membuat instance baru
func NewCircuitBreaker(cfg Config) *CircuitBreaker {
	return &CircuitBreaker{
		state:               StateClosed,
		failures:            0,
		successes:           0,
		failureThreshold:    cfg.FailureThreshold,
		successThreshold:    cfg.SuccessThreshold,
		timeout:             cfg.Timeout,
		halfOpenMaxRequests: cfg.HalfOpenMaxRequests,
	}
}

// GetState mengembalikan state saat ini
func (cb *CircuitBreaker) GetState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// IsOpen mengecek apakah circuit sedang open (reject requests)
func (cb *CircuitBreaker) IsOpen() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Jika state OPEN, cek apakah timeout sudah lewat
	if cb.state == StateOpen {
		if time.Since(cb.lastFailureTime) > cb.timeout {
			// Timeout lewat → transition ke HALF_OPEN untuk test
			cb.state = StateHalfOpen
			cb.halfOpenRequests = 0
			cb.successes = 0
			return false // Izinkan request masuk untuk test
		}
		return true // Circuit masih open
	}

	// Jika state HALF_OPEN
	if cb.state == StateHalfOpen {
		if cb.halfOpenRequests >= cb.halfOpenMaxRequests {
			// Sudah cukup requests di half-open
			if cb.successes >= cb.successThreshold {
				// Cukup success → CLOSE kembali
				cb.state = StateClosed
				cb.failures = 0
				cb.successes = 0
				return false
			}
			// Masih banyak failure → OPEN lagi
			cb.state = StateOpen
			cb.lastFailureTime = time.Now()
			return true
		}
		return false // Izinkan request untuk test
	}

	// State CLOSED - normal operation
	return false
}

// RecordSuccess mencatat sukses response
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateClosed {
		// Reset counter di state closed
		cb.failures = 0
		return
	}

	if cb.state == StateHalfOpen {
		cb.successes++
		cb.halfOpenRequests++

		// Jika cukup success → close
		if cb.successes >= cb.successThreshold {
			cb.state = StateClosed
			cb.failures = 0
			cb.successes = 0
		}
	}
}

// RecordFailure mencatat gagal response
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateClosed {
		cb.failures++
		if cb.failures >= cb.failureThreshold {
			// Cukup failure → open
			cb.state = StateOpen
			cb.lastFailureTime = time.Now()
		}
		return
	}

	if cb.state == StateHalfOpen {
		// Failure di half-open → open lagi
		cb.state = StateOpen
		cb.lastFailureTime = time.Now()
	}
}

// Reset mereset circuit breaker ke state awal
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failures = 0
	cb.successes = 0
	cb.halfOpenRequests = 0
}

// String mengembalikan info circuit breaker
func (cb *CircuitBreaker) String() string {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return fmt.Sprintf(
		"CircuitBreaker{state=%s, failures=%d, successes=%d, halfOpenRequests=%d}",
		cb.state, cb.failures, cb.successes, cb.halfOpenRequests,
	)
}
