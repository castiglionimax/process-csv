package repository

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/castiglionimax/process-csv/internal/domain"
)

const (
	bucketTransactions = "transactions"
	contentType        = "application/zip"
)

func (r Repository) SaveTransactionsInDirectory(ctx context.Context, transactions []domain.Transaction) error {
	var csvData [][]string
	csvData = append(csvData, []string{"account_id", "timestamp", "amount"})
	for _, tx := range transactions {
		csvData = append(csvData, []string{
			string(tx.AccountID),
			strconv.FormatInt(tx.Date.Unix(), 10),
			strconv.FormatFloat(tx.Amount, 'f', -1, 64)})
	}
	csvContent := convertToCSV(csvData)
	_, err := r.minio.PutObject(ctx,
		bucketTransactions,
		fmt.Sprintf("%s%s", uuid.New().String(), ".csv"),
		bytes.NewReader([]byte(csvContent)),
		int64(len(csvContent)),
		minio.PutObjectOptions{ContentType: contentType})
	return err
}

func convertToCSV(data [][]string) string {
	var csvLines []string
	for _, line := range data {
		csvLines = append(csvLines, strings.Join(line, ","))
	}
	return strings.Join(csvLines, "\n")
}

func (r Repository) GetTransactionFromDirectory(ctx context.Context) ([]domain.Transaction, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	objectCh := r.minio.ListObjects(ctx, bucketTransactions, minio.ListObjectsOptions{
		Recursive: true,
	})

	var transactions []domain.Transaction

	for object := range objectCh {
		if object.Err != nil {
			log.Println(object.Err)
			continue
		}
		cvs, err := r.minio.GetObject(ctx, bucketTransactions, object.Key, minio.GetObjectOptions{})
		if err != nil {
			log.Println(err)
			continue
		}

		reader := csv.NewReader(cvs)
		lines, err := reader.ReadAll()
		if err != nil {
			break
		}

		for _, row := range lines {

			parseTimeStamp, err := strconv.ParseInt(row[1], 10, 64)
			if err != nil {
				continue
			}

			parsedAmount, err := strconv.ParseFloat(row[2], 64)
			if err != nil || parsedAmount == 0 {
				continue
			}

			gotten := domain.Transaction{
				AccountID: domain.AccountID(row[0]),
				Date:      time.Unix(parseTimeStamp, 0),
				Amount:    parsedAmount,
			}
			transactions = append(transactions, gotten)
		}

	}
	return transactions, nil
}
func (r Repository) DeleteTransactionsInDirectory(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	objectCh := r.minio.ListObjects(ctx, bucketTransactions, minio.ListObjectsOptions{
		Recursive: true,
	})

	for object := range objectCh {
		err := r.minio.RemoveObject(ctx, bucketTransactions, object.Key, minio.RemoveObjectOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}
