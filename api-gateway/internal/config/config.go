package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server         ServerConfig         `yaml:"server"`
	Services       ServicesConfig       `yaml:"services"`
	RateLimiting   RateLimitingConfig   `yaml:"rate_limiting"`
	CircuitBreaker CircuitBreakerConfig `yaml:"circuit_breaker"`
	Authentication AuthenticationConfig `yaml:"authentication"`
	Logging        LoggingConfig        `yaml:"logging"`
	Metrics        MetricsConfig        `yaml:"metrics"`
}

type ServerConfig struct {
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
}

type ServicesConfig struct {
	Auth     ServiceConfig `yaml:"auth"`
	Merchant ServiceConfig `yaml:"merchant"`
	Payment  ServiceConfig `yaml:"payment"`
}

type ServiceConfig struct {
	URL     string        `yaml:"url"`
	Timeout time.Duration `yaml:"timeout"`
}

type RateLimitingConfig struct {
	Enabled   bool                  `yaml:"enabled"`
	Storage   string                `yaml:"storage"`
	RedisURL  string                `yaml:"redis_url"`
	Global    GlobalRateLimitConfig `yaml:"global"`
	Endpoints []EndpointRateLimit   `yaml:"endpoints"`
}

type GlobalRateLimitConfig struct {
	RequestsPerHour int `yaml:"requests_per_hour"`
}

type EndpointRateLimit struct {
	Pattern           string `yaml:"pattern"`
	RequestsPerMinute int    `yaml:"requests_per_minute,omitempty"`
	RequestsPerHour   int    `yaml:"requests_per_hour,omitempty"`
	RequestsPerSecond int    `yaml:"requests_per_second,omitempty"`
	By                string `yaml:"by"`
}

type CircuitBreakerConfig struct {
	Enabled         bool                        `yaml:"enabled"`
	AuthService     ServiceCircuitBreakerConfig `yaml:"auth_service"`
	MerchantService ServiceCircuitBreakerConfig `yaml:"merchant_service"`
	PaymentService  ServiceCircuitBreakerConfig `yaml:"payment_service"`
}

type ServiceCircuitBreakerConfig struct {
	FailureThreshold int           `yaml:"failure_threshold"`
	Timeout          time.Duration `yaml:"timeout"`
	SuccessThreshold int           `yaml:"success_threshold"`
}

type AuthenticationConfig struct {
	JWT    JWTConfig    `yaml:"jwt"`
	APIKey APIKeyConfig `yaml:"api_key"`
}

type JWTConfig struct {
	Enabled bool   `yaml:"enabled"`
	Secret  string `yaml:"secret"`
	Issuer  string `yaml:"issuer"`
}

type APIKeyConfig struct {
	Enabled       bool   `yaml:"enabled"`
	ValidationURL string `yaml:"validation_url"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

type MetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Port    int    `yaml:"port"`
	Path    string `yaml:"path"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Expand environment variables
	expanded := os.ExpandEnv(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}
