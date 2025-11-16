package service

import (
	"context"
	"errors"

	"github.com/akarashov/urltamer/internal/model"
	"github.com/akarashov/urltamer/internal/repository"
	"github.com/akarashov/urltamer/internal/utils"
)

type URLService struct {
	repo repository.Repository
}

func NewURLService(repo repository.Repository) *URLService {
	return &URLService{repo: repo}
}

func (s *URLService) CreateShortURL(ctx context.Context, originalURL string) (*model.Tamer, error) {
	existing, err := s.repo.GetTamerByOriginalURL(ctx, originalURL)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, errors.New("URL already exists")
	}
	shortURL, err := s.generateUniqueShortURL(ctx)
	if err != nil {
		return nil, err
	}
	tamer := model.Tamer{
		ShortURL:    shortURL,
		OriginalURL: originalURL,
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
	return "", errors.New("short URL conflict")
}

func (s *URLService) GetOriginalURL(ctx context.Context, shortURL string) (string, error) {
	tamer, err := s.repo.GetTamerByShortURL(ctx, shortURL)
	if err != nil {
		return "", err
	}
	if tamer == nil {
		return "", errors.New("short URL not found")
	}
	return tamer.OriginalURL, nil
}

func (s *URLService) GetAllURLs(ctx context.Context) (model.Tamers, error) {
	return s.repo.LoadTamers(ctx)
}

func (s *URLService) Ping(ctx context.Context) bool {
	return s.repo.Ping(ctx)
}
