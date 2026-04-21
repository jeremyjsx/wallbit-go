package wallets

import (
	"context"
	"net/http"
	"net/url"

	"github.com/jeremyjsx/wallbit-go/transport"
)

const getPath = "/api/public/v1/wallets"

// Service issues requests against the Wallbit wallets endpoint.
type Service struct {
	sender transport.Sender
}

// NewService wires a [Service] to the given [transport.Sender].
func NewService(sender transport.Sender) *Service {
	return &Service{sender: sender}
}

// GetRequest filters the wallet list. Both fields are optional; an empty
// field is omitted from the query string and matches every value.
type GetRequest struct {
	Currency string
	Network  string
}

// Wallet is a single deposit address row returned by [Service.Get].
type Wallet struct {
	Address      string `json:"address"`
	Network      string `json:"network"`
	CurrencyCode string `json:"currency_code"`
}

// GetResponse is the top-level envelope for [Service.Get].
type GetResponse struct {
	Data []Wallet `json:"data"`
}

// Get returns the deposit addresses for the authenticated user. A nil req
// returns every wallet; set Currency or Network to narrow the result.
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
