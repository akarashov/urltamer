package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/akarashov/urltamer/internal/config"
	"github.com/akarashov/urltamer/internal/repository"
	"github.com/akarashov/urltamer/internal/service"
)

// Example of plain text shorten endpoint.
func ExampleHandler_RequestEndpoint() {
	ctx := context.Background()
	repo, _ := repository.NewRepository(repository.Config{Type: repository.MemoryType})
	svc := service.NewURLService(repo)
	cfg := &config.Config{Base: "http://example.com"}
	h := New(cfg, svc, ctx)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("http://original.example/1"))
	rr := httptest.NewRecorder()
	h.RequestEndpoint(rr, req)

	fmt.Println(rr.Code)
}

// Example of JSON shorten endpoint.
func ExampleHandler_RequestJSONEndpoint() {
	ctx := context.Background()
	repo, _ := repository.NewRepository(repository.Config{Type: repository.MemoryType})
	svc := service.NewURLService(repo)
	cfg := &config.Config{Base: "http://example.com"}
	h := New(cfg, svc, ctx)

	body := `{"url":"http://original.example/2"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.RequestJSONEndpoint(rr, req)

	fmt.Println(rr.Code)
}

// Example of JSON batch shorten endpoint.
func ExampleHandler_RequestJSONEndpointBatch() {
	ctx := context.Background()
	repo, _ := repository.NewRepository(repository.Config{Type: repository.MemoryType})
	svc := service.NewURLService(repo)
	cfg := &config.Config{Base: "http://example.com"}
	h := New(cfg, svc, ctx)

	batch := []map[string]string{{"correlation_id": "1", "original_url": "http://original.example/3"}}
	b, _ := json.Marshal(batch)
	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.RequestJSONEndpointBatch(rr, req)

	fmt.Println(rr.Code)
}

// Example of redirect (response) endpoint.
func ExampleHandler_ResponseEndpoint() {
	ctx := context.Background()
	repo, _ := repository.NewRepository(repository.Config{Type: repository.MemoryType})
	svc := service.NewURLService(repo)
	cfg := &config.Config{Base: "http://example.com"}
	h := New(cfg, svc, ctx)

	tamer, _ := svc.CreateShortURL(ctx, "http://original.example/4", 7)

	req := httptest.NewRequest(http.MethodGet, "/"+tamer.ShortURL, nil)
	rr := httptest.NewRecorder()
	h.ResponseEndpoint(rr, req)

	fmt.Println(rr.Code)
}

// Example of user URLs retrieval endpoint.
func ExampleHandler_UserURLsEndpoint() {
	ctx := context.Background()
	repo, _ := repository.NewRepository(repository.Config{Type: repository.MemoryType})
	svc := service.NewURLService(repo)
	cfg := &config.Config{Base: "http://example.com"}
	h := New(cfg, svc, ctx)

	userID := 42
	_, _ = svc.CreateShortURL(ctx, "http://original.example/5", userID)

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	req = req.WithContext(context.WithValue(req.Context(), ctxUserID, userID))
	rr := httptest.NewRecorder()
	h.UserURLsEndpoint(rr, req)

	fmt.Println(rr.Code)
}

// Example of delete user URLs endpoint.
func ExampleHandler_DeleteUserURLsEndpoint() {
	ctx := context.Background()
	repo, _ := repository.NewRepository(repository.Config{Type: repository.MemoryType})
	svc := service.NewURLService(repo)
	cfg := &config.Config{Base: "http://example.com"}
	h := New(cfg, svc, ctx)

	userID := 99
	tamer, _ := svc.CreateShortURL(ctx, "http://original.example/6", userID)

	payload, _ := json.Marshal([]string{tamer.ShortURL})
	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), ctxUserID, userID))
	rr := httptest.NewRecorder()
	h.DeleteUserURLsEndpoint(rr, req)

	fmt.Println(rr.Code)
}

// Example of ping endpoint.
func ExampleHandler_PingEndpoint() {
	ctx := context.Background()
	repo, _ := repository.NewRepository(repository.Config{Type: repository.MemoryType})
	svc := service.NewURLService(repo)
	cfg := &config.Config{Base: "http://example.com"}
	h := New(cfg, svc, ctx)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()
	h.PingEndpoint(rr, req)

	fmt.Println(rr.Code)
}
