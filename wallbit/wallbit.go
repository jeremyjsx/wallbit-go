package wallbit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jeremyjsx/wallbit-go/services/accountdetails"
	"github.com/jeremyjsx/wallbit-go/services/apikey"
	"github.com/jeremyjsx/wallbit-go/services/assets"
	"github.com/jeremyjsx/wallbit-go/services/balance"
	"github.com/jeremyjsx/wallbit-go/services/cards"
	"github.com/jeremyjsx/wallbit-go/services/fees"
	"github.com/jeremyjsx/wallbit-go/services/operations"
	"github.com/jeremyjsx/wallbit-go/services/rates"
	"github.com/jeremyjsx/wallbit-go/services/roboadvisor"
	"github.com/jeremyjsx/wallbit-go/services/trades"
	"github.com/jeremyjsx/wallbit-go/services/transactions"
	"github.com/jeremyjsx/wallbit-go/services/wallets"
	"github.com/jeremyjsx/wallbit-go/transport"
)

var ErrMissingAPIKey = errors.New("wallbit client requires a non-empty api key")

// ErrResponseTooLarge is returned by the client when an HTTP response body
// exceeds the configured byte cap (see [Config.MaxResponseBytes] and
// [WithMaxResponseBytes]). The partial payload is discarded because a
// truncated body cannot be distinguished from a well-formed short one by
// the JSON decoder.
var ErrResponseTooLarge = errors.New("wallbit client: response body exceeds configured size limit")

// DefaultMaxResponseBytes is the cap applied to HTTP response bodies when
// neither [Config.MaxResponseBytes] nor [WithMaxResponseBytes] supplies a
// positive value.
const DefaultMaxResponseBytes int64 = 10 << 20

type Client struct {
	apiKey string
	cfg    *Config
	sender transport.Sender

	Balance *balance.Service

	Transactions *transactions.Service

	APIKey *apikey.Service

	Trades *trades.Service

	Fees *fees.Service

	AccountDetails *accountdetails.Service

	Wallets *wallets.Service

	Assets *assets.Service

	Operations *operations.Service

	RoboAdvisor *roboadvisor.Service

	Cards *cards.Service

	Rates *rates.Service
}

func NewClient(apiKey string, opts ...Option) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, ErrMissingAPIKey
	}

	cfg, err := defaultConfig()
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}
	if err := validateBaseURL(cfg); err != nil {
		return nil, err
	}
	cfg = secureConfig(cfg)

	c := &Client{
		apiKey: strings.TrimSpace(apiKey),
		cfg:    cfg,
	}
	c.sender = senderAdapter{client: c}
	wireServices(c)

	return c, nil
}

// NewClientFromConfig builds a client from cfg merged on top of the same defaults as [NewClient].
// Use this for a single configuration value block instead of many [Option] calls.
// Nil cfg returns [ErrNilConfig]. Retry policy: if all RetryPolicy fields are zero, defaults apply.
func NewClientFromConfig(apiKey string, cfg *Config) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, ErrMissingAPIKey
	}
	if cfg == nil {
		return nil, ErrNilConfig
	}
	merged, err := mergeClientConfig(cfg)
	if err != nil {
		return nil, err
	}
	if err := validateBaseURL(merged); err != nil {
		return nil, err
	}
	merged = secureConfig(merged)
	c := &Client{
		apiKey: strings.TrimSpace(apiKey),
		cfg:    merged,
	}
	c.sender = senderAdapter{client: c}
	wireServices(c)
	return c, nil
}

func wireServices(c *Client) {
	c.Balance = balance.NewService(c.sender)
	c.Transactions = transactions.NewService(c.sender)
	c.APIKey = apikey.NewService(c.sender)
	c.Trades = trades.NewService(c.sender)
	c.Fees = fees.NewService(c.sender)
	c.AccountDetails = accountdetails.NewService(c.sender)
	c.Wallets = wallets.NewService(c.sender)
	c.Assets = assets.NewService(c.sender)
	c.Operations = operations.NewService(c.sender)
	c.RoboAdvisor = roboadvisor.NewService(c.sender)
	c.Cards = cards.NewService(c.sender)
	c.Rates = rates.NewService(c.sender)
}

func (c *Client) Config() *Config {
	return c.cfg
}

