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

func (s *Service) Revoke(ctx context.Context) (*RevokeResponse, error) {
	out := &RevokeResponse{}
	if err := s.sender.Send(ctx, http.MethodDelete, revokePath, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}
