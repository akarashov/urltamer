package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/akarashov/urltamer/internal/config"
	"github.com/akarashov/urltamer/internal/model"
	"github.com/akarashov/urltamer/internal/service"
	"go.uber.org/zap"
)

type (
	Handler struct {
		Service *service.URLService
		Context context.Context
		Base    *string
		Tamers  model.Tamers
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

func New(c *config.Config, s *service.URLService, ctx context.Context) *Handler {
	c.Base = strings.TrimRight(c.Base, "/")
	mc, err := s.GetAllURLs(ctx)
	if err != nil {
		log.Printf("Fatality %s", err)
	}
	return &Handler{
		Service: s,
		Context: ctx,
		Base:    &c.Base,
		Tamers:  mc,
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
		tamer, err := h.Service.CreateShortURL(h.Context, request.URL)
		if err != nil {
			http.Error(res, "Double URL", http.StatusBadRequest)
			return
		}
		response.Result = fmt.Sprintf("%s/%s", *h.Base, tamer.ShortURL)
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

func (h *Handler) RequestJSONEndpointBatch(res http.ResponseWriter, req *http.Request) {
	var requestBatchs model.RequestBatchs
	var response model.ResponseBatchs
	var buf bytes.Buffer

	if req.Header.Get("Content-Type") != "application/json" {
		http.Error(res, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal(buf.Bytes(), &requestBatchs); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	for _, requestBatch := range requestBatchs {
		tamer, err := h.Service.CreateShortURL(h.Context, requestBatch.OriginalURL)
		if err != nil {
			http.Error(res, "Double URL", http.StatusBadRequest)
			break
		}
		rsp := model.ResponseBatch{
			CorrelationID: requestBatch.CorrelationID,
			ShortURL:      fmt.Sprintf("%s/%s", *h.Base, tamer.ShortURL)}
		response = append(response, rsp)
	}
	resp, err := json.Marshal(response)
	if err != nil {
		http.Error(res, "Double URL", http.StatusInternalServerError)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	res.Write(resp)
}

func (h *Handler) RequestEndpoint(res http.ResponseWriter, req *http.Request) {
	reqURLb, err := io.ReadAll(req.Body)
	reqURL := string(reqURLb)
	if err != nil || reqURL == "" {
		http.Error(res, "Error body parse", http.StatusBadRequest)
	} else {
		tamer, err := h.Service.CreateShortURL(h.Context, reqURL)
		if err != nil {
			http.Error(res, "Double URL", http.StatusBadRequest)
			return
		}
		res.WriteHeader(http.StatusCreated)
		res.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(res, "%s/%s", *h.Base, tamer.ShortURL)
	}
}

func (h *Handler) ResponseEndpoint(res http.ResponseWriter, req *http.Request) {
	tamer := req.URL.Path[1:]
	originalURL, err := h.Service.GetOriginalURL(h.Context, tamer)
	if err != nil {
		http.Error(res, "Not found", http.StatusBadRequest)
	} else {
		http.Redirect(res, req, originalURL, http.StatusTemporaryRedirect)
	}
}

func (h *Handler) PingEndpoint(res http.ResponseWriter, req *http.Request) {
	if h.Service.Ping(h.Context) {
		res.WriteHeader(http.StatusOK)
		return
	} else {
		http.Error(res, "Not connected to DB", http.StatusInternalServerError)
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
