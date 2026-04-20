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

type Service struct {
	sender transport.Sender
}

func NewService(sender transport.Sender) *Service {
	return &Service{sender: sender}
}

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

type GetResponse struct {
	Data ExchangeRate `json:"data"`
}

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
