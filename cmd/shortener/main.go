package main

import (
	"flag"
	"net/http"

	"github.com/akarashov/urltamer/internal/config"
	"github.com/akarashov/urltamer/internal/handler"
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
	
	h := handler.New(cfg)
	mux := chi.NewRouter()
	
	mux.Post(`/api/shorten`, handler.LoggingMiddlewareRequest(handler.GzipMiddleware(h.RequestJSONEndpoint), *log))
	mux.Get(`/{tamer}`, handler.LoggingMiddlewareResponse(handler.GzipMiddleware(h.ResponseEndpoint), *log))
	mux.Get(`/ping`, handler.LoggingMiddlewareResponse(handler.GzipMiddleware(h.PingEndpoint), *log))
	mux.Post(`/`, handler.LoggingMiddlewareRequest(handler.GzipMiddleware(h.RequestEndpoint), *log))
	log.Infow(
		"Starting server",
		"addr", cfg.Listen,
	)
	err := http.ListenAndServe(cfg.Listen, mux)
	if err != nil {
		log.Fatal(err)
	}
}
