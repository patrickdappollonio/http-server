package middlewares

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type logResponseWriter struct {
	rw           http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (lrw *logResponseWriter) Header() http.Header {
	return lrw.rw.Header()
}

func (lrw *logResponseWriter) Write(p []byte) (int, error) {
	n, err := lrw.rw.Write(p)
	lrw.bytesWritten += int64(n)
	return n, err
}

func (lrw *logResponseWriter) WriteHeader(statusCode int) {
	lrw.rw.WriteHeader(statusCode)
	lrw.statusCode = statusCode
}

// LogRequest middleware
func LogRequest(output io.Writer, format string, redactedQuerystringFields ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap the response writer
			lrw := &logResponseWriter{
				rw: w,
			}

			// Call the next middleware or handler
			next.ServeHTTP(lrw, r)

			// Log the request after all middlewares have completed
			urlpath := r.URL.Path
			if r.URL.Query().Encode() != "" {
				querystrings := r.URL.Query()
				for _, key := range redactedQuerystringFields {
					if _, ok := querystrings[key]; ok {
						querystrings.Set(key, "REDACTED")
					}
				}
				urlpath = r.URL.Path + "?" + querystrings.Encode()
			}

			// Get the status code or 200
			statusCode := lrw.statusCode
			if statusCode == 0 {
				statusCode = http.StatusOK
			}

			// Log the request details
			s := strings.NewReplacer(
				"{http_method}", r.Method,
				"{url}", urlpath,
				"{proto}", r.Proto,
				"{status_code}", fmt.Sprintf("%d", statusCode),
				"{status_text}", http.StatusText(statusCode),
				"{duration}", time.Since(start).String(),
				"{bytes_written}", fmt.Sprintf("%d", lrw.bytesWritten),
			).Replace(format)

			fmt.Fprintln(output, s)
		})
	}
}
