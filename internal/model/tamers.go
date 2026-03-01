// Package model defines the data models used in the URL shortener service.
package model

// Tamer represents a shortened URL entry.
type Tamer struct {
	ID          string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      int    `json:"user_id"`
	DeletedFlag bool   `json:"is_deleted"`
}

// Tamers is a slice of Tamer.
type Tamers []Tamer

// RequestBatch represents a single URL shortening request in a batch.
type RequestBatch struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// RequestBatchs is a slice of RequestBatch.
type RequestBatchs []RequestBatch

// ResponseBatch represents a single URL shortening response in a batch.
type ResponseBatch struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// ResponseBatchs is a slice of ResponseBatch.
type ResponseBatchs []ResponseBatch

// UserURL represents a shortened URL and its original URL for a user.
type UserURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// UserURLs is a slice of UserURL.
type UserURLs []UserURL

// DeleteTamer represents a request to delete a shortened URL.
type DeleteTamer struct {
	UserID   int    `json:"user_id"`
	ShortURL string `json:"short_url"`
}

// AuditEvent represents an audit log event for URL shortening actions.
type AuditEvent struct {
	TS     int    `json:"ts"`                // unix timestamp события
	Action string `json:"action"`            // действие: shorten (создание) или follow (прохождение по ссылке)
	UserID string `json:"user_id,omitempty"` // идентификатор пользователя, если есть
	URL    string `json:"url,omitempty"`     // оригинальный (не сокращенный) URL
}

// InternalStats represents internal statistics of the URL shortener service.
type InternalStats struct {
	URLs  int `json:"urls,omitempty"`  // количество сокращённых URL в сервисе
	Users int `json:"users,omitempty"` // количество пользователей в сервисе
}
