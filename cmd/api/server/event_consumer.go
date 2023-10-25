package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/kafka"

	"github.com/castiglionimax/process-csv/internal/controller"
)

type (
	eventConsumer struct {
		consumer *kafka.Consumer
		controller.EventHandler
	}
)

const (
	createAccount = "account_created"

	saveDebit  = "debit_saved"
	saveCredit = "credit_saved"
)

func newConsumerEvent(consumer *kafka.Consumer) *eventConsumer {
	return &eventConsumer{
		consumer:     consumer,
		EventHandler: controller.NewEventHandler(resolverEventService()),
	}
}

func (c eventConsumer) HandlerAccount() {
	topics := "EventQueue"
	err := c.consumer.Subscribe(topics, nil)
	if err != nil {
		fmt.Println(err)
	}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	run := true
	for run == true {
		select {
		case sig := <-signals:
			fmt.Printf("Terminando debido a la señal: %v\n", sig)
			run = false
		default:
			msg, err := c.consumer.ReadMessage(-1)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error al recibir mensaje: %v\n", err)
				continue
			}

			var body struct {
				EventID     string    `json:"event_id" bson:"event_id"`
				EventType   string    `json:"event_type" bson:"event_type"`
				AggregateID string    `json:"aggregate_id" bson:"aggregate_id"`
				Time        time.Time `json:"time" bson:"time"`
				Data        any       `json:"data" bson:"data"`
			}

			if err = json.Unmarshal(msg.Value, &body); err != nil {
				log.Println(err)
				continue
			}

			switch body.EventType {
			case createAccount:
				req, _ := json.Marshal(body.Data)
				err = retry(func() error {
					return c.EventHandler.SaveAccount(context.TODO(), req)
				})
			case saveCredit, saveDebit:
				log.Printf("paso por aca")
				req, _ := json.Marshal(body.Data)
				err = retry(func() error {
					return c.EventHandler.RegisterTransaction(context.TODO(), req)
				})
			default:
				log.Printf("invalid transaction type")
				continue
			}
		}
	}
	defer c.consumer.Close()
}

func retry(operation func() error) error {
	maxRetries := 5
	for retryCount := 1; retryCount <= maxRetries; retryCount++ {
		err := operation()
		if err == nil {
			return nil
		}

		fmt.Printf("Intento %d: Error: %v\n", retryCount, err)

		waitTime := time.Duration(retryCount) * time.Second
		time.Sleep(waitTime)
	}

	return fmt.Errorf("limit reached")
}

func (c eventConsumer) HandlerSummary() {
	topics := "EventQueue"
	err := c.consumer.Subscribe(topics, nil)
	if err != nil {
		fmt.Println(err)
	}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	run := true
	for run == true {
		select {
		case sig := <-signals:
			fmt.Printf("Terminando debido a la señal: %v\n", sig)
			run = false
		default:
			msg, err := c.consumer.ReadMessage(-1)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error al recibir mensaje: %v\n", err)
				continue
			}

			var body struct {
				EventID     string    `json:"event_id" bson:"event_id"`
				EventType   string    `json:"event_type" bson:"event_type"`
				AggregateID string    `json:"aggregate_id" bson:"aggregate_id"`
				Time        time.Time `json:"time" bson:"time"`
				Data        any       `json:"data" bson:"data"`
			}

			if err = json.Unmarshal(msg.Value, &body); err != nil {
				log.Println(err)
				continue
			}

			switch body.EventType {
			case saveCredit, saveDebit:
				req, _ := json.Marshal(body.Data)
				err = retry(func() error {
					return c.EventHandler.RegisterSummary(context.TODO(), req)
				})
			default:
				log.Printf("invalid transaction type")
				continue
			}
		}
	}
	defer c.consumer.Close()
}