func (c *Client) newRequest(ctx context.Context, method string, path string, body io.Reader) (*http.Request, error) {
	endpoint, err := c.cfg.BaseURL.Parse(path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func (c *Client) send(ctx context.Context, method string, path string, body io.Reader, dest any) (*transport.Metadata, error) {
	req, err := c.newRequest(ctx, method, path, body)
	if err != nil {
		return nil, err
	}

	return c.do(req, dest)
}

func (c *Client) do(req *http.Request, dest any) (*transport.Metadata, error) {
	ctx := req.Context()
	max := c.maxAttempts()

	// attempt is 0-indexed; hooks see the 1-indexed value via attemptNumber.
	for attempt := range max {
		attemptNumber := attempt + 1
		reqTry := req.Clone(ctx)
		c.emitRequestStart(reqTry, attemptNumber)

		start := time.Now()
		res, err := c.cfg.HTTPClient.Do(reqTry)
		dur := time.Since(start)

		statusCode := 0
		if res != nil {
			statusCode = res.StatusCode
		}
		c.emitRequestDone(statusCode, dur, attemptNumber)

		if err != nil {
			if attempt < max-1 && isIdempotentHTTPMethod(req.Method) {
				wait := c.retryWaitBeforeNextAttempt(nil, nil, attempt)
				if err := sleepContext(ctx, wait); err != nil {
					return nil, err
				}
				continue
			}
			return nil, err
		}

		limit := c.maxResponseBytes()
		body, rerr := io.ReadAll(io.LimitReader(res.Body, limit+1))
		res.Body.Close()
		if rerr != nil {
			return nil, rerr
		}

		requestID := res.Header.Get("X-Request-ID")
		meta := &transport.Metadata{
			StatusCode: statusCode,
			Header:     res.Header,
			RequestID:  requestID,
		}

		if int64(len(body)) > limit {
			return meta, fmt.Errorf("%w: limit %d bytes", ErrResponseTooLarge, limit)
		}

		if statusCode >= 400 {
			apiErr := ErrorFromHTTP(statusCode, requestID, body)
			if attempt < max-1 && isIdempotentHTTPMethod(req.Method) && IsRetryable(apiErr) {
				wait := c.retryWaitBeforeNextAttempt(res, apiErr, attempt)
				if err := sleepContext(ctx, wait); err != nil {
					return nil, err
				}
				continue
			}
			return meta, apiErr
		}

		if err := decodeBody(body, dest, statusCode); err != nil {
			return meta, err
		}
		return meta, nil
	}
	return nil, errors.New("wallbit client: internal error: retry loop exited without return")
}

// emitRequestStart fires the OnRequestStart hook when one is configured.
// Centralized so do() doesn't carry hook plumbing inline.
func (c *Client) emitRequestStart(req *http.Request, attempt int) {
	h := c.cfg.Hook
	if h == nil {
		return
	}
	path := ""
	if req.URL != nil {
		path = req.URL.Path
	}
	h.OnRequestStart(&RequestMeta{Method: req.Method, Path: path, Attempt: attempt})
}

// emitRequestDone fires the OnRequestDone hook when one is configured.
func (c *Client) emitRequestDone(statusCode int, dur time.Duration, attempt int) {
	h := c.cfg.Hook
	if h == nil {
		return
	}
	h.OnRequestDone(&ResponseMeta{StatusCode: statusCode, Duration: dur, Attempt: attempt})
}

// decodeBody unmarshals body into dest unless the response carries no
// payload to decode (nil dest, empty body, or 204 No Content). io.ReadAll
// already returned a usable slice when the LimitReader hit EOF, so a
// length check is sufficient to cover both empty bodies and 204s.
func decodeBody(body []byte, dest any, statusCode int) error {
	if dest == nil || len(body) == 0 || statusCode == http.StatusNoContent {
		return nil
	}
	return json.Unmarshal(body, dest)
}

type senderAdapter struct {
	client *Client
}

func (s senderAdapter) Send(ctx context.Context, method string, path string, body io.Reader, dest any) (*transport.Metadata, error) {
	return s.client.send(ctx, method, path, body, dest)
}

func secureConfig(cfg *Config) *Config {
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}
	cloned := *cfg.HTTPClient
	prevRedirect := cloned.CheckRedirect
	cloned.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) > 0 {
			origin := via[0].URL
			if origin != nil && req.URL != nil && !strings.EqualFold(req.URL.Host, origin.Host) {
				return http.ErrUseLastResponse
			}
		}
		if prevRedirect != nil {
			return prevRedirect(req, via)
		}
		return nil
	}
	cfg.HTTPClient = &cloned
	return cfg
}
