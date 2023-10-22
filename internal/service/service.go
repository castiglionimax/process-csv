package service

import (
	"context"
	"errors"

	"github.com/castiglionimax/process-csv/internal/domain"
)

type (
	repository interface {
		CreateAccount(ctx context.Context, account domain.Account) (domain.AccountID, error)
		SaveTransactions(ctx context.Context, transactions []domain.Transaction) error

		SaveTransactionsInDirectory(ctx context.Context, transactions []domain.Transaction) error
		GetTransactionFromDirectory(ctx context.Context) ([]domain.Transaction, error)
		DeleteTransactionsInDirectory(ctx context.Context) error
	}

	Service struct {
		repository repository
	}
)

func NewService(repository repository) (*Service, error) {
	if repository == nil {
		return nil, errors.New("repository should not be nil")
	}
	return &Service{repository: repository}, nil
}

func (s Service) CreateAccount(ctx context.Context, account domain.Account) (domain.AccountID, error) {
	return s.repository.CreateAccount(ctx, account)
}

func (s Service) SaveTransactions(ctx context.Context, transactions []domain.Transaction) error {
	return s.repository.SaveTransactionsInDirectory(ctx, transactions)
}

func (s Service) ProcessFiles(ctx context.Context) error {
	tx, err := s.repository.GetTransactionFromDirectory(ctx)
	if err != nil {
		return err
	}

	if err = s.repository.SaveTransactions(ctx, tx); err != nil {
		return err
	}
	//	return s.repository.DeleteTransactionsInDirectory(ctx)
	return nil
}
