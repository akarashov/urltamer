// Package hadler of the application URL Tamer serve http requests.
package handler

import (
	"context"
	"log"
	"strings"

	"github.com/akarashov/urltamer/internal/config"
	"github.com/akarashov/urltamer/internal/model"
	"github.com/akarashov/urltamer/internal/service"
)

type (
	// Handler is a structure that handles HTTP requests for URL shortening service.
	Handler struct {
		Service *service.URLService
		Context context.Context
		Base    *string
		Tamers  model.Tamers
	}
)

// New creates a new Handler instance.
func New(c *config.Config, s *service.URLService, ctx context.Context) *Handler {
	c.Base = strings.TrimRight(c.Base, "/")
	mc, err := s.GetAllURLs(ctx)
	if err != nil {
		log.Printf("Fatality %s", err)
	}
	return &Handler{
		Service: s,
		Context: ctx,
		Base:    &c.Base,
		Tamers:  mc,
	}
}
