package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/castiglionimax/process-csv/internal/domain"
)

type (
	ProjectionAccount struct {
		mysql *sql.DB
	}
)

const (
	insertAccount       = "INSERT INTO accounts (id, name, amount, number, last_updated) VALUES (?,?, ?, ?,?)"
	UpdateAccountAmount = "UPDATE accounts SET amount = amount + ? WHERE id = ?;"

	updateSummary = "INSERT INTO summaries (account_id, period, credit, credit_qty, debit, debit_qty,last_updated) VALUES (?, ?, ?, ?,?,?,?) ON DUPLICATE KEY UPDATE credit = credit + VALUES(credit), credit_qty = credit_qty + VALUES(credit_qty), debit = debit + VALUES(debit), debit_qty = debit_qty + VALUES(debit_qty), last_updated =  VALUES(last_updated);"
)

func NewProjection(db *sql.DB) *ProjectionAccount {
	return &ProjectionAccount{db}
}

func (p ProjectionAccount) CreateAccount(ctx context.Context, account domain.Account) error {
	insertStatement, err := p.mysql.Prepare(insertAccount)
	if err != nil {
		return err
	}

	_, err = insertStatement.Exec(account.ID, account.Name, 0, account.Number, time.Now().UTC())
	if err != nil {
		return err
	}
	return nil
}

func (p ProjectionAccount) RegisterTransaction(ctx context.Context, tx domain.Transaction) error {
	insertStatement, err := p.mysql.Prepare(UpdateAccountAmount)
	if err != nil {
		return err
	}
	_, err = insertStatement.Exec(tx.Amount, tx.AccountID)
	if err != nil {
		return err
	}
	return nil
}

func (p ProjectionAccount) RegisterSummary(ctx context.Context, tx domain.Transaction) error {
	insertStatement, err := p.mysql.Prepare(updateSummary)
	if err != nil {
		return err
	}
	f := func(current time.Time) string {
		year, month, _ := current.Date()
		return fmt.Sprintf("%d %s", year, month)
	}
	if tx.Amount > 0 {
		_, err = insertStatement.Exec(tx.AccountID, f(tx.Date), tx.Amount, 1, 0, 0, time.Now().UTC())
	} else {
		_, err = insertStatement.Exec(tx.AccountID, f(tx.Date), 0, 0, tx.Amount, 1, time.Now().UTC())
	}
	return err
}
