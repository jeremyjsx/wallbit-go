package wallbit

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type testHook struct {
	started         int
	done            int
	startAttempts   []int
	doneAttempts    []int
	doneStatusCodes []int
}

func (h *testHook) OnRequestStart(m *RequestMeta) {
	h.started++
	h.startAttempts = append(h.startAttempts, m.Attempt)
}
func (h *testHook) OnRequestDone(m *ResponseMeta) {
	h.done++
	h.doneAttempts = append(h.doneAttempts, m.Attempt)
	h.doneStatusCodes = append(h.doneStatusCodes, m.StatusCode)
}

func TestNewClientAndOptions(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); got != "wallbit-test" {
			t.Fatalf("unexpected user-agent %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	hook := &testHook{}
	c, err := NewClient(
		"test-key",
		WithBaseURL(server.URL),
		WithInsecureHTTPForTesting(),
		WithUserAgent("wallbit-test"),
		WithRetryPolicy(RetryPolicy{MaxAttempts: 1, BaseDelay: time.Millisecond, MaxDelay: 10 * time.Millisecond}),
		WithHook(hook),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := c.Balance.GetChecking(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hook.started != 1 || hook.done != 1 {
		t.Fatalf("unexpected hook counters: started=%d done=%d", hook.started, hook.done)
	}
	// A successful single-attempt request must surface Attempt=1 (1-indexed)
	// on both callbacks; regressions to the legacy zero-value would be
	// indistinguishable from an unset field in user code.
	if len(hook.startAttempts) != 1 || hook.startAttempts[0] != 1 {
		t.Fatalf("OnRequestStart attempts: got %v, want [1]", hook.startAttempts)
	}
	if len(hook.doneAttempts) != 1 || hook.doneAttempts[0] != 1 {
		t.Fatalf("OnRequestDone attempts: got %v, want [1]", hook.doneAttempts)
	}
}

func TestHookSeesAttemptIncrementingAcrossRetries(t *testing.T) {
	t.Parallel()

	// The server returns 503 on the first two calls and 200 on the third,
	// driving the client through three hook cycles. This exercises the
	// full retry path including the 5xx branch so we confirm the attempt
	// counter is emitted on BOTH the error-response bookkeeping and the
	// final success, not just the success path.
	var hits int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	hook := &testHook{}
	c, err := NewClient(
		"test-key",
		WithBaseURL(server.URL),
		WithInsecureHTTPForTesting(),
		WithRetryPolicy(RetryPolicy{MaxAttempts: 3, BaseDelay: time.Millisecond, MaxDelay: 2 * time.Millisecond}),
		WithHook(hook),
	)
	if err != nil {
		t.Fatalf("unexpected client construction error: %v", err)
	}

	if _, err := c.Balance.GetChecking(context.Background()); err != nil {
		t.Fatalf("unexpected error after retry recovery: %v", err)
	}

	want := []int{1, 2, 3}
	if !slicesEqual(hook.startAttempts, want) {
		t.Fatalf("OnRequestStart attempts: got %v, want %v", hook.startAttempts, want)
	}
	if !slicesEqual(hook.doneAttempts, want) {
		t.Fatalf("OnRequestDone attempts: got %v, want %v", hook.doneAttempts, want)
	}
	// Status codes cross-check that attempt 3 is the one that observed
	// 200, confirming we count attempts over the actual retry loop rather
	// than emitting a static sequence.
	wantCodes := []int{503, 503, 200}
	if !slicesEqual(hook.doneStatusCodes, wantCodes) {
		t.Fatalf("OnRequestDone status codes: got %v, want %v", hook.doneStatusCodes, wantCodes)
	}
}

func slicesEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestWithBaseURLRejectsHTTPByDefault(t *testing.T) {
	t.Parallel()

	_, err := NewClient("test-key", WithBaseURL("http://127.0.0.1:8080"))
	if !errors.Is(err, ErrInsecureBaseURL) {
		t.Fatalf("expected ErrInsecureBaseURL, got %v", err)
	}
}

func TestClientBlocksCrossHostRedirect(t *testing.T) {
	t.Parallel()

	evil := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("redirect target should not be reached; got %s", r.URL.String())
	}))
	defer evil.Close()

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, evil.URL+"/steal", http.StatusFound)
	}))
	defer origin.Close()

	c, err := NewClient(
		"test-key",
		WithBaseURL(origin.URL),
		WithInsecureHTTPForTesting(),
		WithRetryPolicy(RetryPolicy{MaxAttempts: 1, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.Balance.GetChecking(context.Background())
	if err == nil {
		t.Fatal("expected error due to blocked cross-host redirect")
	}
}

func TestClientEnforcesMaxResponseBytes(t *testing.T) {
	t.Parallel()

	const limit = 128
	big := `{"data":[` + strings.Repeat(`"x",`, limit) + `"x"]}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-ID", "req-too-large")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(big))
	}))
	defer server.Close()

	c, err := NewClient(
		"test-key",
		WithBaseURL(server.URL),
		WithInsecureHTTPForTesting(),
		WithRetryPolicy(RetryPolicy{MaxAttempts: 1, BaseDelay: time.Millisecond, MaxDelay: 10 * time.Millisecond}),
		WithMaxResponseBytes(limit),
	)
	if err != nil {
		t.Fatalf("unexpected client construction error: %v", err)
	}

	_, err = c.Balance.GetChecking(context.Background())
	if !errors.Is(err, ErrResponseTooLarge) {
		t.Fatalf("expected ErrResponseTooLarge, got %v", err)
	}
}

func TestClientAcceptsResponseUpToMaxResponseBytes(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	c, err := NewClient(
		"test-key",
		WithBaseURL(server.URL),
		WithInsecureHTTPForTesting(),
		WithRetryPolicy(RetryPolicy{MaxAttempts: 1, BaseDelay: time.Millisecond, MaxDelay: 10 * time.Millisecond}),
		WithMaxResponseBytes(64),
	)
	if err != nil {
		t.Fatalf("unexpected client construction error: %v", err)
	}

	if _, err := c.Balance.GetChecking(context.Background()); err != nil {
		t.Fatalf("unexpected error on small response: %v", err)
	}
}

func TestClientUsesDefaultMaxResponseBytesWhenUnset(t *testing.T) {
	t.Parallel()

	c, err := NewClient("test-key")
	if err != nil {
		t.Fatalf("unexpected client construction error: %v", err)
	}
	if got := c.maxResponseBytes(); got != DefaultMaxResponseBytes {
		t.Fatalf("maxResponseBytes() = %d, want %d", got, DefaultMaxResponseBytes)
	}
}
