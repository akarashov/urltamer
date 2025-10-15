package main

import (
	"net/http"

	"github.com/akarashov/urltamer/internal/config"
	"github.com/akarashov/urltamer/internal/handler"
	"github.com/go-chi/chi/v5"
)

func main() {
	var Cfg config.Config
	Cfg.New()
	mux := chi.NewRouter()
	newhandler := handler.New(&Cfg)
	mux.Get(`/{tamer}`, handler.ResponseEndpoint)
	mux.Post(`/`, newhandler.RequestEndpoint)
	err := http.ListenAndServe(Cfg.Listen, mux)
	if err != nil {
		panic(err)
	}
}
