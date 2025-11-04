package main

import (
	"flag"
	"net/http"

	"github.com/akarashov/urltamer/internal/config"
	"github.com/akarashov/urltamer/internal/handler"
	"github.com/akarashov/urltamer/internal/model"
	"github.com/go-chi/chi/v5"
)

func main() {
	log := config.NewLogger()
	cfg := config.New()
	flag.Parse()
	ecfg := config.NewEnv()
	if ecfg.Listen != "" {
		cfg.Listen = ecfg.Listen
	}
	if ecfg.Base != "" {
		cfg.Base = ecfg.Base
	}
	if ecfg.FileStoragePath != "" {
		cfg.FileStoragePath = ecfg.FileStoragePath
	}

	handler.Tamers = model.Tamers{}
	err := handler.Tamers.Load(cfg.FileStoragePath)
	if err != nil {
		handler.TamerCounter = 0
	}
	handler.TamerCounter = len(handler.Tamers) 
	mux := chi.NewRouter()
	h := handler.New(cfg)
	mux.Post(`/api/shorten`, handler.LoggingMiddlewareRequest(handler.GzipMiddleware(h.RequestJSONEndpoint), *log))
	mux.Get(`/{tamer}`, handler.LoggingMiddlewareResponse(handler.GzipMiddleware(h.ResponseEndpoint), *log))
	mux.Post(`/`, handler.LoggingMiddlewareRequest(handler.GzipMiddleware(h.RequestEndpoint), *log))
	log.Infow(
		"Starting server",
		"addr", cfg.Listen,
	)
	err = http.ListenAndServe(cfg.Listen, mux)
	if err != nil {
		log.Fatal(err)
	}
}
