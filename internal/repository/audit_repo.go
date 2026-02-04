package repository

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/akarashov/urltamer/internal/model"
)

// AuditFileObserver creates an observer that logs audit events to a file.
type AuditFileObserver struct {
	ID   string
	File string
}

// NewAuditFileObserver creates a new AuditFileObserver.
func NewAuditFileObserver(file string) *AuditFileObserver {
	return &AuditFileObserver{ID: "audit_file", File: file}
}

// Update writes the audit event to the specified file.
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

// GetID returns the ID of the observer.
func (a *AuditFileObserver) GetID() string {
	return a.ID
}

// AuditURLObserver creates an observer that sends audit events to a specified URL.
type AuditURLObserver struct {
	ID  string
	URL string
}

// NewAuditURLObserver creates a new AuditURLObserver.
func NewAuditURLObserver(url string) *AuditURLObserver {
	return &AuditURLObserver{ID: "audit_url", URL: url}
}

// Update sends the audit event to the specified URL via HTTP POST.
func (a *AuditURLObserver) Update(event model.AuditEvent) {
	data, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		log.Printf("Audit URL marshalling error: %v", err)
		return
	}
	resp, err := http.Post(a.URL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Printf("Audit URL POST error: %v", err)
		return
	}
	defer resp.Body.Close()
}

// GetID returns the ID of the observer.
func (a *AuditURLObserver) GetID() string {
	return a.ID
}
