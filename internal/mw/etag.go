package mw

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"hash"
	"net/http"
	"sync"
)

var bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

type etagResponseWriter struct {
	hash    hash.Hash
	headers map[string][]string
	buf     *bytes.Buffer
	status  int
}

// Header returns the header map that will be sent by WriteHeader
func (e *etagResponseWriter) Header() http.Header {
	return e.headers
}

// WriteHeader sends an HTTP response header with the provided status code
func (e *etagResponseWriter) WriteHeader(status int) {
	e.status = status
}

// Write writes the data to the connection as part of an HTTP reply
func (e *etagResponseWriter) Write(p []byte) (int, error) {
	// In Go, a call to Write will always
	// set the status code to 200 if it's not set
	if e.status == 0 {
		e.status = http.StatusOK
	}

	// Write the data to the hash for ETag calculation
	e.hash.Write(p)

	// Write the data to the actual response writer
	return e.buf.Write(p)
}

// Etag middleware
func Etag(enabled bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !enabled {
				next.ServeHTTP(w, r)
				return
			}

			buf := bufPool.Get().(*bytes.Buffer)
			defer func() {
				buf.Reset()
				bufPool.Put(buf)
			}()

			alternateWriter := &etagResponseWriter{
				headers: http.Header{},
				buf:     buf,
				hash:    sha1.New(),
			}

			// Call the next handler and stream the data while hashing
			next.ServeHTTP(alternateWriter, r)

			// If the status is in the range of 200-399, calculate ETag
			if alternateWriter.status >= http.StatusOK && alternateWriter.status < http.StatusBadRequest {
				etag := fmt.Sprintf("%q", hex.EncodeToString(alternateWriter.hash.Sum(nil)))
				alternateWriter.Header().Set("Etag", etag)

				// Check if the ETag matches the client request
				if r.Header.Get("If-None-Match") == etag {
					alternateWriter.WriteHeader(http.StatusNotModified)
				}
			}

			// Pass the response to the actual response writer
			for key, vals := range alternateWriter.headers {
				for _, val := range vals {
					w.Header().Add(key, val)
				}
			}
			w.WriteHeader(alternateWriter.status)
			w.Write(alternateWriter.buf.Bytes())
		})
	}
}
