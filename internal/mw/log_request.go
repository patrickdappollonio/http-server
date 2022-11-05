package mw

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// statusResponseWriter is a wrapper around http.ResponseWriter that
// allows us to capture the status code and bytes written
type statusResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

// newStatusResponseWriter returns a new statusResponseWriter
func newStatusResponseWriter(w http.ResponseWriter) *statusResponseWriter {
	return &statusResponseWriter{w, http.StatusOK, 0}
}

// WriteHeader implements the http.ResponseWriter interface
func (lrw *statusResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// Write implements the http.ResponseWriter interface
func (lrw *statusResponseWriter) Write(b []byte) (int, error) {
	bw, err := lrw.ResponseWriter.Write(b)
	lrw.bytesWritten = bw
	return bw, err
}

// LogRequest is a middleware that logs specific request data using a predefined
// template format. Available options are:
// - {http_method} the HTTP method
// - {url} the URL
// - {proto} the protocol version
// - {status_code} the HTTP status code
// - {status_text} the HTTP status text
// - {duration} the duration of the request
// - {bytes_written} the number of bytes written
func LogRequest(output io.Writer, format string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Capture starting time
			start := time.Now()

			// Create a new instance of our custom responsewriter
			lrw := newStatusResponseWriter(w)

			// Serve the next request
			next.ServeHTTP(lrw, r)

			// Generate a string representation of the log message
			s := strings.NewReplacer(
				"{http_method}", r.Method,
				"{url}", r.URL.String(),
				"{proto}", r.Proto,
				"{status_code}", fmt.Sprintf("%d", lrw.statusCode),
				"{status_text}", http.StatusText(lrw.statusCode),
				"{duration}", time.Since(start).String(),
				"{bytes_written}", fmt.Sprintf("%d", lrw.bytesWritten),
			).Replace(format)

			// Print that log message to the output writer
			fmt.Fprintln(output, s)
		})
	}
}
