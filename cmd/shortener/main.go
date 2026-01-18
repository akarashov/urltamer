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

	if ecfg.AuditFile != "" {
		cfg.AuditFile = ecfg.AuditFile
	}

	if ecfg.AuditURL != "" {
		cfg.AuditURL = ecfg.AuditURL
	}

	auditSubject := service.NewAuditSubject()
	if cfg.AuditFile != "" {
		auditSubject.Register(repository.NewAuditFileObserver(cfg.AuditFile))
	}
	if cfg.AuditURL != "" {
		auditSubject.Register(repository.NewAuditURLObserver(cfg.AuditURL))
	}

	repo, err := repository.NewRepository(repository.Config{Type: repository.MemoryType})
	if cfg.DataBaseDSN != "" {
		repo, err = repository.NewRepository(repository.Config{Type: repository.PostgresType, DSN: cfg.DataBaseDSN})
	} else if cfg.FileStoragePath != "" {
		repo, err = repository.NewRepository(repository.Config{Type: repository.JSONType, Filename: cfg.FileStoragePath})
	}
	defer repo.Close()
	if err != nil {
		log.Errorf("EEEEEEE %s\n", err)
	}
	service := service.NewURLService(repo)
	ctx := context.Background()
	h := handler.New(cfg, service, ctx)
	mux := chi.NewRouter()

	mux.Post(`/`, handler.AuditMiddleware(handler.LoggingMiddlewareRequest(handler.GzipMiddleware(handler.CookieMiddleware(h.RequestEndpoint)), *log), auditSubject)) // TODO: make middleware audit
	mux.Post(`/api/shorten`, handler.AuditMiddleware(handler.LoggingMiddlewareRequest(handler.GzipMiddleware(h.RequestJSONEndpoint), *log), auditSubject))            // TODO: make middleware audit
	mux.Post(`/api/shorten/batch`, handler.LoggingMiddlewareRequest(handler.GzipMiddleware(h.RequestJSONEndpointBatch), *log))

	mux.Get(`/api/user/urls`, handler.LoggingMiddlewareResponse(handler.GzipMiddleware(handler.CookieMiddleware(h.UserURLsEndpoint)), *log))
	mux.Get(`/ping`, handler.LoggingMiddlewareResponse(handler.GzipMiddleware(h.PingEndpoint), *log))
	mux.Get(`/{tamer}`, handler.AuditMiddleware(handler.LoggingMiddlewareResponse(handler.GzipMiddleware(handler.CookieMiddleware(h.ResponseEndpoint)), *log), auditSubject)) // TODO: make middleware audit

	mux.Delete(`/api/user/urls`, handler.LoggingMiddlewareRequest(handler.GzipMiddleware(handler.CookieMiddleware(h.DeleteUserURLsEndpoint)), *log))

	log.Infow("Starting server", "addr", cfg.Listen)

	err = http.ListenAndServe(cfg.Listen, mux)
	if err != nil {
		log.Fatal(err)
	}
}
