package balance

import (
	"context"
	"net/http"

	"github.com/jeremyjsx/wallbit-go/transport"
)

const (
	checkingPath = "/api/public/v1/balance/checking"
	stocksPath   = "/api/public/v1/balance/stocks"
)

// Service issues requests against the Wallbit balance endpoints.
type Service struct {
	sender transport.Sender
}

// NewService wires a [Service] to the given [transport.Sender].
func NewService(sender transport.Sender) *Service {
	return &Service{sender: sender}
}

// CheckingBalance is the balance of a single fiat currency held in the
// user's checking account.
type CheckingBalance struct {
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
}

// CheckingBalanceResponse is the top-level envelope for [Service.GetChecking].
type CheckingBalanceResponse struct {
	Data []CheckingBalance `json:"data"`
}

// StockPosition is a single equity holding in the user's stocks account.
type StockPosition struct {
	Symbol string  `json:"symbol"`
	Shares float64 `json:"shares"`
}

// StocksBalanceResponse is the top-level envelope for [Service.GetStocks].
type StocksBalanceResponse struct {
	Data []StockPosition `json:"data"`
}

// GetChecking returns every fiat balance (checking account) held by the
// authenticated user.
func (s *Service) GetChecking(ctx context.Context) (*transport.Response[CheckingBalanceResponse], error) {
	return transport.SendJSON(ctx, s.sender, http.MethodGet, checkingPath, nil, &CheckingBalanceResponse{})
}

// GetStocks returns every equity position held by the authenticated user.
func (s *Service) GetStocks(ctx context.Context) (*transport.Response[StocksBalanceResponse], error) {
	return transport.SendJSON(ctx, s.sender, http.MethodGet, stocksPath, nil, &StocksBalanceResponse{})
}
