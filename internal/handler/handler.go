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
	Handler struct {
		Service *service.URLService
		Context context.Context
		Base    *string
		Tamers  model.Tamers
		// tamerCh chan model.DeleteTamer
	}
)

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
		// tamerCh: make(chan model.DeleteTamer, 32),
	}
	// return handler
}
