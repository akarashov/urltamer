package repository

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"sync"

	"github.com/akarashov/urltamer/internal/model"
)

// JSONFileRepository is a repository that stores URL mappings in a JSON file.
type JSONFileRepository struct {
	mutex    sync.RWMutex
	filename string
	data     map[int]model.Tamer // ID:Tamer
	URL      map[string]int      // URL:ID
	ID       int
}

// NewJSONFileRepository creates a new JSONFileRepository.
func NewJSONFileRepository(filename string) (*JSONFileRepository, error) {
	repo := &JSONFileRepository{
		filename: filename,
		data:     make(map[int]model.Tamer),
		URL:      make(map[string]int),
		ID:       1,
	}
	err := repo.open()
	if err != nil {
		return nil, err
	}
	return repo, nil
}

// LoadTamers loads all URL mappings from the repository.
func (j *JSONFileRepository) LoadTamers(ctx context.Context) (model.Tamers, error) {
	tamers := make(model.Tamers, 0, len(j.data))
	for _, tamer := range j.data {
		tamers = append(tamers, tamer)
	}
	return tamers, nil
}

// InsertTamer inserts a new URL mapping into the repository.
func (j *JSONFileRepository) InsertTamer(ctx context.Context, tamer model.Tamer) (int64, error) {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	id, exists := j.URL[tamer.OriginalURL]
	if exists { // Update
		existing := j.data[id]
		existing.ShortURL = tamer.ShortURL
		j.data[id] = existing
	} else {
		tamer.ID = strconv.Itoa(j.ID) // New
		j.data[j.ID] = tamer
		j.URL[tamer.OriginalURL] = j.ID
		j.ID++
	}
	err := j.save() // Save to file
	if err != nil {
		return 0, err
	}
	return 1, nil
}

// GetTamerByShortURL retrieves a URL mapping by its short URL.
func (j *JSONFileRepository) GetTamerByShortURL(ctx context.Context, shortURL string) (*model.Tamer, error) {
	j.mutex.RLock()
	defer j.mutex.RUnlock()
	for _, tamer := range j.data {
		if tamer.ShortURL == shortURL {
			return &tamer, nil
		}
	}
	return nil, nil
}

// GetTamerByOriginalURL retrieves a URL mapping by its original URL.
func (j *JSONFileRepository) GetTamerByOriginalURL(ctx context.Context, originalURL string) (*model.Tamer, error) {
	j.mutex.RLock()
	defer j.mutex.RUnlock()
	id, exists := j.URL[originalURL]
	if exists {
		tamer := j.data[id]
		return &tamer, nil
	}
	return nil, nil
}

// GetUserURLs retrieves all URL mappings for a specific user.
func (j *JSONFileRepository) GetUserURLs(ctx context.Context, userID int) (model.Tamers, error) {
	j.mutex.RLock()
	defer j.mutex.RUnlock()
	var tamers model.Tamers
	for _, tamer := range j.data {
		if tamer.UserID == userID {
			tamers = append(tamers, tamer)
		}
	}
	return tamers, nil
}

// Ping checks the connectivity of the repository.
func (j *JSONFileRepository) Ping(ctx context.Context) bool {
	return false
}

// Close saves the data and closes the repository.
func (j *JSONFileRepository) Close() error {
	return j.save()
}

// DeleteTamer marks URL mappings as deleted for a specific user.
func (j *JSONFileRepository) DeleteTamer(ctx context.Context, userID int, shortURLs []string) (int64, error) {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	set := make(map[string]struct{}, len(shortURLs))
	for _, s := range shortURLs {
		set[s] = struct{}{}
	}
	var affected int64
	for id, tamer := range j.data {
		if tamer.UserID != userID {
			continue
		}
		if _, ok := set[tamer.ShortURL]; ok {
			if !tamer.DeletedFlag {
				tamer.DeletedFlag = true
				j.data[id] = tamer
				affected++
			}
		}
	}
	if affected > 0 {
		if err := j.save(); err != nil {
			return 0, err
		}
	}
	return affected, nil
}

// open loads the data from the JSON file into memory.
func (j *JSONFileRepository) open() error {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	file, err := os.ReadFile(j.filename)
	if err != nil {
		return nil
	} else {
		tamers := model.Tamers{}
		err = json.Unmarshal(file, &tamers)
		if err != nil {
			return err
		}
		for _, t := range tamers {
			id, err := strconv.Atoi(t.ID)
			if err != nil {
				return err
			}
			j.data[id] = t
			j.URL[t.OriginalURL] = id
			j.ID = id + 1
		}
		return nil
	}
}

// save writes the in-memory data to the JSON file.
func (j *JSONFileRepository) save() error {

	tamers, err := j.LoadTamers(context.TODO())
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(tamers, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(j.filename, data, 0666)
}
