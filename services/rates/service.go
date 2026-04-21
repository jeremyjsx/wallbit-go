package rates

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jeremyjsx/wallbit-go/transport"
)

const getPath = "/api/public/v1/rates"

// ErrEmptyCurrency is returned by [Service.Get] when source_currency or dest_currency is empty or whitespace-only.
var ErrEmptyCurrency = errors.New("rates: source_currency and dest_currency are required")

// Service issues requests against the Wallbit rates endpoint.
type Service struct {
	sender transport.Sender
}

// NewService wires a [Service] to the given [transport.Sender].
func NewService(sender transport.Sender) *Service {
	return &Service{sender: sender}
}

// GetRequest parameterises a rate lookup by source and destination currency.
// Both fields are required (ISO-like currency codes, e.g. "USD", "ARS").
type GetRequest struct {
	SourceCurrency string
	DestCurrency   string
}

// ExchangeRate is the row returned by the public API. UpdatedAt is nil for
// identity pairs (e.g. USD→USD returns rate 1.0 with no stored row).
type ExchangeRate struct {
	SourceCurrency string     `json:"source_currency"`
	DestCurrency   string     `json:"dest_currency"`
	Pair           string     `json:"pair"`
	Rate           float64    `json:"rate"`
	UpdatedAt      *time.Time `json:"updated_at"`
}

// GetResponse is the top-level envelope for [Service.Get].
type GetResponse struct {
	Data ExchangeRate `json:"data"`
}

// Get fetches the current exchange rate between req.SourceCurrency and
// req.DestCurrency. It returns [ErrEmptyCurrency] if either field is blank.
func (s *Service) Get(ctx context.Context, req GetRequest) (*transport.Response[GetResponse], error) {
	if strings.TrimSpace(req.SourceCurrency) == "" || strings.TrimSpace(req.DestCurrency) == "" {
		return nil, ErrEmptyCurrency
	}

	q := url.Values{}
	q.Set("source_currency", req.SourceCurrency)
	q.Set("dest_currency", req.DestCurrency)
	path := getPath + "?" + q.Encode()

	return transport.SendJSON(ctx, s.sender, http.MethodGet, path, nil, &GetResponse{})
}
