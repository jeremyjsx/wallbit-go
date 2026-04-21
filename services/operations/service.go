package operations

import (
	"context"
	"net/http"
	"time"

	"github.com/jeremyjsx/wallbit-go/transport"
)

const internalPath = "/api/public/v1/operations/internal"

// Well-known source/destination accounts for internal transfers. The API
// may add new values; these constants document the currently supported set.
const (
	AccountDefault    = "DEFAULT"
	AccountInvestment = "INVESTMENT"
)

// Service issues requests against the Wallbit operations endpoints.
type Service struct {
	sender transport.Sender
}

// NewService wires a [Service] to the given [transport.Sender].
func NewService(sender transport.Sender) *Service {
	return &Service{sender: sender}
}

// InternalRequest moves Amount of Currency from the From account to the To
// account, both accounts belonging to the authenticated user.
type InternalRequest struct {
	Currency string  `json:"currency"`
	From     string  `json:"from"`
	To       string  `json:"to"`
	Amount   float64 `json:"amount"`
}

// InvestmentDepositRequest is the high-level request for
// [Service.DepositInvestment]. It is translated server-side into an
// [InternalRequest] from DEFAULT to INVESTMENT.
type InvestmentDepositRequest struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}

// InvestmentWithdrawRequest is the high-level request for
// [Service.WithdrawInvestment]. It is translated server-side into an
// [InternalRequest] from INVESTMENT to DEFAULT.
type InvestmentWithdrawRequest struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}

// Currency is the embedded currency descriptor carried on each
// [Transaction].
type Currency struct {
	Code  string `json:"code"`
	Alias string `json:"alias"`
}

// Transaction is the row returned by any operations endpoint. The API
// returns this object unwrapped (no data envelope).
type Transaction struct {
	UUID            string    `json:"uuid"`
	Type            string    `json:"type"`
	ExternalAddress *string   `json:"external_address"`
	SourceCurrency  Currency  `json:"source_currency"`
	DestCurrency    Currency  `json:"dest_currency"`
	SourceAmount    float64   `json:"source_amount"`
	DestAmount      float64   `json:"dest_amount"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	Comment         *string   `json:"comment"`
}

// Internal moves funds between two accounts of the authenticated user.
// For the common "default ↔ investment" cases prefer
// [Service.DepositInvestment] and [Service.WithdrawInvestment].
func (s *Service) Internal(ctx context.Context, req InternalRequest) (*transport.Response[Transaction], error) {
	return transport.SendJSON(ctx, s.sender, http.MethodPost, internalPath, req, &Transaction{})
}

// DepositInvestment is a shorthand for [Service.Internal] moving funds from
// the caller's default account to their investment account.
func (s *Service) DepositInvestment(ctx context.Context, req InvestmentDepositRequest) (*transport.Response[Transaction], error) {
	return s.Internal(ctx, InternalRequest{
		Currency: req.Currency,
		From:     AccountDefault,
		To:       AccountInvestment,
		Amount:   req.Amount,
	})
}

// WithdrawInvestment is a shorthand for [Service.Internal] moving funds
// from the caller's investment account back to their default account.
func (s *Service) WithdrawInvestment(ctx context.Context, req InvestmentWithdrawRequest) (*transport.Response[Transaction], error) {
	return s.Internal(ctx, InternalRequest{
		Currency: req.Currency,
		From:     AccountInvestment,
		To:       AccountDefault,
		Amount:   req.Amount,
	})
}
