package handler

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/akarashov/urltamer/internal/config"
	"go.uber.org/zap"
)

var (
	mutex  sync.Mutex
	tamers = make(map[string]string)
)

const tamerLength = 8

type (
	Handler struct {
		Base *string
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

func New(c *config.Config) *Handler {
	c.Base = strings.TrimRight(c.Base, "/")
	return &Handler{
		Base: &c.Base,
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
		tamer := makeTamer()
		if _, exist := tamers[tamer]; !exist {
			for _, v := range tamers {
				if request.URL == v {
					http.Error(res, "Double URL", http.StatusBadRequest)
					return
				}
			}
			tamers[tamer] = request.URL

			response.Result = fmt.Sprintf("%s/%s", *h.Base, tamer)
			resp, err := json.Marshal(response)
			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
			res.Header().Set("Content-Type", "application/json")
			res.WriteHeader(http.StatusCreated)
			res.Write(resp)
		} else {
			http.Error(res, "Double Tamer", http.StatusBadRequest)
		}
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
		tamer := makeTamer()
		if _, exist := tamers[tamer]; !exist {
			for _, v := range tamers {
				if reqURL == v {
					http.Error(res, "Double URL", http.StatusBadRequest)
					return
				}
			}
			tamers[tamer] = reqURL
			res.WriteHeader(http.StatusCreated)
			res.Header().Set("Content-Type", "text/plain")
			fmt.Fprintf(res, "%s/%s", *h.Base, tamer)
		} else {
			http.Error(res, "Double Tamer", http.StatusBadRequest)
		}
	}
}

func (h *Handler) ResponseEndpoint(res http.ResponseWriter, req *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()
	tamer := req.URL.Path[1:]
	if reqURL, exist := tamers[tamer]; exist {
		http.Redirect(res, req, reqURL, http.StatusTemporaryRedirect)
	} else {
		http.Error(res, "Not found", http.StatusBadRequest)
	}
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
			compress_res := newCompressWriter(res)
			res = compress_res
			defer compress_res.Close()
		}
		if sendsGzip {
			compress_req, err := newCompressReader(req.Body)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
			req.Body = compress_req
			defer compress_req.Close()
		}
		wrapped(res, req)
	}
}
