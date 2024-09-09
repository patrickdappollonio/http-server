package mw

import (
	"crypto/sha1"
	"encoding/hex"
	"hash"
	"net/http"
)

// etagResponseWriter is a wrapper around http.ResponseWriter that will
// calculate the ETag header for the response as it streams the data.
type etagResponseWriter struct {
	rw     http.ResponseWriter
	hash   hash.Hash
	status int
}

// Header returns the header map that will be sent by WriteHeader.
func (e *etagResponseWriter) Header() http.Header {
	return e.rw.Header()
}

// WriteHeader sends an HTTP response header with the provided status code.
// We don't write the status code just yet to the original response writer,
// since our goal is to calculate the ETag then update the status code if
// there was a match.
func (e *etagResponseWriter) WriteHeader(status int) {
	e.status = status
}

// Write writes the data to the connection as part of an HTTP reply, while
// calculating the ETag on the fly to avoid buffering large amounts of data.
func (e *etagResponseWriter) Write(p []byte) (int, error) {
	if e.status == 0 {
		e.WriteHeader(http.StatusOK)
	}

	// Write the data to the hash for ETag calculation
	e.hash.Write(p)

	// Write the data to the actual response writer
	return e.rw.Write(p)
}

// Etag is a middleware that will calculate the ETag header for the response.
func Etag(enabled bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !enabled {
				next.ServeHTTP(w, r)
				return
			}

			ew := &etagResponseWriter{
				rw:   w,
				hash: sha1.New(),
			}

			// Call the next handler and stream the data while hashing
			next.ServeHTTP(ew, r)

			// If status code is in the range of 200-399, calculate ETag
			if ew.status >= http.StatusOK && ew.status < http.StatusBadRequest {
				etag := hex.EncodeToString(ew.hash.Sum(nil))
				w.Header().Set("ETag", `"`+etag+`"`)

				// Check if the ETag matches the client request
				if r.Header.Get("If-None-Match") == w.Header().Get("ETag") {
					w.WriteHeader(http.StatusNotModified)
					return
				}
			}

			// Write the status code if it hasn't been written yet
			w.WriteHeader(ew.status)
		})
	}
}
