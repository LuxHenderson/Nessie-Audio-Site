package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

var (
	// ErrCircuitOpen is returned when the circuit breaker is open
	ErrCircuitOpen = errors.New("circuit breaker is open")

	// ErrTooManyRequests is returned when too many requests are made in half-open state
	ErrTooManyRequests = errors.New("too many requests")
)

// State represents the circuit breaker state
type State int

const (
	// StateClosed allows all requests through
	StateClosed State = iota

	// StateOpen rejects all requests
	StateOpen

	// StateHalfOpen allows limited requests through to test recovery
	StateHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name            string
	maxFailures     uint
	resetTimeout    time.Duration
	halfOpenMaxReqs uint

	mu              sync.RWMutex
	state           State
	failures        uint
	lastFailureTime time.Time
	halfOpenReqs    uint
}

// Config holds circuit breaker configuration
type Config struct {
	Name            string        // Name for logging/debugging
	MaxFailures     uint          // Failures before opening circuit
	ResetTimeout    time.Duration // Time before trying half-open
	HalfOpenMaxReqs uint          // Max requests in half-open state
}

// New creates a new circuit breaker
func New(cfg Config) *CircuitBreaker {
	if cfg.MaxFailures == 0 {
		cfg.MaxFailures = 5
	}
	if cfg.ResetTimeout == 0 {
		cfg.ResetTimeout = 60 * time.Second
	}
	if cfg.HalfOpenMaxReqs == 0 {
		cfg.HalfOpenMaxReqs = 1
	}

	return &CircuitBreaker{
		name:            cfg.Name,
		maxFailures:     cfg.MaxFailures,
		resetTimeout:    cfg.ResetTimeout,
		halfOpenMaxReqs: cfg.HalfOpenMaxReqs,
		state:           StateClosed,
	}
}

// Execute runs the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() error) error {
	// Check if we can execute
	if err := cb.beforeRequest(); err != nil {
		return err
	}

	// Execute the function
	err := fn()

	// Update circuit breaker based on result
	cb.afterRequest(err)

	return err
}

// beforeRequest checks if the request should be allowed
func (cb *CircuitBreaker) beforeRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return nil

	case StateOpen:
		// Check if we should transition to half-open
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			cb.state = StateHalfOpen
			cb.halfOpenReqs = 0
			return nil
		}
		return ErrCircuitOpen

	case StateHalfOpen:
		// Allow limited requests in half-open state
		if cb.halfOpenReqs >= cb.halfOpenMaxReqs {
			return ErrTooManyRequests
		}
		cb.halfOpenReqs++
		return nil

	default:
		return nil
	}
}

// afterRequest updates the circuit breaker based on request result
func (cb *CircuitBreaker) afterRequest(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.onFailure()
	} else {
		cb.onSuccess()
	}
}

// onFailure handles a failed request
func (cb *CircuitBreaker) onFailure() {
	cb.failures++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.maxFailures {
			cb.state = StateOpen
		}

	case StateHalfOpen:
		// Any failure in half-open goes back to open
		cb.state = StateOpen
	}
}

// onSuccess handles a successful request
func (cb *CircuitBreaker) onSuccess() {
	switch cb.state {
	case StateClosed:
		cb.failures = 0

	case StateHalfOpen:
		// Success in half-open closes the circuit
		cb.state = StateClosed
		cb.failures = 0
		cb.halfOpenReqs = 0
	}
}

// GetState returns the current state (for monitoring/debugging)
func (cb *CircuitBreaker) GetState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetName returns the circuit breaker name
func (cb *CircuitBreaker) GetName() string {
	return cb.name
}

// GetStats returns current statistics
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	stateStr := "closed"
	switch cb.state {
	case StateOpen:
		stateStr = "open"
	case StateHalfOpen:
		stateStr = "half-open"
	}

	return map[string]interface{}{
		"name":     cb.name,
		"state":    stateStr,
		"failures": cb.failures,
	}
}
