package main

import (
	"context"
	"flag"
	"net/http"

	"github.com/akarashov/urltamer/internal/config"
	"github.com/akarashov/urltamer/internal/handler"
	"github.com/akarashov/urltamer/internal/repository"
	"github.com/akarashov/urltamer/internal/service"
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

	if ecfg.DataBaseDSN != "" {
		cfg.DataBaseDSN = ecfg.DataBaseDSN
	}

	if ecfg.FileStoragePath != "" {
		cfg.FileStoragePath = ecfg.FileStoragePath
	}
	repo, err := repository.NewRepository(repository.Config{Type: repository.MemoryType,})
	if cfg.DataBaseDSN != "" {
		repo, err = repository.NewRepository(repository.Config{Type: repository.PostgresType, DSN: cfg.DataBaseDSN,})
	} else if cfg.FileStoragePath != "" {
		repo, err = repository.NewRepository(repository.Config{Type: repository.JSONType, Filename: cfg.FileStoragePath,})
	}
	defer repo.Close()
	if err != nil {
		log.Fatal(err)
	}
	service := service.NewURLService(repo)
	ctx := context.Background()
	h := handler.New(cfg, service, ctx)
	mux := chi.NewRouter()
	mux.Post(`/api/shorten`, handler.LoggingMiddlewareRequest(handler.GzipMiddleware(h.RequestJSONEndpoint), *log))
	mux.Get(`/{tamer}`, handler.LoggingMiddlewareResponse(handler.GzipMiddleware(h.ResponseEndpoint), *log))
	mux.Get(`/ping`, handler.LoggingMiddlewareResponse(handler.GzipMiddleware(h.PingEndpoint), *log))
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
