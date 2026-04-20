package wallbit_test

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/jeremyjsx/wallbit-go/wallbit"
)

type capturedRecord struct {
	Level slog.Level
	Msg   string
	Attrs map[string]any
}

type captureHandler struct {
	mu      sync.Mutex
	records []capturedRecord
	level   slog.Level
}

func (h *captureHandler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.level
}

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	attrs := map[string]any{}
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})
	h.mu.Lock()
	h.records = append(h.records, capturedRecord{Level: r.Level, Msg: r.Message, Attrs: attrs})
	h.mu.Unlock()
	return nil
}

func (h *captureHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(_ string) slog.Handler      { return h }

func (h *captureHandler) snapshot() []capturedRecord {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]capturedRecord, len(h.records))
	copy(out, h.records)
	return out
}

func TestSlogHookEmitsStartAndDoneWithAttrs(t *testing.T) {
	t.Parallel()

	h := &captureHandler{level: slog.LevelDebug}
	hook := wallbit.SlogHook(slog.New(h))

	hook.OnRequestStart(&wallbit.RequestMeta{Method: "GET", Path: "/api/public/v1/balance", Attempt: 1})
	hook.OnRequestDone(&wallbit.ResponseMeta{
		Method:     "GET",
		Path:       "/api/public/v1/balance",
		StatusCode: 200,
		Duration:   75 * time.Millisecond,
		Attempt:    1,
	})

	records := h.snapshot()
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	start := records[0]
	if start.Level != slog.LevelDebug {
		t.Fatalf("start level: got %v, want Debug", start.Level)
	}
	if start.Msg != "wallbit.request.start" {
		t.Fatalf("start msg: %q", start.Msg)
	}
	if start.Attrs["method"] != "GET" || start.Attrs["path"] != "/api/public/v1/balance" || start.Attrs["attempt"] != int64(1) {
		t.Fatalf("unexpected start attrs: %+v", start.Attrs)
	}

	done := records[1]
	if done.Level != slog.LevelInfo {
		t.Fatalf("done level: got %v, want Info", done.Level)
	}
	if done.Msg != "wallbit.request.done" {
		t.Fatalf("done msg: %q", done.Msg)
	}
	if done.Attrs["status"] != int64(200) {
		t.Fatalf("done status: %v", done.Attrs["status"])
	}
	if done.Attrs["duration_ms"] != int64(75) {
		t.Fatalf("done duration_ms: %v", done.Attrs["duration_ms"])
	}
}

func TestSlogHookLevelsByStatus(t *testing.T) {
	t.Parallel()

	cases := []struct {
		status int
		want   slog.Level
	}{
		{status: 0, want: slog.LevelError},
		{status: 200, want: slog.LevelInfo},
		{status: 301, want: slog.LevelInfo},
		{status: 400, want: slog.LevelWarn},
		{status: 404, want: slog.LevelWarn},
		{status: 500, want: slog.LevelError},
		{status: 503, want: slog.LevelError},
	}

	for _, tc := range cases {
		h := &captureHandler{level: slog.LevelDebug}
		hook := wallbit.SlogHook(slog.New(h))
		hook.OnRequestDone(&wallbit.ResponseMeta{Method: "GET", Path: "/x", StatusCode: tc.status, Attempt: 1})
		got := h.snapshot()
		if len(got) != 1 {
			t.Fatalf("status=%d: expected 1 record, got %d", tc.status, len(got))
		}
		if got[0].Level != tc.want {
			t.Fatalf("status=%d: level got %v, want %v", tc.status, got[0].Level, tc.want)
		}
	}
}

func TestSlogHookNilLoggerUsesDefault(t *testing.T) {
	t.Parallel()

	hook := wallbit.SlogHook(nil)
	// Must not panic.
	hook.OnRequestStart(&wallbit.RequestMeta{Method: "GET", Path: "/", Attempt: 1})
	hook.OnRequestDone(&wallbit.ResponseMeta{Method: "GET", Path: "/", StatusCode: 200, Attempt: 1})
}

func TestSlogHookRespectsHandlerLevel(t *testing.T) {
	t.Parallel()

	h := &captureHandler{level: slog.LevelInfo}
	hook := wallbit.SlogHook(slog.New(h))

	hook.OnRequestStart(&wallbit.RequestMeta{Method: "GET", Path: "/", Attempt: 1})
	hook.OnRequestDone(&wallbit.ResponseMeta{Method: "GET", Path: "/", StatusCode: 200, Attempt: 1})

	records := h.snapshot()
	if len(records) != 1 {
		t.Fatalf("expected only done to be emitted at Info level, got %d records", len(records))
	}
	if records[0].Msg != "wallbit.request.done" {
		t.Fatalf("unexpected msg: %q", records[0].Msg)
	}
}
