package main

import (
	"flag"
	"net/http"
	"log"

	"github.com/akarashov/urltamer/internal/config"
	"github.com/akarashov/urltamer/internal/handler"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.New()
	flag.Parse()
	mux := chi.NewRouter()
	h := handler.New(cfg)
	mux.Get(`/{tamer}`, h.ResponseEndpoint)
	mux.Post(`/`, h.RequestEndpoint)
	err := http.ListenAndServe(cfg.Listen, mux)
	if err != nil {
		log.Fatal(err)
	}
}
