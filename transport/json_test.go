package transport

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
)

type fakeSender struct {
	gotCtx    context.Context //nolint:containedctx // test stub mirrors the interface
	gotMethod string
	gotPath   string
	gotBody   []byte

	respond func(dest any) (*Metadata, error)
}

func (f *fakeSender) Send(ctx context.Context, method, path string, body io.Reader, dest any) (*Metadata, error) {
	f.gotCtx = ctx
	f.gotMethod = method
	f.gotPath = path
	if body != nil {
		b, err := io.ReadAll(body)
		if err != nil {
			return nil, err
		}
		f.gotBody = b
	}
	if f.respond != nil {
		return f.respond(dest)
	}
	return &Metadata{StatusCode: http.StatusOK}, nil
}

type pingRequest struct {
	Ping string `json:"ping"`
}

type pingResponse struct {
	Pong string `json:"pong"`
}

func TestSendJSONSkipsBodyWhenRequestIsNil(t *testing.T) {
	t.Parallel()

	f := &fakeSender{}
	out := &pingResponse{}

	res, err := SendJSON(context.Background(), f, http.MethodGet, "/ping", nil, out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.gotBody != nil {
		t.Fatalf("expected nil body, got %q", f.gotBody)
	}
	if res == nil || res.Payload != out {
		t.Fatalf("expected response to wrap the same dest pointer, got %+v", res)
	}
}

func TestSendJSONMarshalsNonNilRequest(t *testing.T) {
	t.Parallel()

	const path = "/operations/internal"
	f := &fakeSender{
		respond: func(dest any) (*Metadata, error) {
			*(dest.(*pingResponse)) = pingResponse{Pong: "ok"}
			return &Metadata{StatusCode: http.StatusOK, RequestID: "req-123"}, nil
		},
	}
	out := &pingResponse{}

	res, err := SendJSON(context.Background(), f, http.MethodPost, path, pingRequest{Ping: "hi"}, out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.gotMethod != http.MethodPost || f.gotPath != path {
		t.Fatalf("method/path: got %s %s, want POST %s", f.gotMethod, f.gotPath, path)
	}
	if string(f.gotBody) != `{"ping":"hi"}` {
		t.Fatalf("body: got %q, want %q", f.gotBody, `{"ping":"hi"}`)
	}
	if res.Payload.Pong != "ok" {
		t.Fatalf("response payload: got %+v, want Pong=ok", res.Payload)
	}
	if res.RequestID != "req-123" || res.StatusCode != http.StatusOK {
		t.Fatalf("meta not propagated: got StatusCode=%d RequestID=%q", res.StatusCode, res.RequestID)
	}
}

func TestSendJSONReturnsMarshalErrorWithoutCallingSender(t *testing.T) {
	t.Parallel()

	f := &fakeSender{
		respond: func(dest any) (*Metadata, error) {
			t.Fatal("sender must not be invoked when marshal fails")
			return nil, nil
		},
	}

	_, err := SendJSON(context.Background(), f, http.MethodPost, "/x", make(chan int), &pingResponse{})
	if err == nil {
		t.Fatal("expected marshal error, got nil")
	}
	var utErr *json.UnsupportedTypeError
	if !errors.As(err, &utErr) {
		t.Fatalf("expected *json.UnsupportedTypeError, got %T: %v", err, err)
	}
}

func TestSendJSONPropagatesSenderError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("simulated transport failure")
	f := &fakeSender{
		respond: func(dest any) (*Metadata, error) {
			return nil, sentinel
		},
	}

	res, err := SendJSON(context.Background(), f, http.MethodPost, "/x", pingRequest{Ping: "x"}, &pingResponse{})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if res != nil {
		t.Fatalf("expected nil response on sender failure, got %+v", res)
	}
}
