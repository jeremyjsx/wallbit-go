package wallbit

import (
	"context"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func isIdempotentHTTPMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodDelete, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

func parseRetryAfterSeconds(h http.Header) (secs int64, ok bool) {
	raw := strings.TrimSpace(h.Get("Retry-After"))
	if raw == "" {
		return 0, false
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}

func sleepContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// retryWaitBeforeNextAttempt returns the backoff before attempt failureIndex+1 (0 = first retry after one failure).
func (c *Client) retryWaitBeforeNextAttempt(res *http.Response, apiErr *Error, failureIndex int) time.Duration {
	p := c.cfg.RetryPolicy
	base := p.BaseDelay
	maxD := p.MaxDelay
	if base <= 0 {
		base = 250 * time.Millisecond
	}
	if maxD <= 0 {
		maxD = 2 * time.Second
	}

	var fromAPI time.Duration
	if apiErr != nil {
		if d := apiErr.RetryAfter(); d > 0 {
			fromAPI = d
		}
	}
	if res != nil {
		if sec, ok := parseRetryAfterSeconds(res.Header); ok {
			if d := time.Duration(sec) * time.Second; d > fromAPI {
				fromAPI = d
			}
		}
	}
	if fromAPI > 0 {
		if maxD > 0 && fromAPI > maxD {
			return maxD
		}
		return fromAPI
	}

	d := base
	for i := 0; i < failureIndex; i++ {
		next := d * 2
		if next > maxD {
			d = maxD
			break
		}
		d = next
	}
	return jitter(d)
}

// jitter applies equal jitter to a deterministic backoff duration so that
// many clients driven by the same retry schedule do not synchronize on the
// same wake-up instants (the classic thundering-herd problem when a
// dependency recovers from a 5xx incident).
//
// A zero or negative input returns zero.
func jitter(d time.Duration) time.Duration {
	if d <= 0 {
		return 0
	}
	half := d / 2
	if half <= 0 {
		return d
	}

	return half + time.Duration(rand.Int64N(int64(half)+1))
}

func (c *Client) maxAttempts() int {
	n := c.cfg.RetryPolicy.MaxAttempts
	if n < 1 {
		return 1
	}
	return n
}

// maxResponseBytes returns the effective body-size cap for this client,
// falling back to [DefaultMaxResponseBytes] when the configuration does
// not specify a positive value. Centralizing the default here keeps the
// read path in [Client.do] free of branching and makes it trivial for
// tests to inject a small limit via [WithMaxResponseBytes].
func (c *Client) maxResponseBytes() int64 {
	if c.cfg.MaxResponseBytes > 0 {
		return c.cfg.MaxResponseBytes
	}
	return DefaultMaxResponseBytes
}
