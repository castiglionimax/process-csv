package server

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"log"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	_ "github.com/go-sql-driver/mysql"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/castiglionimax/process-csv/internal/controller"
	"github.com/castiglionimax/process-csv/internal/repository"
	"github.com/castiglionimax/process-csv/internal/service"
)

func resolveController() controller.Controller {
	ctr, _ := controller.NewController(resolverService())
	return *ctr
}

func resolverService() *service.Service {
	srv, _ := service.NewService(
		repository.NewRepository(resolverQueueProducer(),
			"EventQueue",
			resolveEventStore(),
			resolverRelationDatabase(),
			resolverObjectStorage()))
	return srv
}

func resolverQueueProducer() *kafka.Producer {
	uri := os.Getenv("KAFKA_URI")
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": uri,
		"client.id":         "foo",
		"acks":              "all",
		"auto.offset.reset": "smallest"})
	if err != nil {
		panic(err)
	}
	return producer
}

func resolverQueueConsumer(groupId string) *kafka.Consumer {
	uri := os.Getenv("KAFKA_URI")
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  uri,
		"group.id":           groupId,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": "false",
	})

	if err != nil {
		panic(err)
	}

	return consumer
}

func resolveEventStore() *mongo.Client {
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	uri := os.Getenv("MONGO_URI")
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	err = mongoClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatalf("ping mongodb error :%v", err)
		return nil
	}
	fmt.Println("ping success")
	return mongoClient
}

func resolverEventService() *service.EventService {
	return service.NewEventService(repository.NewProjection(resolverRelationDatabase()))
}

func resolverRelationDatabase() *sql.DB {
	uri := os.Getenv("MYSQL_URI")
	db, err := sql.Open("mysql", uri)
	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	return db
}

func resolverObjectStorage() *minio.Client {
	uri := os.Getenv("MINIO_ENDPOINT")
	usr := os.Getenv("MINIO_ROOT_USER")
	pass := os.Getenv("MINIO_ROOT_PASSWORD")

	minioClient, err := minio.New(uri, &minio.Options{
		Creds:  credentials.NewStaticV4(usr, pass, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalln(err)
	}

	bucketName := "transactions"
	location := "us-east-1"
	ctx := context.TODO()
	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Printf("We already own %s\n", bucketName)
		} else {
			log.Fatalln(err)
		}
	} else {
		log.Printf("Successfully created %s\n", bucketName)
	}
	return minioClient
}
