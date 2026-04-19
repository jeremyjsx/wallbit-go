package wallbit

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/jeremyjsx/wallbit-go/internal/httpx"
)

// ErrNilConfig is returned by [NewClientFromConfig] when cfg is nil.
var ErrNilConfig = errors.New("wallbit client: config is required")

// ErrInvalidBaseURL is returned when WithBaseURL receives a malformed URL.
var ErrInvalidBaseURL = errors.New("wallbit client: base url must include valid scheme and host")

// ErrInsecureBaseURL is returned when a non-HTTPS base URL is configured without explicit opt-in.
var ErrInsecureBaseURL = errors.New("wallbit client: non-https base url requires WithInsecureHTTPForTesting")

const defaultBaseURL = "https://api.wallbit.io"

type Option func(*Config) error

type Config struct {
	BaseURL     *url.URL
	HTTPClient  *http.Client
	UserAgent   string
	RetryPolicy httpx.RetryPolicy
	Hook        httpx.Hook
	// AllowInsecureHTTPForTesting permits HTTP (non-TLS) base URLs.
	// Keep this false in production.
	AllowInsecureHTTPForTesting bool
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
			return fmt.Errorf("%w: %v", ErrInvalidBaseURL, err)
		}
		if parsed.Scheme == "" || parsed.Host == "" {
			return ErrInvalidBaseURL
		}
		switch parsed.Scheme {
		case "https", "http":
		default:
			return fmt.Errorf("%w: unsupported scheme %q", ErrInvalidBaseURL, parsed.Scheme)
		}
		cfg.BaseURL = parsed
		return nil
	}
}

// validateBaseURL enforces HTTPS unless AllowInsecureHTTPForTesting is set.
// Runs after all options apply so option order does not matter.
func validateBaseURL(cfg *Config) error {
	if cfg.BaseURL == nil {
		return ErrInvalidBaseURL
	}
	if cfg.BaseURL.Scheme == "http" && !cfg.AllowInsecureHTTPForTesting {
		return ErrInsecureBaseURL
	}
	return nil
}

// WithInsecureHTTPForTesting allows using HTTP base URLs in tests/local development.
// Do not enable this in production.
func WithInsecureHTTPForTesting() Option {
	return func(cfg *Config) error {
		cfg.AllowInsecureHTTPForTesting = true
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
	out.AllowInsecureHTTPForTesting = cfg.AllowInsecureHTTPForTesting
	return out, nil
}
