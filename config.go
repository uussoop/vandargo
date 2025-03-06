// Package vandargo provides a secure integration with the Vandar payment gateway
// config.go contains configuration structures and initialization functions
package vandargo

import (
	"errors"
	"time"
)

// Config holds all configuration options for the Vandar client
type Config struct {
	// APIKey is the authentication key for Vandar API
	APIKey string

	// BaseURL is the base URL for the Vandar API
	BaseURL string

	// SandboxMode determines whether to use the sandbox environment
	SandboxMode bool

	// Timeout is the HTTP client timeout in seconds
	Timeout int

	// CallbackURL is the URL that Vandar will redirect to after payment
	CallbackURL string

	// MaxRetries is the maximum number of retry attempts for failed requests
	MaxRetries int

	// RetryWaitTime is the initial wait time between retries (exponential backoff)
	RetryWaitTime time.Duration

	// EncryptionKey is used for encrypting sensitive data
	EncryptionKey string

	// IPAllowList contains allowed IP addresses for callbacks (optional)
	IPAllowList []string
}

// DefaultConfig returns a Config with safe default values
func DefaultConfig() Config {
	return Config{
		BaseURL:       "https://api.vandar.io",
		SandboxMode:   true,
		Timeout:       30,
		MaxRetries:    3,
		RetryWaitTime: 2 * time.Second,
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.APIKey == "" {
		return errors.New("api key is required")
	}

	if c.BaseURL == "" {
		return errors.New("base url is required")
	}

	if c.CallbackURL == "" {
		return errors.New("callback url is required")
	}

	if c.Timeout <= 0 {
		return errors.New("timeout must be greater than 0")
	}

	return nil
}

// configImpl implements the ConfigInterface
type configImpl struct {
	config Config
}

// NewConfig creates a new configuration instance
func NewConfig(config Config) (ConfigInterface, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &configImpl{
		config: config,
	}, nil
}

// GetAPIKey returns the Vandar API key
func (c *configImpl) GetAPIKey() string {
	return c.config.APIKey
}

// GetBaseURL returns the base URL for the Vandar API
func (c *configImpl) GetBaseURL() string {
	return c.config.BaseURL
}

// IsSandboxMode returns whether the integration is in sandbox mode
func (c *configImpl) IsSandboxMode() bool {
	return c.config.SandboxMode
}

// GetTimeout returns the HTTP client timeout duration
func (c *configImpl) GetTimeout() int {
	return c.config.Timeout
}

// GetCallbackURL returns the URL for payment callbacks
func (c *configImpl) GetCallbackURL() string {
	return c.config.CallbackURL
}

// ConfigWrapper wraps the Config struct to implement ConfigInterface
type ConfigWrapper struct {
	Config
}

// GetAPIKey returns the API key from the wrapped Config
func (c *ConfigWrapper) GetAPIKey() string {
	return c.Config.APIKey
}

// GetBaseURL returns the base URL from the wrapped Config
func (c *ConfigWrapper) GetBaseURL() string {
	return c.Config.BaseURL
}

// IsSandboxMode returns the sandbox mode from the wrapped Config
func (c *ConfigWrapper) IsSandboxMode() bool {
	return c.Config.SandboxMode
}

// GetTimeout returns the timeout from the wrapped Config
func (c *ConfigWrapper) GetTimeout() int {
	return c.Config.Timeout
}

// GetCallbackURL returns the callback URL from the wrapped Config
func (c *ConfigWrapper) GetCallbackURL() string {
	return c.Config.CallbackURL
}
