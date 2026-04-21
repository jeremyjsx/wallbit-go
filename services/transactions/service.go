package transactions

import (
	"context"
	"iter"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jeremyjsx/wallbit-go/transport"
)

const listPath = "/api/public/v1/transactions"

// Service issues requests against the Wallbit transactions endpoint.
type Service struct {
	sender transport.Sender
}

// NewService wires a [Service] to the given [transport.Sender].
func NewService(sender transport.Sender) *Service {
	return &Service{sender: sender}
}

// ListRequest parameterises a call to [Service.List]. Every field is
// optional; nil/zero values are omitted from the query string. Page and
// Limit are pointers so the caller can distinguish "unset" from "0"; use
// [github.com/jeremyjsx/wallbit-go/wallbit.Ptr] as a convenience.
type ListRequest struct {
	Page       *int
	Limit      *int
	Status     string
	Type       string
	Currency   string
	FromDate   *time.Time
	ToDate     *time.Time
	FromAmount *float64
	ToAmount   *float64
}

// CurrencyRef is the embedded currency descriptor carried on each
// [Transaction]. Code is the ISO-like currency identifier, Alias is the
// human-friendly label surfaced by the API.
type CurrencyRef struct {
	Code  string `json:"code"`
	Alias string `json:"alias"`
}

// Transaction is a single row returned by [Service.List]. ExternalAddress
// and Comment are nil when the API omits them.
type Transaction struct {
	UUID            string      `json:"uuid"`
	Type            string      `json:"type"`
	ExternalAddress *string     `json:"external_address"`
	SourceCurrency  CurrencyRef `json:"source_currency"`
	DestCurrency    CurrencyRef `json:"dest_currency"`
	SourceAmount    float64     `json:"source_amount"`
	DestAmount      float64     `json:"dest_amount"`
	Status          string      `json:"status"`
	CreatedAt       time.Time   `json:"created_at"`
	Comment         *string     `json:"comment"`
}

// ListData is the paginated payload embedded in [ListResponse]. CurrentPage
// equals Pages on the last page; Count is the total number of rows across
// every page.
type ListData struct {
	Data        []Transaction `json:"data"`
	Pages       int           `json:"pages"`
	CurrentPage int           `json:"current_page"`
	Count       int           `json:"count"`
}

// ListResponse is the top-level envelope for [Service.List].
type ListResponse struct {
	Data ListData `json:"data"`
}

// List fetches a single page of transactions matching req's filters. A nil
// req returns the first page with server defaults. For lazy iteration over
// every page, use [Service.ListAll].
func (s *Service) List(ctx context.Context, req *ListRequest) (*transport.Response[ListResponse], error) {
	path := listPath
	if req != nil {
		q := url.Values{}
		if req.Page != nil {
			q.Set("page", strconv.Itoa(*req.Page))
		}
		if req.Limit != nil {
			q.Set("limit", strconv.Itoa(*req.Limit))
		}
		if req.Status != "" {
			q.Set("status", req.Status)
		}
		if req.Type != "" {
			q.Set("type", req.Type)
		}
		if req.Currency != "" {
			q.Set("currency", req.Currency)
		}
		if req.FromDate != nil {
			q.Set("from_date", req.FromDate.Format("2006-01-02"))
		}
		if req.ToDate != nil {
			q.Set("to_date", req.ToDate.Format("2006-01-02"))
		}
		if req.FromAmount != nil {
			q.Set("from_amount", strconv.FormatFloat(*req.FromAmount, 'f', -1, 64))
		}
		if req.ToAmount != nil {
			q.Set("to_amount", strconv.FormatFloat(*req.ToAmount, 'f', -1, 64))
		}
		if encoded := q.Encode(); encoded != "" {
			path = path + "?" + encoded
		}
	}

	return transport.SendJSON(ctx, s.sender, http.MethodGet, path, nil, &ListResponse{})
}

// ListAll returns an iterator that walks every page of results for the given
// filters, starting from req.Page if set (default 1) and advancing one page
// per batch until current_page >= pages. It issues one HTTP request per
// page; pass a higher Limit to reduce round trips.
//
// The iterator stops early when yield returns false, when ctx is cancelled,
// or on the first error. Errors are yielded with a zero-value Transaction.
// The caller's *ListRequest is not mutated.
func (s *Service) ListAll(ctx context.Context, req *ListRequest) iter.Seq2[Transaction, error] {
	return func(yield func(Transaction, error) bool) {
		var pageReq ListRequest
		if req != nil {
			pageReq = *req
		}
		page := 1
		if pageReq.Page != nil {
			page = *pageReq.Page
		}
		for {
			if err := ctx.Err(); err != nil {
				yield(Transaction{}, err)
				return
			}
			pageReq.Page = &page
			out, err := s.List(ctx, &pageReq)
			if err != nil {
				yield(Transaction{}, err)
				return
			}
			for _, tx := range out.Payload.Data.Data {
				if !yield(tx, nil) {
					return
				}
			}
			if len(out.Payload.Data.Data) == 0 || out.Payload.Data.CurrentPage >= out.Payload.Data.Pages {
				return
			}
			page++
		}
	}
}
