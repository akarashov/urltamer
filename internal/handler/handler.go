package handler

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/akarashov/urltamer/internal/config"
	"github.com/akarashov/urltamer/internal/model"
	"github.com/akarashov/urltamer/internal/repository"
	"go.uber.org/zap"
)

var (
	mutex        sync.Mutex
	Tamers       model.Tamers
	TamerCounter int
)

const tamerLength = 8

type (
	Handler struct {
		Base     *string
		FilePath string
	}

	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}

	Request struct {
		URL string `json:"url"`
	}

	Response struct {
		Result string `json:"result"`
	}
)

func makeTamer() string {
	buffer := make([]byte, tamerLength)
	rand.Read(buffer)
	tamer := base64.RawURLEncoding.EncodeToString(buffer)
	return tamer[:tamerLength]
}

func (h *Handler) WriteTamer(originalURL string) (string, error) {
	tamer := makeTamer()
	for _, it := range Tamers {
		if it.ShortURL == tamer {
			return tamer, fmt.Errorf("double URL")
		}
		if it.OriginalURL == originalURL {
			return tamer, fmt.Errorf("double URL")
		}
	}
	TamerCounter++
	newTamer := model.Tamer{UUID: strconv.Itoa(TamerCounter), ShortURL: tamer, OriginalURL: originalURL}
	Tamers = append(Tamers, newTamer)
	if h != nil && h.FilePath != "" {
		if err := repository.SaveToFile(Tamers, h.FilePath); err != nil {
			return tamer, err
		}
	}

	return tamer, nil
}

func New(c *config.Config) *Handler {
	c.Base = strings.TrimRight(c.Base, "/")
	return &Handler{
		Base:     &c.Base,
		FilePath: c.FileStoragePath,
	}
}

func (h *Handler) RequestJSONEndpoint(res http.ResponseWriter, req *http.Request) {
	var request Request
	var response Response
	var buf bytes.Buffer

	if req.Header.Get("Content-Type") != "application/json" {
		http.Error(res, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}
	mutex.Lock()
	defer mutex.Unlock()
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal(buf.Bytes(), &request); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	if request.URL == "" {
		http.Error(res, "Error body parse", http.StatusBadRequest)
	} else {
		tamer, err := h.WriteTamer(request.URL)
		if err != nil {
			http.Error(res, "Double URL", http.StatusBadRequest)
			return
		}
		response.Result = fmt.Sprintf("%s/%s", *h.Base, tamer)
		resp, err := json.Marshal(response)
		if err != nil {
			http.Error(res, "Double URL", http.StatusInternalServerError)
			return
		}
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)
		res.Write(resp)
	}
}

func (h *Handler) RequestEndpoint(res http.ResponseWriter, req *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()
	reqURLb, err := io.ReadAll(req.Body)
	reqURL := string(reqURLb)
	if err != nil || reqURL == "" {
		http.Error(res, "Error body parse", http.StatusBadRequest)
	} else {
		tamer, err := h.WriteTamer(reqURL)
		if err != nil {
			http.Error(res, "Double URL", http.StatusBadRequest)
			return
		}
		res.WriteHeader(http.StatusCreated)
		res.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(res, "%s/%s", *h.Base, tamer)
	}
}

func (h *Handler) ResponseEndpoint(res http.ResponseWriter, req *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()
	tamer := req.URL.Path[1:]
	for _, it := range Tamers {
		if it.ShortURL == tamer {
			http.Redirect(res, req, it.OriginalURL, http.StatusTemporaryRedirect)
			return
		}
	}
	http.Error(res, "Not found", http.StatusBadRequest)
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func LoggingMiddlewareRequest(wrapped http.HandlerFunc, sl zap.SugaredLogger) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		start := time.Now()
		wrapped(res, req)
		duration := time.Since(start)
		sl.Infoln(
			"uri", req.RequestURI,
			"method", req.Method,
			"duration", duration,
		)
	}
}

func LoggingMiddlewareResponse(wrapped http.HandlerFunc, sl zap.SugaredLogger) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lres := loggingResponseWriter{
			ResponseWriter: res,
			responseData:   responseData,
		}
		wrapped(&lres, req)
		sl.Infoln(
			"status", responseData.status,
			"size", responseData.size,
		)
	}
}

func GzipMiddleware(wrapped http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		acceptEncoding := req.Header.Get("Accept-Encoding")
		contentEncoding := req.Header.Get("Content-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if supportsGzip {
			compressRes := newCompressWriter(res)
			res = compressRes
			defer compressRes.Close()
		}
		if sendsGzip {
			compressReq, err := newCompressReader(req.Body)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
			req.Body = compressReq
			defer compressReq.Close()
		}
		wrapped(res, req)
	}
}
