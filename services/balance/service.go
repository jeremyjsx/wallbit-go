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

type Service struct {
	sender transport.Sender
}

func NewService(sender transport.Sender) *Service {
	return &Service{sender: sender}
}

type CheckingBalance struct {
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
}

type CheckingBalanceResponse struct {
	Data []CheckingBalance `json:"data"`
}

type StockPosition struct {
	Symbol string  `json:"symbol"`
	Shares float64 `json:"shares"`
}

type StocksBalanceResponse struct {
	Data []StockPosition `json:"data"`
}

func (s *Service) GetChecking(ctx context.Context) (*transport.Response[CheckingBalanceResponse], error) {
	return transport.SendJSON(ctx, s.sender, http.MethodGet, checkingPath, nil, &CheckingBalanceResponse{})
}

func (s *Service) GetStocks(ctx context.Context) (*transport.Response[StocksBalanceResponse], error) {
	return transport.SendJSON(ctx, s.sender, http.MethodGet, stocksPath, nil, &StocksBalanceResponse{})
}
