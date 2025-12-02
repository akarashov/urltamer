package model

type Tamer struct {
	ID          string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      int    `json:"user_id"`
	DeletedFlag bool   `json:"is_deleted"`
}

type Tamers []Tamer

type RequestBatch struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type RequestBatchs []RequestBatch

type ResponseBatch struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type ResponseBatchs []ResponseBatch

type UserURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type UserURLs []UserURL

type DeleteTamer struct {
	UserID      int    `json:"user_id"`
	ShortURL    string `json:"short_url"`

}
