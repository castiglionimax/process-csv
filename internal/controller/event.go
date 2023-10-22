package controller

import (
	"context"
	"encoding/json"

	"github.com/castiglionimax/process-csv/internal/domain"
)

type (
	eventService interface {
		CreateAccount(ctx context.Context, account domain.Account) error
		RegisterTransaction(ctx context.Context, tx domain.Transaction) error
		RegisterSummary(ctx context.Context, tx domain.Transaction) error
	}

	EventHandler struct {
		eventService eventService
	}
)

func NewEventHandler(eventService eventService) EventHandler {
	return EventHandler{eventService: eventService}
}

func (h EventHandler) SaveAccount(ctx context.Context, body []byte) error {
	var acc domain.Account

	if err := json.Unmarshal(body, &acc); err != nil {
		return err
	}

	return h.eventService.CreateAccount(ctx, acc)
}

func (h EventHandler) RegisterTransaction(ctx context.Context, body []byte) error {
	var tx domain.Transaction

	if err := json.Unmarshal(body, &tx); err != nil {
		return err
	}

	return h.eventService.RegisterTransaction(ctx, tx)
}

func (h EventHandler) RegisterSummary(ctx context.Context, body []byte) error {
	var tx domain.Transaction

	if err := json.Unmarshal(body, &tx); err != nil {
		return err
	}

	return h.eventService.RegisterSummary(ctx, tx)
}
