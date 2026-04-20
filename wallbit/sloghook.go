package wallbit

import (
	"context"
	"log/slog"
)

// SlogHook returns a [Hook] that emits structured logs to the given
// [*slog.Logger] for every HTTP attempt the client performs. When logger is
// nil, [slog.Default] is used.
//
// Each attempt produces two records (filter with the logger's level):
//
//   - "wallbit.request.start" at [slog.LevelDebug]
//   - "wallbit.request.done"  at [slog.LevelInfo]  for 2xx/3xx
//     [slog.LevelWarn]  for 4xx
//     [slog.LevelError] for 5xx or transport errors (status == 0)
//
// Attributes on both records: method, path, attempt.
// Additional attributes on the done record: status, duration_ms.
//
// The returned hook is safe for concurrent use.
func SlogHook(logger *slog.Logger) Hook {
	if logger == nil {
		logger = slog.Default()
	}
	return &slogHook{logger: logger}
}

type slogHook struct {
	logger *slog.Logger
}

func (h *slogHook) OnRequestStart(m *RequestMeta) {
	h.logger.LogAttrs(context.Background(), slog.LevelDebug, "wallbit.request.start",
		slog.String("method", m.Method),
		slog.String("path", m.Path),
		slog.Int("attempt", m.Attempt),
	)
}

func (h *slogHook) OnRequestDone(m *ResponseMeta) {
	h.logger.LogAttrs(context.Background(), slogLevelForStatus(m.StatusCode), "wallbit.request.done",
		slog.String("method", m.Method),
		slog.String("path", m.Path),
		slog.Int("attempt", m.Attempt),
		slog.Int("status", m.StatusCode),
		slog.Int64("duration_ms", m.Duration.Milliseconds()),
	)
}

func slogLevelForStatus(status int) slog.Level {
	switch {
	case status == 0, status >= 500:
		return slog.LevelError
	case status >= 400:
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}
