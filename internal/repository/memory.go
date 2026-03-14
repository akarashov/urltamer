package repository

import (
	"context"
	"strconv"
	"sync"

	"github.com/akarashov/urltamer/internal/model"
)

// MemoryRepository is an in-memory implementation of the Repository interface.
type MemoryRepository struct {
	mutex sync.RWMutex
	data  map[int]model.Tamer
	URL   map[string]int // URL:ID
	ID    int
}

// NewMemoryRepository creates a new MemoryRepository.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		data: make(map[int]model.Tamer),
		URL:  make(map[string]int),
		ID:   1,
	}
}

// LoadTamers loads all URL mappings from the repository.
func (m *MemoryRepository) LoadTamers(ctx context.Context) (model.Tamers, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	tamers := make(model.Tamers, 0, len(m.data))
	for _, tamer := range m.data {
		tamers = append(tamers, tamer)
	}
	return tamers, nil
}

// InsertTamer inserts a new URL mapping into the repository.
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

// GetTamerByShortURL retrieves a URL mapping by its short URL.
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

// GetTamerByOriginalURL retrieves a URL mapping by its original URL.
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

// GetUserURLs retrieves all URL mappings for a specific user.
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

// Ping checks the health of the repository.
func (m *MemoryRepository) Ping(ctx context.Context) bool {
	return false
}

// Close closes the repository.
func (m *MemoryRepository) Close() error {
	return nil
}

// DeleteTamer marks URL mappings as deleted for a specific user.
func (m *MemoryRepository) DeleteTamer(ctx context.Context, userID int, shortURLs []string) (int64, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	set := make(map[string]struct{}, len(shortURLs))
	for _, s := range shortURLs {
		set[s] = struct{}{}
	}
	var affected int64
	for id, tamer := range m.data {
		if tamer.UserID != userID {
			continue
		}
		if _, ok := set[tamer.ShortURL]; ok {
			if !tamer.DeletedFlag {
				tamer.DeletedFlag = true
				m.data[id] = tamer
				affected++
			}
		}
	}
	return affected, nil
}

// GetInternalStats retrieves internal statistics of the URL shortener service.
func (m *MemoryRepository) GetInternalStats(ctx context.Context) (*model.InternalStats, error) {
	var internalStats model.InternalStats
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	users := make(map[int]struct{})
	for _, d := range m.data {
		users[d.UserID] = struct{}{}
	}
	internalStats.Users = len(users)
	internalStats.URLs = len(m.data)
	return &internalStats, nil
}
