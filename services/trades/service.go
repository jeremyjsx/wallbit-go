package trades

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/jeremyjsx/wallbit-go/transport"
)

const createPath = "/api/public/v1/trades"

type Service struct {
	sender transport.Sender
}

func NewService(sender transport.Sender) *Service {
	return &Service{sender: sender}
}

type CreateRequest struct {
	Symbol      string   `json:"symbol"`
	Direction   string   `json:"direction"`
	Currency    string   `json:"currency"`
	OrderType   string   `json:"order_type"`
	Amount      *float64 `json:"amount,omitempty"`
	Shares      *float64 `json:"shares,omitempty"`
	StopPrice   *float64 `json:"stop_price,omitempty"`
	LimitPrice  *float64 `json:"limit_price,omitempty"`
	TimeInForce *string  `json:"time_in_force,omitempty"`
}

type Trade struct {
	Symbol      string   `json:"symbol"`
	Direction   string   `json:"direction"`
	Amount      float64  `json:"amount"`
	Shares      float64  `json:"shares"`
	Status      string   `json:"status"`
	OrderType   string   `json:"order_type"`
	LimitPrice  *float64 `json:"limit_price"`
	StopPrice   *float64 `json:"stop_price"`
	TimeInForce *string  `json:"time_in_force"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

type CreateResponse struct {
	Data Trade `json:"data"`
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (*transport.Response[CreateResponse], error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	out := &CreateResponse{}
	meta, err := s.sender.Send(ctx, http.MethodPost, createPath, bytes.NewBuffer(payload), out)
	if err != nil {
		return nil, err
	}
	return transport.NewResponse(meta, out), nil
}
