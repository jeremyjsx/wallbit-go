package client

import (
	"net/http"
	"net/url"
	"time"

	"github.com/jeremyjsx/wallbit-go/internal/httpx"
)

const defaultBaseURL = "https://api.wallbit.io"

type Option func(*Config) error

type Config struct {
	BaseURL     *url.URL
	HTTPClient  *http.Client
	UserAgent   string
	RetryPolicy httpx.RetryPolicy
	Hook        httpx.Hook
}

func defaultConfig() (*Config, error) {
	parsed, err := url.Parse(defaultBaseURL)
	if err != nil {
		return nil, err
	}

	return &Config{
		BaseURL: parsed,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		UserAgent:   "wallbit-go-sdk/0.1.0",
		RetryPolicy: httpx.DefaultRetryPolicy(),
	}, nil
}

func WithBaseURL(raw string) Option {
	return func(cfg *Config) error {
		parsed, err := url.Parse(raw)
		if err != nil {
			return err
		}
		cfg.BaseURL = parsed
		return nil
	}
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(cfg *Config) error {
		if httpClient != nil {
			cfg.HTTPClient = httpClient
		}
		return nil
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(cfg *Config) error {
		if timeout > 0 {
			cfg.HTTPClient.Timeout = timeout
		}
		return nil
	}
}

func WithUserAgent(userAgent string) Option {
	return func(cfg *Config) error {
		if userAgent != "" {
			cfg.UserAgent = userAgent
		}
		return nil
	}
}

func WithRetryPolicy(policy httpx.RetryPolicy) Option {
	return func(cfg *Config) error {
		cfg.RetryPolicy = policy
		return nil
	}
}

func WithHook(hook httpx.Hook) Option {
	return func(cfg *Config) error {
		cfg.Hook = hook
		return nil
	}
}
