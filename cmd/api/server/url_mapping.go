package server

import (
	"github.com/castiglionimax/process-csv/internal/controller"
	"net/http"

	"github.com/go-chi/chi"
)

type mapping struct {
	controller controller.Controller
}

func newMapping() *mapping {
	return &mapping{
		controller: resolveController(),
	}
}

func (m mapping) mapUrlsToControllers(route *chi.Mux) {
	route.Get("/ping", alive())

	route.Post("/accounts", m.controller.CreateAccount)

	route.Post("/csv/upload", m.controller.UploadHandler)

	route.Post("/csv", m.controller.CreateCsv)

	route.Post("/csv/process", m.controller.ProcessFiles)

}

func alive() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("pong"))
	}
}
