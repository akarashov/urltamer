package handler

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/akarashov/urltamer/internal/config"
	"github.com/akarashov/urltamer/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestResponseEndpoint(t *testing.T) {
	type want struct {
		statusCode int
		location   string
	}
	tests := []struct {
		name string
		res  http.ResponseWriter
		req  *http.Request
		want want
	}{{
		name: "Valid Tamer",
		res:  httptest.NewRecorder(),
		req:  httptest.NewRequest(http.MethodGet, "/QAZwsxed", nil),
		want: want{
			statusCode: http.StatusTemporaryRedirect,
			location:   "http://example.com",
		}}, {
		name: "inValid Tamer",
		res:  httptest.NewRecorder(),
		req:  httptest.NewRequest(http.MethodGet, "/QAZwsxrf", nil),
		want: want{
			statusCode: http.StatusBadRequest,
			location:   "",
		},
	}, {
		name: "Valid Tamer with query",
		res:  httptest.NewRecorder(),
		req:  httptest.NewRequest(http.MethodGet, "/QAZwsxed?iddqd=idkfa", nil),
		want: want{
			statusCode: http.StatusTemporaryRedirect,
			location:   "http://example.com",
		},
	}, {
		name: "Blank Tamer",
		res:  httptest.NewRecorder(),
		req:  httptest.NewRequest(http.MethodGet, "/", nil),
		want: want{
			statusCode: http.StatusBadRequest,
			location:   "",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(&config.Config{Base: "http://127.0.0.1:8080/", Listen: ":8080"})
			c.Tamers = append(c.Tamers, model.Tamer{UUID: "1", ShortURL: "QAZwsxed", OriginalURL: "http://example.com"})
			c.ResponseEndpoint(tt.res, tt.req)
			result := tt.res.(*httptest.ResponseRecorder)
			assert.Equal(t, tt.want.location, result.Header().Get("Location"))
			assert.Equal(t, tt.want.statusCode, result.Code)
		})
	}
}

func TestRequestEndpoint(t *testing.T) {
	type want struct {
		statusCode int
		body       string
	}
	tstFile := "./test.json"

	tests := []struct {
		name string
		res  http.ResponseWriter
		req  *http.Request
		want want
	}{
		{
			name: "Valid URL",
			res:  httptest.NewRecorder(),
			req:  httptest.NewRequest(http.MethodPost, "/", strings.NewReader("http://example.dev")),
			want: want{
				statusCode: http.StatusCreated,
				body:       "http://127.0.0.1:8080/",
			}}, {
			name: "Blank URL",
			res:  httptest.NewRecorder(),
			req:  httptest.NewRequest(http.MethodPost, "/", strings.NewReader("")),
			want: want{
				statusCode: http.StatusBadRequest,
				body:       "Error body parse",
			},
		}, {
			name: "Double URL",
			res:  httptest.NewRecorder(),
			req:  httptest.NewRequest(http.MethodPost, "/", strings.NewReader("http://example.com")),
			want: want{
				statusCode: http.StatusBadRequest,
				body:       "Double URL",
			},
		}, {
			name: "Long URL",
			res:  httptest.NewRecorder(),
			req:  httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://make-a-url-longer.nathanvarner.com/redirect-to-new-url/index.html?long-url=First%20there%20was%20nothing%20and%20then%20a%20new%20character%20came%20along%20h%20and%20then%20a%20new%20character%20came%20along%20t%20and%20then%20a%20new%20character%20came%20along%20t%20and%20then%20a%20new%20character%20came%20along%20p%20and%20then%20a%20new%20character%20came%20along%20s%20and%20then%20a%20new%20character%20came%20along%20:%20and%20then%20a%20new%20character%20came%20along%20/%20and%20then%20a%20new%20character%20came%20along%20/%20and%20then%20a%20new%20character%20came%20along%20y%20and%20then%20a%20new%20character%20came%20along%20a%20and%20then%20a%20new%20character%20came%20along%20.%20and%20then%20a%20new%20character%20came%20along%20r%20and%20then%20a%20new%20character%20came%20along%20u")),
			want: want{
				statusCode: http.StatusCreated,
				body:       "http://127.0.0.1:8080/",
			},
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(&config.Config{Base: "http://127.0.0.1:8080/", Listen: ":8080", FileStoragePath: tstFile})
			c.Tamers = append(c.Tamers, model.Tamer{UUID: "1", ShortURL: "QAZwsxed", OriginalURL: "http://example.com"})
			c.RequestEndpoint(tt.res, tt.req)
			result := tt.res.(*httptest.ResponseRecorder)
			assert.Contains(t, result.Body.String(), tt.want.body)
			assert.Equal(t, tt.want.statusCode, result.Code)
			
		})
		
	}
	os.Remove(tstFile)
}

