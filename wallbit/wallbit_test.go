package wallbit

import (
	"context"
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
