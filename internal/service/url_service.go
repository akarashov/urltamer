package service

import (
	"context"
	"errors"

	"github.com/akarashov/urltamer/internal/model"
	"github.com/akarashov/urltamer/internal/repository"
	"github.com/akarashov/urltamer/internal/utils"
)

var ErrURLAlreadyExists = errors.New("URL already exists")
var ErrShortURLConflict = errors.New("short URL conflict")
var ErrShortURLNotFound = errors.New("short URL not found")

type URLService struct {
	repo repository.Repository
}

func NewURLService(repo repository.Repository) *URLService {
	return &URLService{repo: repo}
}

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

func (s *URLService) GetOriginalURL(ctx context.Context, shortURL string) (string, error) {
	tamer, err := s.repo.GetTamerByShortURL(ctx, shortURL)
	if err != nil {
		return "", err
	}
	if tamer == nil {
		return "", ErrShortURLNotFound
	}
	return tamer.OriginalURL, nil
}

func (s *URLService) GetAllURLs(ctx context.Context) (model.Tamers, error) {
	return s.repo.LoadTamers(ctx)
}

func (s *URLService) GetUserURLs(ctx context.Context, userID int) (model.Tamers, error) {
	return s.repo.GetUserURLs(ctx, userID)
}

func (s *URLService) Ping(ctx context.Context) bool {
	return s.repo.Ping(ctx)
}

func (s *URLService) GetTamerByOriginalURL(ctx context.Context, originalURL string) (string, error) {
	tamer, err := s.repo.GetTamerByOriginalURL(ctx, originalURL)
	return tamer.ShortURL, err
}
