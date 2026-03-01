// Package repository defines the interface for URL storage repositories.
package repository

import (
	"context"

	"github.com/akarashov/urltamer/internal/model"
)

// Repository defines the methods for interacting with the URL storage.
type Repository interface {
	LoadTamers(ctx context.Context) (model.Tamers, error)
	GetUserURLs(ctx context.Context, userID int) (model.Tamers, error)
	InsertTamer(ctx context.Context, tamer model.Tamer) (int64, error)
	GetTamerByShortURL(ctx context.Context, shortURL string) (*model.Tamer, error)
	GetTamerByOriginalURL(ctx context.Context, originalURL string) (*model.Tamer, error)
	DeleteTamer(ctx context.Context, userID int, shortURLs []string) (int64, error)
	Ping(ctx context.Context) bool
	GetInternalStats(ctx context.Context) (*model.InternalStats, error)
	Close() error
}
