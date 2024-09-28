package redirects

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestRedirectionEngine(t *testing.T) {
	tests := []struct {
		name                 string
		rules                string
		visitedPath          string
		expectHittingHandler bool
		expectStatusCode     int
		expectLocation       string
		expectError          bool
	}{
		{
			name:                 "empty content",
			rules:                "",
			expectHittingHandler: true,
			expectStatusCode:     http.StatusOK,
		},
		{
			name:             "simple rule - permanent",
			rules:            "/old /new permanent",
			visitedPath:      "/old",
			expectStatusCode: http.StatusMovedPermanently,
			expectLocation:   "/new",
		},
		{
			name:             "simple rule - temporary",
			rules:            "/old /new temporary",
			visitedPath:      "/old",
			expectStatusCode: http.StatusFound,
			expectLocation:   "/new",
		},
		{
			name:                 "simple rule - no match",
			rules:                "/old /new temporary",
			visitedPath:          "/something",
			expectHittingHandler: true,
			expectStatusCode:     http.StatusOK,
		},
		{
			name: "simple rule - with comment",
			rules: `# This is a comment
							/old /new temporary`,
			visitedPath:      "/old",
			expectStatusCode: http.StatusFound,
			expectLocation:   "/new",
		},
		{
			name:             "simple rule - inline comment",
			rules:            "/old /new temporary # This is a comment",
			visitedPath:      "/old",
			expectStatusCode: http.StatusFound,
			expectLocation:   "/new",
		},
		{
			name:             "simple rule - with deleting query parameters",
			rules:            "/old /new temporary",
			visitedPath:      "/old?foo=bar",
			expectStatusCode: http.StatusFound,
			expectLocation:   "/new",
		},
		{
			name:             "simple rule - keep query parameters during redirect - temporary",
			rules:            "/old?! /new temporary",
			visitedPath:      "/old?name=foo",
			expectStatusCode: http.StatusFound,
			expectLocation:   "/new?name=foo",
		},
		{
			name:             "on the root - keep query parameters during redirect - permanent",
			rules:            "/?! /api/v1 permanent",
			visitedPath:      "/?name=foo",
			expectStatusCode: http.StatusMovedPermanently,
			expectLocation:   "/api/v1?name=foo",
		},
		{
			name:        "invalid rule - completely invalid",
			rules:       "this is an invalid rule",
			expectError: true,
		},
		{
			name:        "invalid rule - only one parameter",
			rules:       "/old",
			expectError: true,
		},
		{
			name:        "invalid rule - missing status code",
			rules:       "/old /new",
			expectError: true,
		},
		{
			name:        "invalid rule - invalid status code",
			rules:       "/old /new invalid",
			expectError: true,
		},
		{
			name:        "invalid rule - too many parameters",
			rules:       "/old /new temporary extra",
			expectError: true,
		},
		{
			name: "multiple rules - first one matches",
			rules: `# This is a comment
							/old /new temporary
							/old2 /new2 temporary`,
			visitedPath:      "/old",
			expectStatusCode: http.StatusFound,
			expectLocation:   "/new",
		},
		{
			name: "multiple rules - second one matches",
			rules: `# This is a comment
							/old /new temporary
							/old2 /new2 temporary`,
			visitedPath:      "/old2",
			expectStatusCode: http.StatusFound,
			expectLocation:   "/new2",
		},
		{
			name: "multiple rules - no match",
			rules: `# This is a comment
							/old /new temporary
							/old2 /new2 temporary`,
			visitedPath:          "/something",
			expectHittingHandler: true,
			expectStatusCode:     http.StatusOK,
		},
		{
			name: "multiple rules - first one matches, keep query parameters",
			rules: `# This is a comment
							/old?! /new temporary
							/old2 /new2 temporary`,
			visitedPath:      "/old?name=foo",
			expectStatusCode: http.StatusFound,
			expectLocation:   "/new?name=foo",
		},
		{
			name: "multiple rules - second one matches, keep query parameters",
			rules: `# This is a comment
							/old /new temporary
							/old2?! /new2 temporary`,
			visitedPath:      "/old2?name=foo",
			expectStatusCode: http.StatusFound,
			expectLocation:   "/new2?name=foo",
		},
		{
			name: "multiple rules - no match, keep query parameters",
			rules: `# This is a comment
							/old?! /new temporary
							/old2 /new2 temporary`,
			visitedPath:          "/something?name=foo",
			expectHittingHandler: true,
			expectStatusCode:     http.StatusOK,
		},
		{
			name:             "single rule - include one parameter",
			rules:            "/posts/:id /posts?id=:id temporary",
			visitedPath:      "/posts/123",
			expectStatusCode: http.StatusFound,
			expectLocation:   "/posts?id=123",
		},
		{
			name:             "single rule - include multiple parameters",
			rules:            "/posts/:id/comments/:comment /posts?id=:id&comment=:comment temporary",
			visitedPath:      "/posts/123/comments/456",
			expectStatusCode: http.StatusFound,
			expectLocation:   "/posts?id=123&comment=456",
		},
		{
			name:             "single rule - include multiple parameters, keep query parameters",
			rules:            "/posts/:id/comments/:comment?! /posts?id=:id&comment=:comment temporary",
			visitedPath:      "/posts/123/comments/456?name=foo",
			expectStatusCode: http.StatusFound,
			expectLocation:   "/posts?id=123&comment=456&name=foo",
		},
		{
			name:             "single rule - include multiple parameters, keep query parameters, overwrite repeated parameters",
			rules:            "/posts/:id?! /posts?id=:id temporary",
			visitedPath:      "/posts/123?id=456",
			expectStatusCode: http.StatusFound,
			expectLocation:   "/posts?id=123",
		},
		{
			name:             "single rule - from querystring to path parameters",
			rules:            "/posts?id=:id /posts/:id temporary",
			visitedPath:      "/posts?id=123",
			expectStatusCode: http.StatusFound,
			expectLocation:   "/posts/123",
		},
		{
			name:             "single rule - from querystring to path parameters, keep query parameters",
			rules:            "/posts?id=:id?! /posts/:id temporary",
			visitedPath:      "/posts?id=123&name=foo",
			expectStatusCode: http.StatusFound,
			expectLocation:   "/posts/123?name=foo",
		},
		{
			name:             "single rule - from querystring to path parameters, keep query parameters, overwrite repeated parameters",
			rules:            "/posts?id=:id?! /posts/:id temporary",
			visitedPath:      "/posts?id=123&id=456",
			expectStatusCode: http.StatusFound,
			expectLocation:   "/posts/123",
		},
		{
			name:                 "single rule - query parameters, incomplete request url",
			rules:                "/posts?id=:id /posts temporary",
			visitedPath:          "/posts",
			expectHittingHandler: true,
			expectStatusCode:     http.StatusOK,
		},
		{
			name:             "splat",
			rules:            "/posts/* /posts temporary",
			visitedPath:      "/posts/123",
			expectStatusCode: http.StatusFound,
			expectLocation:   "/posts",
		},
		{
			name:             "splat - alternate form",
			rules:            "/posts/:splat /posts temporary",
			visitedPath:      "/posts/123/456",
			expectStatusCode: http.StatusFound,
			expectLocation:   "/posts",
		},
		{
			name:             "splat - alternate form - use the splat parameter",
			rules:            "/posts/:splat /articles/:splat temporary",
			visitedPath:      "/posts/123/456",
			expectStatusCode: http.StatusFound,
			expectLocation:   "/articles/123/456",
		},
		{
			name:             "splat - keep query parameters",
			rules:            "/posts/*?! /posts temporary",
			visitedPath:      "/posts/123?name=foo",
			expectStatusCode: http.StatusFound,
			expectLocation:   "/posts?name=foo",
		},
		{
			name:             "query param just key",
			rules:            "/posts?! /posts temporary",
			visitedPath:      "/posts?name",
			expectStatusCode: http.StatusFound,
			expectLocation:   "/posts?name=",
		},
		{
			name:             "query param just key after previous key",
			rules:            "/posts?! /posts temporary",
			visitedPath:      "/posts?name&foo=bar",
			expectStatusCode: http.StatusFound,
			expectLocation:   "/posts?name=&foo=bar",
		},
		{
			name:             "weird query parameters",
			rules:            "/posts?! /posts-alt temporary",
			visitedPath:      "/posts?&&&",
			expectStatusCode: http.StatusFound,
			expectLocation:   "/posts-alt",
		},
		{
			name:             "query and splat",
			rules:            "/posts/*?! /posts/example temporary",
			visitedPath:      "/posts/technology/2024/12/22/hello-world?name=foo",
			expectStatusCode: http.StatusFound,
			expectLocation:   "/posts/example?name=foo",
		},
		{
			name:             "query and splat - alternate form",
			rules:            "/posts/:splat?! /posts/example/:splat temporary",
			visitedPath:      "/posts/technology/2024/12/22/hello-world?name=foo",
			expectStatusCode: http.StatusFound,
			expectLocation:   "/posts/example/technology/2024/12/22/hello-world?name=foo",
		},
		{
			name:        "splat parameter used not at the end",
			rules:       "/:splat/foo/bar /posts/:splat temporary",
			visitedPath: "/hello/foo/bar",
			expectError: true,
		},
		{
			name:        "splat wildcard used not at the end",
			rules:       "/foo/*/bar /posts temporary",
			visitedPath: "/foo/123/bar",
			expectError: true,
		},
		{
			name:             "redirect to absolute url",
			rules:            "/articles/:id https://www.example.com/blog/articles/:id permanent",
			visitedPath:      "/articles/123",
			expectStatusCode: http.StatusMovedPermanently,
			expectLocation:   "https://www.example.com/blog/articles/123",
		},
		{
			name:             "redirect to absolute url - keep query parameters",
			rules:            "/articles/:id?! https://www.example.com/blog/articles/:id permanent",
			visitedPath:      "/articles/123?name=foo",
			expectStatusCode: http.StatusMovedPermanently,
			expectLocation:   "https://www.example.com/blog/articles/123?name=foo",
		},
		{
			name:             "redirect to absolute url - overwrite repeated parameters",
			rules:            "/articles/:id?! https://www.example.com/blog/articles/:id permanent",
			visitedPath:      "/articles/123?id=456",
			expectStatusCode: http.StatusMovedPermanently,
			expectLocation:   "https://www.example.com/blog/articles/123",
		},
		{
			name:             "redirect to absolute url - use splat wildcard",
			rules:            "/articles/* https://www.example.com/blog/articles permanent",
			visitedPath:      "/articles/123/456",
			expectStatusCode: http.StatusMovedPermanently,
			expectLocation:   "https://www.example.com/blog/articles",
		},
		{
			name:             "redirect to absolute url - use splat wildcard - keep query parameters",
			rules:            "/articles/*?! https://www.example.com/blog/articles permanent",
			visitedPath:      "/articles/123/456?name=foo",
			expectStatusCode: http.StatusMovedPermanently,
			expectLocation:   "https://www.example.com/blog/articles?name=foo",
		},
		{
			name:             "redirect to absolute url - use splat parameter",
			rules:            "/articles/:splat https://www.example.com/blog/articles/:splat permanent",
			visitedPath:      "/articles/123/456",
			expectStatusCode: http.StatusMovedPermanently,
			expectLocation:   "https://www.example.com/blog/articles/123/456",
		},
		{
			name:             "redirect to absolute url - use splat parameter - keep query parameters",
			rules:            "/articles/:splat?! https://www.example.com/blog/articles/:splat permanent",
			visitedPath:      "/articles/123/456?name=foo",
			expectStatusCode: http.StatusMovedPermanently,
			expectLocation:   "https://www.example.com/blog/articles/123/456?name=foo",
		},
		{
			name:             "escaped colon in path parameter",
			rules:            `/:id/foo\:bar /posts/:id temporary`,
			visitedPath:      "/123/foo:bar",
			expectStatusCode: http.StatusFound,
			expectLocation:   "/posts/123",
		},
		{
			name:             "escaped colon in querystring parameter",
			rules:            `/:id?name=foo\:bar /posts/:id temporary`,
			visitedPath:      "/123?name=foo:bar",
			expectStatusCode: http.StatusFound,
			expectLocation:   "/posts/123",
		},
		{
			name:        "unescaped parameter within text",
			rules:       "/foo:id/foo/* /foo:id/foo temporary",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redirector, err := New(tt.rules)
			if tt.expectError {
				if err == nil {
					t.Fatalf("expecting error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("not expecting error, got: %v", err)
			}

			handlerHit := false
			handler := redirector.Middleware(os.Stderr)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !tt.expectHittingHandler {
					t.Fatalf("not expecting hitting the handler, it should've redirected: URL requested: %q", r.URL.String())
				}

				handlerHit = true
				w.WriteHeader(http.StatusOK)
			}))

			rec := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, tt.visitedPath, nil)
			if err != nil {
				t.Fatalf("not expecting error, got: %v", err)
			}

			handler.ServeHTTP(rec, req)

			if tt.expectHittingHandler && !handlerHit {
				t.Fatalf("expecting hitting the handler, but it didn't")
			}

			if rec.Code != tt.expectStatusCode {
				t.Fatalf("expected status code %d, got %d - response: %s", tt.expectStatusCode, rec.Code, rec.Body.String())
			}

			if loc := rec.Header().Get("Location"); loc != tt.expectLocation {
				t.Fatalf("expected location %q, got %q", tt.expectLocation, loc)
			}
		})
	}
}
