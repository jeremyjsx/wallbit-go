package wallbit

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/jeremyjsx/wallbit-go/internal/httpx"
)

// ErrNilConfig is returned by [NewClientFromConfig] when cfg is nil.
var ErrNilConfig = errors.New("wallbit client: config is required")

const defaultBaseURL = "https://api.wallbit.io"

type Option func(*Config) error

type Config struct {
	BaseURL     *url.URL
	HTTPClient  *http.Client
	UserAgent   string
	RetryPolicy httpx.RetryPolicy
	Hook        httpx.Hook
}

type RetryPolicy = httpx.RetryPolicy
type Hook = httpx.Hook
type RequestMeta = httpx.RequestMeta
type ResponseMeta = httpx.ResponseMeta

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

// WithRetryPolicy sets automatic retries for idempotent methods (GET, HEAD, DELETE, OPTIONS, TRACE)
// on transport failures and on retryable API responses (HTTP 429 and 5xx; see errorsx.IsRetryable).
// POST, PATCH, and PUT are never retried by the client.
func WithRetryPolicy(policy httpx.RetryPolicy) Option {
	return func(cfg *Config) error {
		cfg.RetryPolicy = policy
		return nil
	}
}

// WithHook registers a hook invoked around each HTTP attempt (including retries).
// Implementations must be safe for concurrent use; different requests may call the hook in parallel.
func WithHook(hook httpx.Hook) Option {
	return func(cfg *Config) error {
		cfg.Hook = hook
		return nil
	}
}

// mergeClientConfig returns a copy of defaults overlaid with non-zero fields from cfg.
func mergeClientConfig(cfg *Config) (*Config, error) {
	out, err := defaultConfig()
	if err != nil {
		return nil, err
	}
	if cfg.BaseURL != nil {
		out.BaseURL = cfg.BaseURL
	}
	if cfg.HTTPClient != nil {
		out.HTTPClient = cfg.HTTPClient
	}
	if cfg.UserAgent != "" {
		out.UserAgent = cfg.UserAgent
	}
	if cfg.RetryPolicy.MaxAttempts > 0 || cfg.RetryPolicy.BaseDelay > 0 || cfg.RetryPolicy.MaxDelay > 0 {
		out.RetryPolicy = cfg.RetryPolicy
	}
	out.Hook = cfg.Hook
	return out, nil
}
