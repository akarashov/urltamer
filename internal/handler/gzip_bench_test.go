package handler

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http/httptest"
	"testing"

	"net/http"
)

func BenchmarkGzipMiddleware_ResponseCompression(b *testing.B) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write(bytes.Repeat([]byte("a"), 1024))
	}
	wrapped := GzipMiddleware(handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://example.test/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		wrapped(rr, req)
		res := rr.Result()
		io.ReadAll(res.Body)
		res.Body.Close()
	}
}

func BenchmarkGzipMiddleware_RequestDecompression(b *testing.B) {
	// prepare gzipped payload
	payload := bytes.Repeat([]byte("x"), 1024)
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, _ = gw.Write(payload)
	_ = gw.Close()

	handler := func(w http.ResponseWriter, r *http.Request) {
		io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}
	wrapped := GzipMiddleware(handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "http://example.test/", io.NopCloser(bytes.NewReader(buf.Bytes())))
		req.Header.Set("Content-Encoding", "gzip")
		wrapped(rr, req)
		rr.Result().Body.Close()
	}
}

func BenchmarkCompressWriter_Write(b *testing.B) {
	rr := httptest.NewRecorder()
	cw := newCompressWriter(rr)
	payload := bytes.Repeat([]byte("z"), 512)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cw.Write(payload)
		if err != nil {
			b.Fatalf("compress write error: %v", err)
		}
	}
	_ = cw.Close()
}
