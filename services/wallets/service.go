package wallets

import (
	"context"
	"net/http"
	"net/url"

	"github.com/jeremyjsx/wallbit-go/transport"
)

const getPath = "/api/public/v1/wallets"

type Service struct {
	sender transport.Sender
}

func NewService(sender transport.Sender) *Service {
	return &Service{sender: sender}
}

type GetRequest struct {
	Currency string
	Network  string
}

type Wallet struct {
	Address      string `json:"address"`
	Network      string `json:"network"`
	CurrencyCode string `json:"currency_code"`
}

type GetResponse struct {
	Data []Wallet `json:"data"`
}

func (s *Service) Get(ctx context.Context, req *GetRequest) (*transport.Response[GetResponse], error) {
	path := getPath
	if req != nil {
		q := url.Values{}
		if req.Currency != "" {
			q.Set("currency", req.Currency)
		}
		if req.Network != "" {
			q.Set("network", req.Network)
		}
		if encoded := q.Encode(); encoded != "" {
			path = path + "?" + encoded
		}
	}

	return transport.SendJSON(ctx, s.sender, http.MethodGet, path, nil, &GetResponse{})
}
