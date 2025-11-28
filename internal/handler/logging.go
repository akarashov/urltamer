package handler

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
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
