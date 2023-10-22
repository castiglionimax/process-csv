package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

const defaultPort = ":8080"

func StartApplication() {
	route := chi.NewRouter()
	route.Use(middleware.Timeout(60 * time.Second))
	mapping := newMapping()
	mapping.mapUrlsToControllers(route)
	serverport := os.Getenv("PORT")

	go newConsumerEvent(resolverQueueConsumer("account")).HandlerAccount()
	go newConsumerEvent(resolverQueueConsumer("summary")).HandlerSummary()

	if serverport == "" {
		serverport = defaultPort
	} else {
		serverport = fmt.Sprintf(":%s", serverport)
	}

	log.Default().Printf("PORT: %s", serverport)

	err := http.ListenAndServe(serverport, route)
	if err != nil {
		panic(err)
	}
}
