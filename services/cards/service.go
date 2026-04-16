package cards

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jeremyjsx/wallbit-go/internal/httpx"
)

const updateStatusPathFormat = "/api/public/v1/cards/%s/status"

const (
	StatusActive    = "ACTIVE"
	StatusSuspended = "SUSPENDED"
)

type Service struct {
	sender httpx.Sender
}

func NewService(sender httpx.Sender) *Service {
	return &Service{sender: sender}
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

type CardStatus struct {
	UUID   string `json:"uuid"`
	Status string `json:"status"`
}

type UpdateStatusResponse struct {
	Data CardStatus `json:"data"`
}

func (s *Service) Block(ctx context.Context, cardUUID string) (*UpdateStatusResponse, error) {
	return s.updateStatus(ctx, cardUUID, StatusSuspended)
}

func (s *Service) Unblock(ctx context.Context, cardUUID string) (*UpdateStatusResponse, error) {
	return s.updateStatus(ctx, cardUUID, StatusActive)
}

func (s *Service) updateStatus(ctx context.Context, cardUUID string, status string) (*UpdateStatusResponse, error) {
	payload, err := json.Marshal(updateStatusRequest{Status: status})
	if err != nil {
		return nil, err
	}

	out := &UpdateStatusResponse{}
	path := fmt.Sprintf(updateStatusPathFormat, cardUUID)
	if err := s.sender.Send(ctx, http.MethodPatch, path, bytes.NewBuffer(payload), out); err != nil {
		return nil, err
	}
	return out, nil
}
