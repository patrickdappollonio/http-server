package mw

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"hash"
	"net/http"
)

// etagResponseWriter is a wrapper around http.ResponseWriter that will
// calculate the ETag header for the response.
type etagResponseWriter struct {
	rw     http.ResponseWriter
	buf    *bytes.Buffer
	status int
	hash   hash.Hash
}

// Header returns the header map that will be sent by WriteHeader.
func (e etagResponseWriter) Header() http.Header {
	return e.rw.Header()
}

// WriteHeader sends an HTTP response header with the provided status code.
func (e *etagResponseWriter) WriteHeader(status int) {
	e.status = status
}

// Write writes the data to the connection as part of an HTTP reply.
func (e *etagResponseWriter) Write(p []byte) (int, error) {
	if e.status == 0 {
		e.status = http.StatusOK
	}

	e.hash.Write(p)
	return e.buf.Write(p)
}

// hasHashed returns true if the response has been hashed, or if it's a fit
// candidate for having a hash.
func (e *etagResponseWriter) hasHashed() bool {
	return e.hash != nil && e.buf.Len() > 0 && e.status < http.StatusBadRequest && e.status != http.StatusNoContent
}

// Etag is a middleware that will calculate the ETag header for the response.
func Etag(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ew := etagResponseWriter{
			rw:   w,
			buf:  bytes.NewBuffer(nil),
			hash: sha1.New(),
		}

		next.ServeHTTP(&ew, r)

		if !ew.hasHashed() {
			w.Write(ew.buf.Bytes())
			return
		}

		sum := hex.EncodeToString(ew.hash.Sum(nil))
		w.Header().Set("ETag", "\""+sum+"\"")

		if r.Header.Get("If-None-Match") == sum {
			w.WriteHeader(http.StatusNotModified)
			w.Write(nil)
			return
		}

		w.WriteHeader(ew.status)
		w.Write(ew.buf.Bytes())
	})
}
