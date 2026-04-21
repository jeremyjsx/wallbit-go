package cards

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/jeremyjsx/wallbit-go/transport"
)

const listPath = "/api/public/v1/cards"
const updateStatusPathFormat = "/api/public/v1/cards/%s/status"

// Card status values accepted by the update-status endpoint.
const (
	StatusActive    = "ACTIVE"
	StatusSuspended = "SUSPENDED"
)

// ErrEmptyCardUUID is returned by [Service.Block] and [Service.Unblock] when cardUUID is empty or whitespace-only.
var ErrEmptyCardUUID = errors.New("cards: card uuid is required")

// Service issues requests against the Wallbit cards endpoints.
type Service struct {
	sender transport.Sender
}

// NewService wires a [Service] to the given [transport.Sender].
func NewService(sender transport.Sender) *Service {
	return &Service{sender: sender}
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

// CardStatus is the minimal card row returned by the update-status endpoint.
type CardStatus struct {
	UUID   string `json:"uuid"`
	Status string `json:"status"`
}

// Card describes a card row as returned by [Service.List]. Expiration is
// nil when the API does not expose it.
type Card struct {
	UUID        string  `json:"uuid"`
	Status      string  `json:"status"`
	CardType    string  `json:"card_type"`
	CardNetwork string  `json:"card_network"`
	CardLast4   string  `json:"card_last4"`
	Expiration  *string `json:"expiration"`
}

// ListResponse is the top-level envelope for [Service.List].
type ListResponse struct {
	Data []Card `json:"data"`
}

// UpdateStatusResponse is the top-level envelope for [Service.Block] and
// [Service.Unblock].
type UpdateStatusResponse struct {
	Data CardStatus `json:"data"`
}

// List returns every card visible to the authenticated user, regardless of
// status.
func (s *Service) List(ctx context.Context) (*transport.Response[ListResponse], error) {
	return transport.SendJSON(ctx, s.sender, http.MethodGet, listPath, nil, &ListResponse{})
}

// Block suspends the card identified by cardUUID. It returns
// [ErrEmptyCardUUID] if cardUUID is empty or whitespace-only.
func (s *Service) Block(ctx context.Context, cardUUID string) (*transport.Response[UpdateStatusResponse], error) {
	return s.updateStatus(ctx, cardUUID, StatusSuspended)
}

// Unblock re-activates the card identified by cardUUID. It returns
// [ErrEmptyCardUUID] if cardUUID is empty or whitespace-only.
func (s *Service) Unblock(ctx context.Context, cardUUID string) (*transport.Response[UpdateStatusResponse], error) {
	return s.updateStatus(ctx, cardUUID, StatusActive)
}

func (s *Service) updateStatus(ctx context.Context, cardUUID string, status string) (*transport.Response[UpdateStatusResponse], error) {
	if strings.TrimSpace(cardUUID) == "" {
		return nil, ErrEmptyCardUUID
	}
	path := fmt.Sprintf(updateStatusPathFormat, cardUUID)
	return transport.SendJSON(ctx, s.sender, http.MethodPatch, path, updateStatusRequest{Status: status}, &UpdateStatusResponse{})
}
