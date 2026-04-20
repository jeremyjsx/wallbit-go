package roboadvisor

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/jeremyjsx/wallbit-go/transport"
)

const (
	balancePath  = "/api/public/v1/roboadvisor/balance"
	depositPath  = "/api/public/v1/roboadvisor/deposit"
	withdrawPath = "/api/public/v1/roboadvisor/withdraw"
)

type Service struct {
	sender transport.Sender
}

func NewService(sender transport.Sender) *Service {
	return &Service{sender: sender}
}

type RiskProfile struct {
	RiskLevel int    `json:"risk_level"`
	Name      string `json:"name"`
}

type Performance struct {
	NetDeposits      float64 `json:"net_deposits"`
	NetProfits       float64 `json:"net_profits"`
	TotalDeposits    float64 `json:"total_deposits"`
	TotalWithdrawals float64 `json:"total_withdrawals"`
}

type Allocation struct {
	Cash       float64 `json:"cash"`
	Securities float64 `json:"securities"`
}

type Asset struct {
	Symbol                   string  `json:"symbol"`
	Shares                   float64 `json:"shares"`
	MarketValue              float64 `json:"market_value"`
	Price                    float64 `json:"price"`
	DailyVariationPercentage float64 `json:"daily_variation_percentage"`
	Weight                   float64 `json:"weight"`
	Logo                     string  `json:"logo"`
}

type Portfolio struct {
	ID                      int          `json:"id"`
	Label                   *string      `json:"label"`
	Category                *string      `json:"category"`
	PortfolioType           string       `json:"portfolio_type"`
	Balance                 float64      `json:"balance"`
	PortfolioValue          float64      `json:"portfolio_value"`
	Cash                    float64      `json:"cash"`
	CashAvailableWithdrawal float64      `json:"cash_available_withdrawal"`
	RiskProfile             *RiskProfile `json:"risk_profile"`
	Performance             Performance  `json:"performance"`
	Assets                  []Asset      `json:"assets"`
	Allocation              Allocation   `json:"allocation"`
	HasPendingTransactions  bool         `json:"has_pending_transactions"`
}

type GetBalanceResponse struct {
	Data []Portfolio `json:"data"`
}

type AccountType string

const (
	AccountTypeDefault    AccountType = "DEFAULT"
	AccountTypeInvestment AccountType = "INVESTMENT"
)

type DepositRequest struct {
	RoboAdvisorID int         `json:"robo_advisor_id"`
	Amount        float64     `json:"amount"`
	From          AccountType `json:"from"`
}

type WithdrawRequest struct {
	RoboAdvisorID int         `json:"robo_advisor_id"`
	Amount        float64     `json:"amount"`
	To            AccountType `json:"to"`
}

type Transaction struct {
	UUID      string  `json:"uuid"`
	Type      string  `json:"type"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at"`
}

type DepositResponse struct {
	Data Transaction `json:"data"`
}

type WithdrawResponse struct {
	Data Transaction `json:"data"`
}

func (s *Service) GetBalance(ctx context.Context) (*transport.Response[GetBalanceResponse], error) {
	out := &GetBalanceResponse{}
	meta, err := s.sender.Send(ctx, http.MethodGet, balancePath, nil, out)
	if err != nil {
		return nil, err
	}
	return transport.NewResponse(meta, out), nil
}

func (s *Service) Deposit(ctx context.Context, req DepositRequest) (*transport.Response[DepositResponse], error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	out := &DepositResponse{}
	meta, err := s.sender.Send(ctx, http.MethodPost, depositPath, bytes.NewBuffer(payload), out)
	if err != nil {
		return nil, err
	}
	return transport.NewResponse(meta, out), nil
}

func (s *Service) Withdraw(ctx context.Context, req WithdrawRequest) (*transport.Response[WithdrawResponse], error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	out := &WithdrawResponse{}
	meta, err := s.sender.Send(ctx, http.MethodPost, withdrawPath, bytes.NewBuffer(payload), out)
	if err != nil {
		return nil, err
	}
	return transport.NewResponse(meta, out), nil
}
