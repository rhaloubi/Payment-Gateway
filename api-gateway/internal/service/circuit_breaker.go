package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/rhaloubi/api-gateway/internal/config"
)

type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

type CircuitBreaker struct {
	mu       sync.RWMutex
	circuits map[string]*Circuit
	config   *config.Config
}

type Circuit struct {
	state           CircuitState
	failures        int
	successes       int
	lastFailureTime time.Time
	lastStateChange time.Time
	config          config.ServiceCircuitBreakerConfig
}

func NewCircuitBreaker(cfg *config.Config) *CircuitBreaker {
	cb := &CircuitBreaker{
		circuits: make(map[string]*Circuit),
		config:   cfg,
	}

	cb.circuits["auth"] = &Circuit{
		state:  StateClosed,
		config: cfg.CircuitBreaker.AuthService,
	}
	cb.circuits["merchant"] = &Circuit{
		state:  StateClosed,
		config: cfg.CircuitBreaker.MerchantService,
	}
	cb.circuits["payment"] = &Circuit{
		state:  StateClosed,
		config: cfg.CircuitBreaker.PaymentService,
	}

	return cb
}

func (cb *CircuitBreaker) Allow(service string) error {
	cb.mu.RLock()
	circuit, exists := cb.circuits[service]
	cb.mu.RUnlock()

	if !exists {
		return fmt.Errorf("circuit not found for service: %s", service)
	}

	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch circuit.state {
	case StateClosed:
		return nil
	case StateOpen:
		if time.Since(circuit.lastStateChange) > circuit.config.Timeout {
			circuit.state = StateHalfOpen
			circuit.successes = 0
			return nil
		}
		return fmt.Errorf("circuit breaker open for service: %s", service)
	case StateHalfOpen:
		return nil
	default:
		return fmt.Errorf("unknown circuit state")
	}
}

func (cb *CircuitBreaker) RecordSuccess(service string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	circuit, exists := cb.circuits[service]
	if !exists {
		return
	}

	switch circuit.state {
	case StateClosed:
		circuit.failures = 0
	case StateHalfOpen:
		circuit.successes++
		if circuit.successes >= circuit.config.SuccessThreshold {
			circuit.state = StateClosed
			circuit.failures = 0
			circuit.successes = 0
			circuit.lastStateChange = time.Now()
		}
	}
}

func (cb *CircuitBreaker) RecordFailure(service string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	circuit, exists := cb.circuits[service]
	if !exists {
		return
	}

	circuit.lastFailureTime = time.Now()
	circuit.failures++

	switch circuit.state {
	case StateClosed:
		if circuit.failures >= circuit.config.FailureThreshold {
			circuit.state = StateOpen
			circuit.lastStateChange = time.Now()
		}
	case StateHalfOpen:
		circuit.state = StateOpen
		circuit.successes = 0
		circuit.lastStateChange = time.Now()
	}
}

func (cb *CircuitBreaker) GetState(service string) CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if circuit, exists := cb.circuits[service]; exists {
		return circuit.state
	}
	return StateClosed
}
