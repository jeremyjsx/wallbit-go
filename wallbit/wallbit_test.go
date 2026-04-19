package wallbit

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type testHook struct {
	started int
	done    int
}

func (h *testHook) OnRequestStart(*RequestMeta) { h.started++ }
func (h *testHook) OnRequestDone(*ResponseMeta) { h.done++ }

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
