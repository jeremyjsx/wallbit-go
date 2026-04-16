package balance

import (
	"context"
	"net/http"

	"github.com/jeremyjsx/wallbit-go/internal/httpx"
)

const (
	checkingPath = "/api/public/v1/balance/checking"
	stocksPath   = "/api/public/v1/balance/stocks"
)

type Service struct {
	sender httpx.Sender
}

func NewService(sender httpx.Sender) *Service {
	return &Service{sender: sender}
}

type CheckingBalanceResponse struct {
	Data []map[string]any `json:"data"`
}

type StockPosition struct {
	Symbol string  `json:"symbol"`
	Shares float64 `json:"shares"`
}

type StocksBalanceResponse struct {
	Data []StockPosition `json:"data"`
}

func (s *Service) GetChecking(ctx context.Context) (*CheckingBalanceResponse, error) {
	out := &CheckingBalanceResponse{}
	if err := s.sender.Send(ctx, http.MethodGet, checkingPath, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Service) GetStocks(ctx context.Context) (*StocksBalanceResponse, error) {
	out := &StocksBalanceResponse{}
	if err := s.sender.Send(ctx, http.MethodGet, stocksPath, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}
