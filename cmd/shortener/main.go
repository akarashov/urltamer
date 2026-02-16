// URL Shortener Service (tamer)
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/akarashov/urltamer/internal/config"
	"github.com/akarashov/urltamer/internal/handler"
	"github.com/akarashov/urltamer/internal/repository"
	"github.com/akarashov/urltamer/internal/service"
	"github.com/akarashov/urltamer/internal/utils"
	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/acme/autocert"
)

var buildVersion string
var buildDate string
var buildCommit string

func main() {

	fmt.Printf("Build version: %s\n", utils.NA(buildVersion))
	fmt.Printf("Build date: %s\n", utils.NA(buildDate))
	fmt.Printf("Build commit: %s\n", utils.NA(buildCommit))

	log := config.NewLogger()
	cfg := config.New()
	flag.Parse()
	ecfg := config.NewEnv()

	config.ApplyEnvOverrides(cfg, ecfg)

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

	// middleware that need extra params
	gzipMw := handler.Adapt(handler.GzipMiddleware)
	cookieMw := handler.Adapt(handler.CookieMiddleware)
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

	// start pprof server with graceful shutdown support
	pprofSrv := &http.Server{Addr: config.PprofAddr, Handler: nil}
	go func() {
		log.Infow("Starting pprof", "addr", config.PprofAddr)
		if perr := pprofSrv.ListenAndServe(); perr != nil && perr != http.ErrServerClosed {
			log.Warn(perr)
		}
	}()

	// start main server with graceful shutdown support
	srv := &http.Server{Addr: cfg.Listen, Handler: mux}
	go func() {
		if cfg.EnableHTTPS {
			manager := &autocert.Manager{
				Cache:  autocert.DirCache("certs"),
				Prompt: autocert.AcceptTOS,
			}
			srv.TLSConfig = manager.TLSConfig()
			log.Infow("Starting https server", "addr", cfg.Listen)
			if serr := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				log.Error(serr)
			}
		} else {
			log.Infow("Starting http server", "addr", cfg.Listen)
			if serr := srv.ListenAndServe(); serr != nil && serr != http.ErrServerClosed {
				log.Error(serr)
			}
		}
	}()

	// wait for interrupt signal to gracefully shutdown servers
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Infow("Shutting down servers")
	ctxShut, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = srv.Shutdown(ctxShut); err != nil {
		log.Warnw("Error shutting down main server", "err", err)
	}
	if err = pprofSrv.Shutdown(ctxShut); err != nil {
		log.Warnw("Error shutting down pprof server", "err", err)
	}
}
