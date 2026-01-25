package handler

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// compressWriter wraps an http.ResponseWriter to provide gzip compression.
type compressWriter struct {
	w             http.ResponseWriter
	zw            *gzip.Writer
	headerWritten bool
}

// compressReader wraps an io.ReadCloser to provide gzip decompression.
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// GzipMiddleware is a middleware that handles gzip compression and decompression for HTTP requests and responses.
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

// Header returns the header map that will be sent by Write.
func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// Write writes the data to the connection as part of an HTTP reply.
func (c *compressWriter) Write(p []byte) (int, error) {
	if !c.headerWritten {
		c.w.Header().Set("Content-Encoding", "gzip")
		c.w.WriteHeader(http.StatusOK)
		c.headerWritten = true
	}
	return c.zw.Write(p)
}

// WriteHeader sends an HTTP response header with the provided status code.
func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
	c.headerWritten = true
}

// Close closes the gzip writer.
func (c *compressWriter) Close() error {
	return c.zw.Close()
}

// Read reads data from the gzip reader.
func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close closes the gzip reader and the underlying reader.
func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

// newCompressWriter creates a new compressWriter.
func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

// newCompressReader creates a new compressReader.
func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}
