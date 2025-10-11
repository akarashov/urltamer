package main

import (
	"net/http"

	"github.com/akarashov/urltamer/internal/handler"
	"github.com/go-chi/chi/v5"
)

func main() {

	mux := chi.NewRouter()
	mux.Get(`/{tamer}`, handler.ResponseEndpoint)
	mux.Post(`/`, handler.RequestEndpoint)
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
