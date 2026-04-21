package assets

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/jeremyjsx/wallbit-go/transport"
)

const listPath = "/api/public/v1/assets"

// ErrEmptySymbol is returned by [Service.Get] when symbol is empty or whitespace-only.
var ErrEmptySymbol = errors.New("assets: symbol is required")

// Service issues requests against the Wallbit assets endpoints.
type Service struct {
	sender transport.Sender
}

// NewService wires a [Service] to the given [transport.Sender].
func NewService(sender transport.Sender) *Service {
	return &Service{sender: sender}
}

// ListRequest parameterises a call to [Service.List]. Every field is
// optional; unset values are omitted from the query string. Use
// [github.com/jeremyjsx/wallbit-go/wallbit.Ptr] to set Page and Limit.
type ListRequest struct {
	Category string
	Search   string
	Page     *int
	Limit    *int
}

// Dividend captures the most recent dividend event announced for an asset,
// when available. All fields are nil when the instrument does not pay
// dividends or when the API has not published the event yet.
type Dividend struct {
	Amount      *float64 `json:"amount"`
	Yield       *float64 `json:"yield"`
	ExDate      *string  `json:"ex_date"`
	PaymentDate *string  `json:"payment_date"`
}

// Asset is a tradable instrument row. Pointer fields carry the server's
// own omitempty: they are nil whenever the API would omit them.
type Asset struct {
	Symbol        string    `json:"symbol"`
	Name          string    `json:"name"`
	Price         float64   `json:"price"`
	AssetType     *string   `json:"asset_type"`
	Exchange      *string   `json:"exchange"`
	Sector        *string   `json:"sector"`
	MarketCapM    *string   `json:"market_cap_m"`
	Description   *string   `json:"description"`
	DescriptionES *string   `json:"description_es"`
	Country       *string   `json:"country"`
	CEO           *string   `json:"ceo"`
	Employees     *string   `json:"employees"`
	LogoURL       string    `json:"logo_url"`
	Dividend      *Dividend `json:"dividend"`
}

// ListResponse is the top-level envelope for [Service.List]. CurrentPage
// equals Pages on the last page; Count is the total number of rows across
// every page.
type ListResponse struct {
	Data        []Asset `json:"data"`
	Pages       int     `json:"pages"`
	CurrentPage int     `json:"current_page"`
	Count       int     `json:"count"`
}

// GetResponse is the top-level envelope for [Service.Get].
type GetResponse struct {
	Data Asset `json:"data"`
}

// Get fetches a single asset by its ticker symbol. An empty symbol returns
// [ErrEmptySymbol].
func (s *Service) Get(ctx context.Context, symbol string) (*transport.Response[GetResponse], error) {
	if strings.TrimSpace(symbol) == "" {
		return nil, ErrEmptySymbol
	}
	path := fmt.Sprintf("%s/%s", listPath, url.PathEscape(symbol))
	return transport.SendJSON(ctx, s.sender, http.MethodGet, path, nil, &GetResponse{})
}

// List fetches a single page of assets matching req's filters. A nil req
// returns the first page with server defaults. For lazy iteration over
// every page, use [Service.ListAll].
func (s *Service) List(ctx context.Context, req *ListRequest) (*transport.Response[ListResponse], error) {
	path := listPath
	if req != nil {
		q := url.Values{}
		if req.Category != "" {
			q.Set("category", req.Category)
		}
		if req.Search != "" {
			q.Set("search", req.Search)
		}
		if req.Page != nil {
			q.Set("page", strconv.Itoa(*req.Page))
		}
		if req.Limit != nil {
			q.Set("limit", strconv.Itoa(*req.Limit))
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
// or on the first error. Errors are yielded with a zero-value Asset. The
// caller's *ListRequest is not mutated.
func (s *Service) ListAll(ctx context.Context, req *ListRequest) iter.Seq2[Asset, error] {
	return func(yield func(Asset, error) bool) {
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
				yield(Asset{}, err)
				return
			}
			pageReq.Page = &page
			out, err := s.List(ctx, &pageReq)
			if err != nil {
				yield(Asset{}, err)
				return
			}
			for _, a := range out.Payload.Data {
				if !yield(a, nil) {
					return
				}
			}
			if len(out.Payload.Data) == 0 || out.Payload.CurrentPage >= out.Payload.Pages {
				return
			}
			page++
		}
	}
}
