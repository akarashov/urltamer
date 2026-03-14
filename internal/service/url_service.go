// Package service provides URL shortening services.
package service

import (
	"context"
	"errors"

	"github.com/akarashov/urltamer/internal/model"
	"github.com/akarashov/urltamer/internal/repository"
	"github.com/akarashov/urltamer/internal/utils"
)

// ErrURLAlreadyExists is returned when trying to create a short URL for an original URL that already exists.
var ErrURLAlreadyExists = errors.New("URL already exists")

// ErrShortURLConflict is returned when a unique short URL cannot be generated.
var ErrShortURLConflict = errors.New("short URL conflict")

// ErrShortURLNotFound is returned when a short URL does not exist in the repository.
var ErrShortURLNotFound = errors.New("short URL not found")

// ErrURLDeleted is returned when the requested URL has been marked as deleted.
var ErrURLDeleted = errors.New("URL deleted")

// URLService provides methods for URL shortening operations.
type URLService struct {
	repo repository.Repository
}

// NewURLService creates a new instance of URLService.
func NewURLService(repo repository.Repository) *URLService {
	return &URLService{repo: repo}
}

// CreateShortURL creates a new short URL for the given original URL and user ID.
func (s *URLService) CreateShortURL(ctx context.Context, originalURL string, userID int) (*model.Tamer, error) {
	existing, err := s.repo.GetTamerByOriginalURL(ctx, originalURL)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, ErrURLAlreadyExists
	}
	shortURL, err := s.generateUniqueShortURL(ctx)
	if err != nil {
		return nil, err
	}
	tamer := model.Tamer{
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		UserID:      userID,
	}
	// If the underlying repository supports creating users, ensure the user exists
	type userCreator interface {
		CreateUser(ctx context.Context, userID int) (int64, error)
	}
	if uc, ok := s.repo.(userCreator); ok {
		// ignore error (user may already exist)
		_, _ = uc.CreateUser(ctx, userID)
	}
	_, err = s.repo.InsertTamer(ctx, tamer)
	if err != nil {
		return nil, err
	}
	return &tamer, nil
}

// GetOriginalURL retrieves the original URL for a given short URL.
func (s *URLService) GetOriginalURL(ctx context.Context, shortURL string) (string, error) {
	tamer, err := s.repo.GetTamerByShortURL(ctx, shortURL)
	if err != nil {
		return "", err
	}
	if tamer == nil {
		return "", ErrShortURLNotFound
	}
	if tamer.DeletedFlag {
		return "", ErrURLDeleted
	}
	return tamer.OriginalURL, nil
}

// GetAllURLs retrieves all URL mappings from the repository.
func (s *URLService) GetAllURLs(ctx context.Context) (model.Tamers, error) {
	return s.repo.LoadTamers(ctx)
}

// GetUserURLs retrieves all URL mappings for a specific user.
func (s *URLService) GetUserURLs(ctx context.Context, userID int) (model.Tamers, error) {
	return s.repo.GetUserURLs(ctx, userID)
}

// DeleteTamer marks URL mappings as deleted for a specific user.
func (s *URLService) DeleteTamer(ctx context.Context, userID int, shortURLs []string) (int64, error) {
	affected, err := s.repo.DeleteTamer(ctx, userID, shortURLs)
	if err != nil {
		return -1, err
	}
	return affected, nil
}

// Ping checks the connectivity of the repository.
func (s *URLService) Ping(ctx context.Context) bool {
	return s.repo.Ping(ctx)
}

// GetTamerByOriginalURL retrieves a URL mapping by its original URL.
func (s *URLService) GetTamerByOriginalURL(ctx context.Context, originalURL string) (string, error) {
	tamer, err := s.repo.GetTamerByOriginalURL(ctx, originalURL)
	return tamer.ShortURL, err
}

// generateUniqueShortURL generates a unique short URL that does not exist in the repository.
func (s *URLService) generateUniqueShortURL(ctx context.Context) (string, error) {
	for range utils.MaxAttempts {
		shortURL := utils.GenerateShortURL()
		existing, err := s.repo.GetTamerByShortURL(ctx, shortURL)
		if err != nil {
			return "", err
		}
		if existing == nil {
			return shortURL, nil
		}
	}
	return "", ErrShortURLConflict
}

// GetInternalStats retrieves internal statistics of the URL shortener service.
func (s *URLService) GetInternalStats(ctx context.Context) (*model.InternalStats, error) {
	return s.repo.GetInternalStats(ctx)
}
