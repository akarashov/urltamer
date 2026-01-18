package repository

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/akarashov/urltamer/internal/model"
)

// реализация наблюдателя для записи в файл
type AuditFileObserver struct {
	ID   string
	File string
}

func NewAuditFileObserver(file string) *AuditFileObserver {
	return &AuditFileObserver{ID: "audit_file", File: file}
}

func (a *AuditFileObserver) Update(event model.AuditEvent) {
	data, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		log.Printf("Audit file marshalling error: %v", err)
		return
	}
	data = append(data, '\n')
	f, err := os.OpenFile(a.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Audit file open file error: %v", err)
		return
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		log.Printf("Audit file write error: %v", err)
	}
}

func (a *AuditFileObserver) GetID() string {
	return a.ID
}

// реализация наблюдателя для записи в URL
type AuditURLObserver struct {
	ID  string
	URL string
}

func NewAuditURLObserver(url string) *AuditURLObserver {
	return &AuditURLObserver{ID: "audit_url", URL: url}
}

func (a *AuditURLObserver) Update(event model.AuditEvent) {

	data, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		log.Printf("Audit file marshalling error: %v", err)
		return
	}
	body := io.Reader(bytes.NewBuffer(data))
	http.Post(a.URL, "Content-Type: application/json", body)
}

func (a *AuditURLObserver) GetID() string {
	return a.ID
}
