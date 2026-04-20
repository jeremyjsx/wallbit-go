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

const (
	StatusActive    = "ACTIVE"
	StatusSuspended = "SUSPENDED"
)

// ErrEmptyCardUUID is returned by [Service.Block] and [Service.Unblock] when cardUUID is empty or whitespace-only.
var ErrEmptyCardUUID = errors.New("cards: card uuid is required")

type Service struct {
	sender transport.Sender
}

func NewService(sender transport.Sender) *Service {
	return &Service{sender: sender}
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

type CardStatus struct {
	UUID   string `json:"uuid"`
	Status string `json:"status"`
}

type Card struct {
	UUID        string  `json:"uuid"`
	Status      string  `json:"status"`
	CardType    string  `json:"card_type"`
	CardNetwork string  `json:"card_network"`
	CardLast4   string  `json:"card_last4"`
	Expiration  *string `json:"expiration"`
}

type ListResponse struct {
	Data []Card `json:"data"`
}

type UpdateStatusResponse struct {
	Data CardStatus `json:"data"`
}

func (s *Service) List(ctx context.Context) (*transport.Response[ListResponse], error) {
	return transport.SendJSON(ctx, s.sender, http.MethodGet, listPath, nil, &ListResponse{})
}

func (s *Service) Block(ctx context.Context, cardUUID string) (*transport.Response[UpdateStatusResponse], error) {
	return s.updateStatus(ctx, cardUUID, StatusSuspended)
}

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
