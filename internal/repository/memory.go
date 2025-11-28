package repository

import (
	"context"
	"strconv"
	"sync"

	"github.com/akarashov/urltamer/internal/model"
)

type MemoryRepository struct {
	mutex sync.RWMutex
	data  map[int]model.Tamer
	URL   map[string]int // URL:ID
	ID    int
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		data: make(map[int]model.Tamer),
		URL:  make(map[string]int),
		ID:   1,
	}
}

func (m *MemoryRepository) LoadTamers(ctx context.Context) (model.Tamers, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	tamers := make(model.Tamers, 0, len(m.data))
	for _, tamer := range m.data {
		tamers = append(tamers, tamer)
	}
	return tamers, nil
}

func (m *MemoryRepository) InsertTamer(ctx context.Context, tamer model.Tamer) (int64, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	id, exists := m.URL[tamer.OriginalURL]
	if exists { // Update
		existing := m.data[id]
		existing.ShortURL = tamer.ShortURL
		m.data[id] = existing
		return 1, nil
	}
	tamer.ID = strconv.Itoa(m.ID) // New
	m.data[m.ID] = tamer
	m.URL[tamer.OriginalURL] = m.ID
	m.ID++
	return 1, nil
}

func (m *MemoryRepository) GetTamerByShortURL(ctx context.Context, shortURL string) (*model.Tamer, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	for _, tamer := range m.data {
		if tamer.ShortURL == shortURL {
			return &tamer, nil
		}
	}
	return nil, nil
}

func (m *MemoryRepository) GetTamerByOriginalURL(ctx context.Context, originalURL string) (*model.Tamer, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	id, exists := m.URL[originalURL]
	if exists {
		tamer := m.data[id]
		return &tamer, nil
	}
	return nil, nil
}

func (m *MemoryRepository) GetUserURLs(ctx context.Context, userID int) (model.Tamers, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	var tamers model.Tamers
	for _, tamer := range m.data {
		if tamer.UserID == userID {
			tamers = append(tamers, tamer)
		}
	}
	return tamers, nil
}

func (m *MemoryRepository) Ping(ctx context.Context) bool {
	return false
}

func (m *MemoryRepository) Close() error {
	return nil
}