func TestRequestJSONEndpoint(t *testing.T) {
	type want struct {
		statusCode int
		body       string
	}
	tstFile := "./test.json"

	tests := []struct {
		name string
		res  http.ResponseWriter
		req  *http.Request
		want want
	}{
		{
			name: "Valid URL",
			res:  httptest.NewRecorder(),
			req:  httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url":"http://example.dev"}`)),
			want: want{
				statusCode: http.StatusCreated,
				body:       "http://127.0.0.1:8080/",
			}}, {
			name: "Blank URL",
			res:  httptest.NewRecorder(),
			req:  httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url":""}`)),
			want: want{
				statusCode: http.StatusBadRequest,
				body:       "Error body parse",
			},
		}, {
			name: "Double URL",
			res:  httptest.NewRecorder(),
			req:  httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url":"http://example.com"}`)),
			want: want{
				statusCode: http.StatusBadRequest,
				body:       "Double URL",
			},
		}, {
			name: "Long URL",
			res:  httptest.NewRecorder(),
			req:  httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url":"https://make-a-url-longer.nathanvarner.com/redirect-to-new-url/index.html?long-url=First%20there%20was%20nothing%20and%20then%20a%20new%20character%20came%20along%20h%20and%20then%20a%20new%20character%20came%20along%20t%20and%20then%20a%20new%20character%20came%20along%20t%20and%20then%20a%20new%20character%20came%20along%20p%20and%20then%20a%20new%20character%20came%20along%20s%20and%20then%20a%20new%20character%20came%20along%20:%20and%20then%20a%20new%20character%20came%20along%20/%20and%20then%20a%20new%20character%20came%20along%20/%20and%20then%20a%20new%20character%20came%20along%20y%20and%20then%20a%20new%20character%20came%20along%20a%20and%20then%20a%20new%20character%20came%20along%20.%20and%20then%20a%20new%20character%20came%20along%20r%20and%20then%20a%20new%20character%20came%20along%20u"}`)),
			want: want{
				statusCode: http.StatusCreated,
				body:       "http://127.0.0.1:8080/",
			},
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(&config.Config{Base: "http://127.0.0.1:8080/", Listen: ":8080", FileStoragePath: tstFile})
			c.Tamers = append(c.Tamers, model.Tamer{UUID: "1", ShortURL: "QAZwsxed", OriginalURL: "http://example.com"})
			tt.req.Header.Set("Content-Type", "application/json")
			c.RequestJSONEndpoint(tt.res, tt.req)
			result := tt.res.(*httptest.ResponseRecorder)
			assert.Equal(t, tt.want.statusCode, result.Code)
			if tt.name == "Double URL" || tt.name == "Blank URL" {
				assert.Contains(t, result.Body.String(), tt.want.body)
			} else {
				var respObj Response
				err := json.Unmarshal(result.Body.Bytes(), &respObj)
				assert.NoError(t, err)
				assert.Contains(t, respObj.Result, tt.want.body)
			}

		})
	}
	os.Remove(tstFile)
}

func TestGzipMiddleware(t *testing.T) {
	gzipBody := func(s string) []byte {
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		gz.Write([]byte(s))
		gz.Close()
		return buf.Bytes()
	}

	tests := []struct {
		name            string
		acceptEncoding  string
		contentEncoding string
		requestBody     []byte
		expectGzipResp  bool
		expectStatus    int
		expectRespBody  string
	}{
		{
			name:           "Compression",
			acceptEncoding: "gzip",
			requestBody:    nil,
			expectGzipResp: true,
			expectStatus:   http.StatusOK,
			expectRespBody: "http://example.com",
		},
		{
			name:           "No compression",
			acceptEncoding: "",
			requestBody:    nil,
			expectGzipResp: false,
			expectStatus:   http.StatusOK,
			expectRespBody: "http://example.com",
		},
		{
			name:            "Decompression",
			contentEncoding: "gzip",
			requestBody:     gzipBody("http://example.com"),
			expectGzipResp:  false,
			expectStatus:    http.StatusOK,
			expectRespBody:  "http://example.com",
		},
		{
			name:            "Compression and Decompression",
			acceptEncoding:  "gzip",
			contentEncoding: "gzip",
			requestBody:     gzipBody("http://example.com"),
			expectGzipResp:  true,
			expectStatus:    http.StatusOK,
			expectRespBody:  "http://example.com",
		},
		{
			name:            "Invalid",
			contentEncoding: "gzip",
			requestBody:     []byte("not gzip"),
			expectGzipResp:  false,
			expectStatus:    http.StatusInternalServerError,
			expectRespBody:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				if tt.name == "Invalid" {
					assert.Fail(t, "handler should not be called for invalid gzip")
					return
				}
				b, err := io.ReadAll(r.Body)
				assert.NoError(t, err)
				if tt.requestBody != nil {
					assert.Equal(t, "http://example.com", string(b))
				}
				w.Write([]byte("http://example.com"))
			}
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(tt.requestBody))
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}
			if tt.contentEncoding != "" {
				req.Header.Set("Content-Encoding", tt.contentEncoding)
			}
			rec := httptest.NewRecorder()
			GzipMiddleware(handler)(rec, req)
			result := rec.Result()
			defer result.Body.Close()
			assert.Equal(t, tt.expectStatus, result.StatusCode)
			if tt.expectGzipResp {
				assert.Equal(t, "gzip", result.Header.Get("Content-Encoding"))
				gr, err := gzip.NewReader(result.Body)
				assert.NoError(t, err)
				body, err := io.ReadAll(gr)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectRespBody, string(body))
			} else {
				body, err := io.ReadAll(result.Body)
				assert.NoError(t, err)
				if tt.expectRespBody != "" {
					assert.Equal(t, tt.expectRespBody, string(body))
				}
			}
		})
	}
}
