package apikey

import (
	"context"
	"net/http"

	"github.com/jeremyjsx/wallbit-go/transport"
)

const revokePath = "/api/public/v1/api-key"

// Service issues requests against the Wallbit API-key endpoint.
type Service struct {
	sender transport.Sender
}

// NewService wires a [Service] to the given [transport.Sender].
func NewService(sender transport.Sender) *Service {
	return &Service{sender: sender}
}

// RevokeResponse is the top-level envelope for [Service.Revoke].
type RevokeResponse struct {
	Message string `json:"message"`
}

// Revoke invalidates the API key carried by the client. Subsequent calls
// with the same key will fail authentication; issue a new key out-of-band
// to keep using the SDK.
func (s *Service) Revoke(ctx context.Context) (*transport.Response[RevokeResponse], error) {
	return transport.SendJSON(ctx, s.sender, http.MethodDelete, revokePath, nil, &RevokeResponse{})
}
