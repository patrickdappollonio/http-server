package middlewares

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Test that responses under the size limit receive an ETag header.
func TestEtag_UnderLimitSetsHeader(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "hello world")
	})
	mw := Etag(true, 64)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	mw(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d", rr.Code, http.StatusOK)
	}
	if etag := rr.Header().Get("ETag"); etag == "" {
		t.Error("expected ETag header to be set")
	}
}

// Test that a matching If-None-Match header returns 304 Not Modified.
func TestEtag_IfNoneMatchReturnsNotModified(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "hello world")
	})
	mw := Etag(true, 64)

	// First request to obtain the ETag
	rr1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("GET", "/", nil)
	mw(handler).ServeHTTP(rr1, req1)
	etag := rr1.Header().Get("ETag")
	if etag == "" {
		t.Fatal("missing ETag from first response")
	}

	// Second request with If-None-Match
	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.Header.Set("If-None-Match", etag)
	mw(handler).ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusNotModified {
		t.Fatalf("status = %d; want %d", rr2.Code, http.StatusNotModified)
	}
	if got := rr2.Header().Get("ETag"); got != etag {
		t.Errorf("ETag = %q; want %q", got, etag)
	}
}

// Test that responses over the size limit do not include an ETag header.
func TestEtag_OverLimitNoHeader(t *testing.T) {
	body := strings.Repeat("a", 20)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, body)
	})
	mw := Etag(true, 10)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	mw(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d", rr.Code, http.StatusOK)
	}
	if etag := rr.Header().Get("ETag"); etag != "" {
		t.Errorf("unexpected ETag header %q", etag)
	}
}
