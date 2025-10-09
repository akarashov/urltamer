package main

import (
	"net/http"

	"github.com/akarashov/urltamer/internal/handler"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`GET /`, handler.ResponseEndpoint)
	mux.HandleFunc(`POST /{$}`, handler.RequestEndpoint)
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
