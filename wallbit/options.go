package wallbit

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// ErrNilConfig is returned by [NewClientFromConfig] when cfg is nil.
var ErrNilConfig = errors.New("wallbit client: config is required")

// ErrInvalidBaseURL is returned when [WithBaseURL] receives a malformed URL.
var ErrInvalidBaseURL = errors.New("wallbit client: base url must include valid scheme and host")

// ErrInsecureBaseURL is returned when a non-HTTPS base URL is configured
// without the explicit [WithInsecureHTTPForTesting] opt-in.
var ErrInsecureBaseURL = errors.New("wallbit client: non-https base url requires WithInsecureHTTPForTesting")

const defaultBaseURL = "https://api.wallbit.io"

// Option configures a [Client] at construction time. Pass options to
// [NewClient]; later options override earlier ones for the same field.
type Option func(*Config) error

// Config holds the resolved configuration of a [Client]. Construct it
// directly only when using [NewClientFromConfig]; for the common case use
// [NewClient] with functional [Option] values.
type Config struct {
	BaseURL     *url.URL
	HTTPClient  *http.Client
	UserAgent   string
	RetryPolicy RetryPolicy
	Hook        Hook
	// AllowInsecureHTTPForTesting permits HTTP (non-TLS) base URLs.
	// Keep this false in production.
	AllowInsecureHTTPForTesting bool
	// MaxResponseBytes caps the number of bytes the client reads from any
	// HTTP response body. When a response exceeds this bound the client
	// returns [ErrResponseTooLarge] instead of a partial payload, so a
	// hostile or buggy upstream cannot exhaust process memory. Zero or
	// negative values select the default (see [DefaultMaxResponseBytes]).
	MaxResponseBytes int64
}

// RetryPolicy controls how the [Client] retries idempotent requests on
// transport failures and on retryable API responses (HTTP 429 and 5xx).
//
// MaxAttempts is the total number of attempts including the first one.
// A value < 1 means no retries (single attempt). BaseDelay is the initial
// backoff before the first retry; subsequent delays grow exponentially and
// are capped by MaxDelay. The client always honors Retry-After when present.
type RetryPolicy struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

// DefaultRetryPolicy returns the policy used by [NewClient] when none is
// supplied: up to 3 attempts with 250ms base delay, capped at 2s.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts: 3,
		BaseDelay:   250 * time.Millisecond,
		MaxDelay:    2 * time.Second,
	}
}

// Hook is invoked around every HTTP attempt performed by the [Client],
// including retries. Implementations must be safe for concurrent use:
// different requests may call the hook in parallel.
type Hook interface {
	OnRequestStart(*RequestMeta)
	OnRequestDone(*ResponseMeta)
}

// RequestMeta is the context passed to [Hook.OnRequestStart] for a single
// HTTP attempt. Path is the URL path (without the base URL or query string).
type RequestMeta struct {
	Method string
	Path   string
}

// ResponseMeta is the context passed to [Hook.OnRequestDone] for a single
// HTTP attempt. StatusCode is 0 when the transport returned an error before
// receiving a response.
type ResponseMeta struct {
	StatusCode int
	Duration   time.Duration
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
		UserAgent:   defaultUserAgent(),
		RetryPolicy: DefaultRetryPolicy(),
	}, nil
}

// defaultUserAgent builds the canonical User-Agent used when the caller
// does not override it via [WithUserAgent]. The format is
// "wallbit-go-sdk/<version>"; the version comes from [resolveVersion].
// See [Version] for the resolution precedence.
func defaultUserAgent() string {
	return "wallbit-go-sdk/" + resolveVersion()
}

// WithBaseURL overrides the default API base URL. By default only HTTPS is
// accepted; pair with [WithInsecureHTTPForTesting] to allow HTTP for local
// servers and [net/http/httptest] in tests.
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

// WithInsecureHTTPForTesting allows using HTTP base URLs in tests and local
// development. Do not enable this in production: the API key would travel in
// cleartext.
func WithInsecureHTTPForTesting() Option {
	return func(cfg *Config) error {
		cfg.AllowInsecureHTTPForTesting = true
		return nil
	}
}

// WithHTTPClient supplies a custom [*http.Client]. The client is cloned and
// hardened: a [http.Client.CheckRedirect] hook is installed to block
// cross-host redirects so that the API key is never sent to a foreign host.
// Nil is ignored.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(cfg *Config) error {
		if httpClient != nil {
			cfg.HTTPClient = httpClient
		}
		return nil
	}
}

// WithTimeout sets the HTTP client timeout. Non-positive values are ignored.
func WithTimeout(timeout time.Duration) Option {
	return func(cfg *Config) error {
		if timeout > 0 {
			cfg.HTTPClient.Timeout = timeout
		}
		return nil
	}
}

// WithUserAgent overrides the default User-Agent header sent with every
// request. Empty strings are ignored.
func WithUserAgent(userAgent string) Option {
	return func(cfg *Config) error {
		if userAgent != "" {
			cfg.UserAgent = userAgent
		}
		return nil
	}
}

// WithMaxResponseBytes overrides the default cap on HTTP response body
// size. The client reads up to n bytes from res.Body and returns
// [ErrResponseTooLarge] if the server sends more, without decoding the
// partial payload. Values <= 0 are ignored so that [DefaultMaxResponseBytes]
// remains in effect; pass a very large value if you legitimately need to
// consume multi-gigabyte responses.
func WithMaxResponseBytes(n int64) Option {
	return func(cfg *Config) error {
		if n > 0 {
			cfg.MaxResponseBytes = n
		}
		return nil
	}
}

// WithRetryPolicy sets automatic retries for idempotent methods (GET, HEAD,
// DELETE, OPTIONS, TRACE) on transport failures and on retryable API
// responses (HTTP 429 and 5xx; see [IsRetryable]). POST, PATCH and PUT are
// never retried by the client.
func WithRetryPolicy(policy RetryPolicy) Option {
	return func(cfg *Config) error {
		cfg.RetryPolicy = policy
		return nil
	}
}

// WithHook registers a [Hook] invoked around each HTTP attempt (including
// retries). Implementations must be safe for concurrent use.
func WithHook(hook Hook) Option {
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
	if cfg.MaxResponseBytes > 0 {
		out.MaxResponseBytes = cfg.MaxResponseBytes
	}
	out.Hook = cfg.Hook
	out.AllowInsecureHTTPForTesting = cfg.AllowInsecureHTTPForTesting
	return out, nil
}
