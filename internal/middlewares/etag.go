package middlewares

import (
	"bytes"
	"fmt"
	"hash"
	"hash/fnv"
	"io"
	"net/http"
	"sync"
)

// Etag returns a middleware that generates an ETag for responses
// with a body size less than or equal to maxBodySize. If 'enabled' is false,
// it simply returns the next handler without wrapping it.
func Etag(enabled bool, maxBodySize int64) func(http.Handler) http.Handler {
	if !enabled {
		// Middleware is disabled; return the next handler as-is.
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	bufferPool := sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	hashPool := sync.Pool{
		New: func() interface{} {
			return fnv.New64a()
		},
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip ETag generation for HEAD requests since they don't return a body
			if r.Method == http.MethodHead {
				next.ServeHTTP(w, r)
				return
			}

			// Get a buffer and hasher from the pools.
			buf := bufferPool.Get().(*bytes.Buffer)
			buf.Reset()
			defer bufferPool.Put(buf)

			hasher := hashPool.Get().(hash.Hash)
			hasher.Reset()
			defer hashPool.Put(hasher)

			// Wrap the ResponseWriter.
			rw := &etagResponseWriter{
				ResponseWriter: w,
				buf:            buf,
				hasher:         hasher,
				maxSize:        maxBodySize,
				headers:        w.Header(),
				statusCode:     0, // Indicates no status code has been set yet
			}

			next.ServeHTTP(rw, r)

			// If headers have not been written yet, proceed
			if !rw.headerWritten {
				// Ensure the status code is set
				if rw.statusCode == 0 {
					rw.statusCode = http.StatusOK
				}

				// Only proceed if the response was fully buffered and status code is in the 200 range.
				if rw.size > 0 && rw.size <= rw.maxSize && rw.err == nil && rw.statusCode >= 200 && rw.statusCode < 300 {
					// Compute the ETag.
					etag := fmt.Sprintf("\"%x\"", hasher.Sum(nil))
					// Set the ETag header.
					rw.headers.Set("ETag", etag)

					// Check If-None-Match header.
					if match := r.Header.Get("If-None-Match"); match != "" && match == etag {
						// Return 304 Not Modified.
						rw.statusCode = http.StatusNotModified
						rw.headers.Del("Content-Type")
						rw.headers.Del("Content-Length")
						rw.writeHeader()
						return
					}
				}

				// Write headers and buffered response.
				rw.writeHeader()
				_, _ = io.Copy(w, buf)
			}
			// For responses where headers have already been written, we do not alter the response.
		})
	}
}

// etagResponseWriter wraps http.ResponseWriter to compute the SHA1 hash
// of the response body up to a maximum size.
type etagResponseWriter struct {
	http.ResponseWriter
	buf           *bytes.Buffer
	hasher        hash.Hash
	size          int64
	maxSize       int64
	statusCode    int
	headers       http.Header
	headerWritten bool
	err           error
}

// WriteHeader implements the http.ResponseWriter interface.
func (w *etagResponseWriter) WriteHeader(statusCode int) {
	if !w.headerWritten {
		w.statusCode = statusCode
	}
}

// writeHeader writes the headers to the underlying ResponseWriter.
func (w *etagResponseWriter) writeHeader() {
	if !w.headerWritten {
		w.ResponseWriter.WriteHeader(w.statusCode)
		w.headerWritten = true
	}
}

// Write implements the http.ResponseWriter interface.
func (w *etagResponseWriter) Write(p []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}

	n := len(p)
	w.size += int64(n)

	if w.size <= w.maxSize && !w.headerWritten {
		// Buffer the data.
		_, err := w.buf.Write(p)
		if err != nil {
			w.err = err
			return 0, fmt.Errorf("failed to buffer response: %w", err)
		}
		// Update the hash.
		_, err = w.hasher.Write(p)
		if err != nil {
			w.err = err
			return 0, fmt.Errorf("failed to hash response: %w", err)
		}
		return n, nil
	}

	// If headers have not been written yet, write them now.
	if !w.headerWritten {
		if w.statusCode == 0 {
			w.statusCode = http.StatusOK
		}
		w.writeHeader()
		// Write any buffered data.
		if w.buf.Len() > 0 {
			_, err := w.ResponseWriter.Write(w.buf.Bytes())
			if err != nil {
				w.err = err
				return 0, fmt.Errorf("failed to write buffered response: %w", err)
			}
			w.buf.Reset()
		}
	}
	// Write the current data directly.
	n, err := w.ResponseWriter.Write(p)
	if err != nil {
		w.err = err
		return 0, fmt.Errorf("failed to write response: %w", err)
	}

	return n, nil
}
