package repository

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/castiglionimax/process-csv/internal/domain"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"
	"github.com/harranali/mailing"
	"github.com/minio/minio-go/v7"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

type (
	producer interface {
		Produce(msg *kafka.Message, deliveryChan chan kafka.Event) error
	}

	Repository struct {
		producer *kafka.Producer
		mongo    *mongo.Client
		topic    string
		mysql    *sql.DB
		minio    *minio.Client
		mailer   *mailing.Mailer
	}
)

func NewRepository(producer *kafka.Producer, topic string, mongo *mongo.Client, mysql *sql.DB, minio *minio.Client, mailer *mailing.Mailer) *Repository {
	return &Repository{producer: producer, topic: topic, mongo: mongo, mysql: mysql, minio: minio, mailer: mailer}
}

const (
	createAccount = "account_created"
	saveDebit     = "debit_saved"
	saveCredit    = "credit_saved"
)

func (r Repository) CreateAccount(ctx context.Context, account domain.Account) (domain.AccountID, error) {
	account.ID = domain.AccountID(uuid.New().String())
	eventModel := newModel(createAccount, account.ID.String(), account, calculateHash(account))

	if err := r.apply(ctx, eventModel); err != nil {
		return "", err
	}
	return account.ID, nil
}

func (r Repository) SaveTransactions(ctx context.Context, transactions []domain.Transaction) error {
	var (
		transactionType string
		err             error
	)
	for _, transaction := range transactions {
		if transaction.Amount > 0 {
			transactionType = saveCredit
		} else {
			transactionType = saveDebit
		}
		errApply := r.apply(ctx, newModel(transactionType, transaction.AccountID.String(), transaction, calculateHash(transaction)))
		if errApply != nil {
			err = errors.Join(err, errApply)
		}
	}
	return err
}

func (r Repository) apply(ctx context.Context, event interface{}) error {
	coll := r.mongo.Database("event_store").Collection("accounts")

	session, err := r.mongo.StartSession()
	if err != nil {
		return err
	}

	defer session.EndSession(ctx)

	if err = session.StartTransaction(); err != nil {
		return err
	}

	return mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		_, err = coll.InsertOne(ctx, event)
		if err != nil {
			return err
		}

		enAccount, err := json.Marshal(event)
		if err != nil {
			return err
		}

		deliveryChan := make(chan kafka.Event, 10000)
		err = r.producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &r.topic, Partition: kafka.PartitionAny},
			Value:          enAccount},
			deliveryChan,
		)

		if err != nil {
			_ = session.AbortTransaction(ctx)
			return err
		}

		if err = session.CommitTransaction(sc); err != nil {
			return err
		}
		return nil
	})
}

func calculateHash[T any](x T) string {
	data, err := json.Marshal(x)
	if err != nil {
		log.Default().Printf("fail trying to calculateHash: %s", err)
	}
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}
