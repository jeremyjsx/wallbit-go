package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
)

// SendJSON marshals req (when non-nil) and forwards the call to sender,
// wrapping the result into a [*Response][T] with the decoded dest. Pass
// a literal nil for req on GET/DELETE calls. T is inferred from dest.
func SendJSON[T any](ctx context.Context, sender Sender, method, path string, req any, dest *T) (*Response[T], error) {
	var body io.Reader
	if req != nil {
		payload, err := json.Marshal(req)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(payload)
	}
	meta, err := sender.Send(ctx, method, path, body, dest)
	if err != nil {
		return nil, err
	}
	return NewResponse(meta, dest), nil
}
