package controller

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/render"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/castiglionimax/process-csv/internal/domain"
	pkgError "github.com/castiglionimax/process-csv/pkg/error"
)

type (
	Service interface {
		CreateAccount(ctx context.Context, account domain.Account) (domain.AccountID, error)
		SaveTransactions(ctx context.Context, transactions []domain.Transaction) error
		ProcessFiles(ctx context.Context) error
	}

	Controller struct {
		service Service
	}
)

func NewController(service Service) (*Controller, error) {
	if service == nil {
		return nil, errors.New("service should not be nil")
	}
	return &Controller{
		service: service,
	}, nil
}

func (c Controller) CreateAccount(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, pkgError.ErrReadingBody.Error(), http.StatusBadRequest)
		return
	}

	var req domain.Account

	if err = json.Unmarshal(data, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	account, err := c.service.CreateAccount(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, account)
	w.WriteHeader(http.StatusCreated)
}

func (c Controller) UploadHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("csv")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

	var transactions []domain.Transaction

	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		parsedTimestamp, err := strconv.ParseInt(line[1], 10, 64)
		if err != nil {
			continue
		}

		parsedAmount, err := strconv.ParseFloat(line[2], 64)
		if err != nil || parsedAmount == 0 {
			continue
		}

		gotten := domain.Transaction{
			AccountID: domain.AccountID(line[0]),
			Date:      time.Unix(parsedTimestamp, 0).UTC(),
			Amount:    parsedAmount,
		}

		transactions = append(transactions, gotten)
	}

	if err = c.service.SaveTransactions(r.Context(), transactions); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (c Controller) ProcessFiles(w http.ResponseWriter, r *http.Request) {
	if err := c.service.ProcessFiles(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (c Controller) CreateCsv(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, pkgError.ErrReadingBody.Error(), http.StatusBadRequest)
		return
	}

	var req []struct {
		AccountId string `json:"account_id"`
		Timestamp int64  `json:"timestamp"`
		Amount    string `json:"amount"`
	}

	if err = json.Unmarshal(data, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var transactions []domain.Transaction

	for _, object := range req {
		parsedAmount, err := strconv.ParseFloat(object.Amount, 64)
		if err != nil || parsedAmount == 0 {
			fmt.Printf("error reading line %v", err)
			continue
		}

		gotten := domain.Transaction{
			AccountID: domain.AccountID(object.AccountId),
			Date:      time.Unix(object.Timestamp, 0),
			Amount:    parsedAmount,
		}
		transactions = append(transactions, gotten)
	}

	if err = c.service.SaveTransactions(r.Context(), transactions); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
