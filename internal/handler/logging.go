package handler

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// responseData holds information about the HTTP response.
type responseData struct {
	status int
	size   int
}

// loggingResponseWriter wraps http.ResponseWriter to capture response data.
type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

// Write writes data to the connection as part of an HTTP reply and records the size.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader sends an HTTP response header with the provided status code and records it.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// LoggingMiddlewareRequest logs details about the incoming HTTP request.
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

// LoggingMiddlewareResponse logs details about the outgoing HTTP response.
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
