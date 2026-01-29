package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/akarashov/urltamer/internal/model"
	"github.com/akarashov/urltamer/internal/repository"
	"github.com/akarashov/urltamer/internal/service"
	"go.uber.org/zap"
)

type mockSubject struct{}

func (m *mockSubject) Register(o service.Observer)   {}
func (m *mockSubject) Deregister(o service.Observer) {}
func (m *mockSubject) Notify(e model.AuditEvent)     {}

func makeHandler() (*Handler, func()) {
	base := "http://base"
	repo := repository.NewMemoryRepository()
	svc := service.NewURLService(repo)
	cleanup := func() {
		if err := repo.Close(); err != nil {
			// can't call b.Fatalf here; callers should handle failures if needed
		}
	}
	return &Handler{Service: svc, Context: context.Background(), Base: &base}, cleanup
}

func BenchmarkPingEndpoint(b *testing.B) {
	h, cleanup := makeHandler()
	defer cleanup()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		h.PingEndpoint(rr, req)
	}
}

func BenchmarkIsConflictResolver(b *testing.B) {
	h, cleanup := makeHandler()
	defer cleanup()
	for i := 0; i < b.N; i++ {
		h.isConflictResolver(nil)
	}
}

func BenchmarkRequestEndpoint(b *testing.B) {
	h, cleanup := makeHandler()
	defer cleanup()
	body := bytes.NewBufferString("https://example")
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(bytes.NewReader(body.Bytes())))
		h.RequestEndpoint(rr, req)
	}
}

func BenchmarkResponseEndpoint(b *testing.B) {
	h, cleanup := makeHandler()
	defer cleanup()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/s", nil)
		req.URL.Path = "/s"
		h.ResponseEndpoint(rr, req)
	}
}

func BenchmarkRequestJSONEndpoint(b *testing.B) {
	h, cleanup := makeHandler()
	defer cleanup()
	payload := map[string]string{"url": "https://json"}
	data, err := json.Marshal(payload)
	if err != nil {
		b.Fatalf("json marshal error: %v", err)
	}
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/json", io.NopCloser(bytes.NewReader(data)))
		req.Header.Set("Content-Type", "application/json")
		h.RequestJSONEndpoint(rr, req)
	}
}

func BenchmarkRequestJSONEndpointBatch(b *testing.B) {
	h, cleanup := makeHandler()
	defer cleanup()
	batch := []model.RequestBatch{{CorrelationID: "1", OriginalURL: "https://a"}}
	data, err := json.Marshal(batch)
	if err != nil {
		b.Fatalf("json marshal error: %v", err)
	}
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/jsonb", io.NopCloser(bytes.NewReader(data)))
		req.Header.Set("Content-Type", "application/json")
		h.RequestJSONEndpointBatch(rr, req)
	}
}

func BenchmarkDeleteUserURLsEndpoint(b *testing.B) {
	h, cleanup := makeHandler()
	defer cleanup()
	shortURLs := []string{"s1"}
	data, err := json.Marshal(shortURLs)
	if err != nil {
		b.Fatalf("json marshal error: %v", err)
	}
	// create token cookie
	token, err := BuildJWTString(1, secretKey, time.Hour)
	if err != nil {
		b.Fatalf("BuildJWTString error: %v", err)
	}
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/del", io.NopCloser(bytes.NewReader(data)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Cookie", cookieName+"="+token)
		h.DeleteUserURLsEndpoint(rr, req)
	}
}

func BenchmarkUserURLsEndpoint(b *testing.B) {
	h, cleanup := makeHandler()
	defer cleanup()
	token, err := BuildJWTString(1, secretKey, time.Hour)
	if err != nil {
		b.Fatalf("BuildJWTString error: %v", err)
	}
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/user", nil)
		req.Header.Set("Cookie", cookieName+"="+token)
		h.UserURLsEndpoint(rr, req)
	}
}

func BenchmarkCookieMiddleware(b *testing.B) {
	mw := CookieMiddleware(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		mw(rr, req)
	}
}

func BenchmarkUserIDFromRequest_Context(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), ctxUserID, 42))
	for i := 0; i < b.N; i++ {
		UserIDFromRequest(req)
	}
}

func BenchmarkUserIDFromRequest_Cookie(b *testing.B) {
	token, err := BuildJWTString(2, secretKey, time.Hour)
	if err != nil {
		b.Fatalf("BuildJWTString error: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Cookie", cookieName+"="+token)
	for i := 0; i < b.N; i++ {
		UserIDFromRequest(req)
	}
}

func BenchmarkBuildJWTString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := BuildJWTString(1, secretKey, time.Minute)
		if err != nil {
			b.Fatalf("BuildJWTString error: %v", err)
		}
	}
}

func BenchmarkGetUserID(b *testing.B) {
	token, err := BuildJWTString(3, secretKey, time.Minute)
	if err != nil {
		b.Fatalf("BuildJWTString error: %v", err)
	}
	for i := 0; i < b.N; i++ {
		GetUserID(token)
	}
}

func BenchmarkGenerateUserID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generateUserID()
	}
}

func BenchmarkLoggingMiddlewareRequest(b *testing.B) {
	sl := zap.NewNop().Sugar()
	mw := LoggingMiddlewareRequest(func(w http.ResponseWriter, r *http.Request) {}, *sl)
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/logr", nil)
		mw(rr, req)
	}
}

func BenchmarkLoggingMiddlewareResponse(b *testing.B) {
	sl := zap.NewNop().Sugar()
	mw := LoggingMiddlewareResponse(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello"))
	}, *sl)
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/logr", nil)
		mw(rr, req)
	}
}

func BenchmarkAuditMiddleware(b *testing.B) {
	subj := &mockSubject{}
	mw := AuditMiddleware(func(w http.ResponseWriter, r *http.Request) {}, subj)
	payload := []byte("https://audit")
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/audit", io.NopCloser(bytes.NewReader(payload)))
		mw(rr, req)
	}
}
