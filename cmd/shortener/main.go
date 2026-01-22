package main

import (
	"context"
	"flag"
	"net/http"
	_ "net/http/pprof"

	"github.com/akarashov/urltamer/internal/config"
	"github.com/akarashov/urltamer/internal/handler"
	"github.com/akarashov/urltamer/internal/repository"
	"github.com/akarashov/urltamer/internal/service"
	"github.com/go-chi/chi/v5"
)

const (
	pprofAddr = ":9090"
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

	// Adapt existing handler-style middleware (func(http.HandlerFunc) http.HandlerFunc)
	adapt := func(mw func(http.HandlerFunc) http.HandlerFunc) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mw(http.HandlerFunc(next.ServeHTTP))(w, r)
			})
		}
	}

	// middleware that need extra params
	gzipMw := adapt(handler.GzipMiddleware)
	cookieMw := adapt(handler.CookieMiddleware)
	logReqMw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.LoggingMiddlewareRequest(http.HandlerFunc(next.ServeHTTP), *log)(w, r)
		})
	}
	logResMw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.LoggingMiddlewareResponse(http.HandlerFunc(next.ServeHTTP), *log)(w, r)
		})
	}

	// Common middleware
	mux.Use(gzipMw)
	mux.Use(cookieMw)
	mux.Use(logResMw)
	mux.Use(logReqMw)

	// Routes
	mux.Post(`/`, handler.AuditMiddleware(h.RequestEndpoint, auditSubject))
	mux.Post(`/api/shorten`, handler.AuditMiddleware(h.RequestJSONEndpoint, auditSubject))
	mux.Post(`/api/shorten/batch`, h.RequestJSONEndpointBatch)

	mux.Get(`/api/user/urls`, h.UserURLsEndpoint)
	mux.Get(`/ping`, h.PingEndpoint)
	mux.Get(`/{tamer}`, handler.AuditMiddleware(h.ResponseEndpoint, auditSubject))

	mux.Delete(`/api/user/urls`, h.DeleteUserURLsEndpoint)

	go func() {
		log.Infow("Starting pprof", "addr", pprofAddr)
		perr := http.ListenAndServe(pprofAddr, nil)
		if perr != nil {
			log.Warn(perr)
		}
	}()

	log.Infow("Starting server", "addr", cfg.Listen)
	err = http.ListenAndServe(cfg.Listen, mux)
	if err != nil {
		log.Fatal(err)
	}
}
