package apikey

import (
	"context"
	"net/http"

	"github.com/jeremyjsx/wallbit-go/transport"
)

const revokePath = "/api/public/v1/api-key"

type Service struct {
	sender transport.Sender
}

func NewService(sender transport.Sender) *Service {
	return &Service{sender: sender}
}

type RevokeResponse struct {
	Message string `json:"message"`
}

func (s *Service) Revoke(ctx context.Context) (*transport.Response[RevokeResponse], error) {
	return transport.SendJSON(ctx, s.sender, http.MethodDelete, revokePath, nil, &RevokeResponse{})
}
