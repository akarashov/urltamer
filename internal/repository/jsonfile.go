package repository

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"sync"

	"github.com/akarashov/urltamer/internal/model"
)

type JSONFileRepository struct {
	mutex    sync.RWMutex
	filename string
	data     map[int]model.Tamer // ID:Tamer
	URL      map[string]int      // URL:ID
	ID       int
}

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

func (j *JSONFileRepository) LoadTamers(ctx context.Context) (model.Tamers, error) {
	tamers := make(model.Tamers, 0, len(j.data))
	for _, tamer := range j.data {
		tamers = append(tamers, tamer)
	}
	return tamers, nil
}

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

func (j *JSONFileRepository) Ping(ctx context.Context) bool {
	return false
}

func (j *JSONFileRepository) Close() error {
	return j.save()
}
