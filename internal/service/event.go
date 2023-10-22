package service

import (
	"context"
	"github.com/castiglionimax/process-csv/internal/domain"
)

type (
	projection interface {
		CreateAccount(ctx context.Context, account domain.Account) error
		RegisterTransaction(ctx context.Context, tx domain.Transaction) error
		RegisterSummary(ctx context.Context, tx domain.Transaction) error
	}

	EventService struct {
		repository projection
	}
)

func NewEventService(projection projection) *EventService {
	return &EventService{repository: projection}
}

func (e EventService) CreateAccount(ctx context.Context, account domain.Account) error {
	return e.repository.CreateAccount(ctx, account)
}

func (e EventService) RegisterTransaction(ctx context.Context, tx domain.Transaction) error {
	return e.repository.RegisterTransaction(ctx, tx)
}

func (e EventService) RegisterSummary(ctx context.Context, tx domain.Transaction) error {
	return e.repository.RegisterSummary(ctx, tx)
}
