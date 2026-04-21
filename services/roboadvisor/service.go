package roboadvisor

import (
	"context"
	"net/http"
	"time"

	"github.com/jeremyjsx/wallbit-go/transport"
)

const (
	balancePath  = "/api/public/v1/roboadvisor/balance"
	depositPath  = "/api/public/v1/roboadvisor/deposit"
	withdrawPath = "/api/public/v1/roboadvisor/withdraw"
)

// Service issues requests against the Wallbit robo-advisor endpoints.
type Service struct {
	sender transport.Sender
}

// NewService wires a [Service] to the given [transport.Sender].
func NewService(sender transport.Sender) *Service {
	return &Service{sender: sender}
}

// RiskProfile describes the portfolio's risk tier as configured on the API
// side. RiskLevel is a numeric bucket, Name is the human label.
type RiskProfile struct {
	RiskLevel int    `json:"risk_level"`
	Name      string `json:"name"`
}

// Performance aggregates cash flow and P&L for a portfolio over its full
// lifetime, denominated in the portfolio's settlement currency (USD).
type Performance struct {
	NetDeposits      float64 `json:"net_deposits"`
	NetProfits       float64 `json:"net_profits"`
	TotalDeposits    float64 `json:"total_deposits"`
	TotalWithdrawals float64 `json:"total_withdrawals"`
}

// Allocation is the current cash-vs-securities split of a portfolio.
type Allocation struct {
	Cash       float64 `json:"cash"`
	Securities float64 `json:"securities"`
}

// Asset is a single holding within a [Portfolio].
type Asset struct {
	Symbol                   string  `json:"symbol"`
	Shares                   float64 `json:"shares"`
	MarketValue              float64 `json:"market_value"`
	Price                    float64 `json:"price"`
	DailyVariationPercentage float64 `json:"daily_variation_percentage"`
	Weight                   float64 `json:"weight"`
	Logo                     string  `json:"logo"`
}

// Portfolio is a managed robo-advisor account. Label and Category are
// user-assigned metadata and may be nil when never set.
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

// GetBalanceResponse is the top-level envelope for [Service.GetBalance].
type GetBalanceResponse struct {
	Data []Portfolio `json:"data"`
}

// AccountType is the source/destination of a robo-advisor movement: the
// caller's default cash account or a specific investment portfolio.
type AccountType string

// Known [AccountType] values.
const (
	AccountTypeDefault    AccountType = "DEFAULT"
	AccountTypeInvestment AccountType = "INVESTMENT"
)

// DepositRequest funds the portfolio identified by RoboAdvisorID with
// Amount units pulled from the From account.
type DepositRequest struct {
	RoboAdvisorID int         `json:"robo_advisor_id"`
	Amount        float64     `json:"amount"`
	From          AccountType `json:"from"`
}

// WithdrawRequest moves Amount units out of the portfolio identified by
// RoboAdvisorID into the To account.
type WithdrawRequest struct {
	RoboAdvisorID int         `json:"robo_advisor_id"`
	Amount        float64     `json:"amount"`
	To            AccountType `json:"to"`
}

// Transaction describes a single deposit or withdrawal against a portfolio.
type Transaction struct {
	UUID      string    `json:"uuid"`
	Type      string    `json:"type"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// DepositResponse is the top-level envelope for [Service.Deposit].
type DepositResponse struct {
	Data Transaction `json:"data"`
}

// WithdrawResponse is the top-level envelope for [Service.Withdraw].
type WithdrawResponse struct {
	Data Transaction `json:"data"`
}

// GetBalance returns every robo-advisor portfolio visible to the caller,
// including the cash balance, current valuation and asset breakdown.
func (s *Service) GetBalance(ctx context.Context) (*transport.Response[GetBalanceResponse], error) {
	return transport.SendJSON(ctx, s.sender, http.MethodGet, balancePath, nil, &GetBalanceResponse{})
}

// Deposit moves funds from the caller's source account into the portfolio
// identified in req.
func (s *Service) Deposit(ctx context.Context, req DepositRequest) (*transport.Response[DepositResponse], error) {
	return transport.SendJSON(ctx, s.sender, http.MethodPost, depositPath, req, &DepositResponse{})
}

// Withdraw moves funds out of the portfolio identified in req into the
// caller's destination account.
func (s *Service) Withdraw(ctx context.Context, req WithdrawRequest) (*transport.Response[WithdrawResponse], error) {
	return transport.SendJSON(ctx, s.sender, http.MethodPost, withdrawPath, req, &WithdrawResponse{})
}
