package main

import "net/http"

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLRW(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
