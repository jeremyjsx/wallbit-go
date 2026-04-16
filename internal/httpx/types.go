package httpx

import (
	"context"
	"io"
	"time"
)

type Hook interface {
	OnRequestStart(*RequestMeta)
	OnRequestDone(*ResponseMeta)
}

type RequestMeta struct {
	Method string
	Path   string
}

type ResponseMeta struct {
	StatusCode int
	Duration   time.Duration
}

type RetryPolicy struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

type Sender interface {
	Send(ctx context.Context, method string, path string, body io.Reader, dest any) error
}

func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts: 3,
		BaseDelay:   250 * time.Millisecond,
		MaxDelay:    2 * time.Second,
	}
}
