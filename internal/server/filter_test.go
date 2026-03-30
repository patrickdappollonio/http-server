package server

import "testing"

func Test_isFiltered(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		prefix   []string
		suffix   []string
		match    []string
		want     bool
	}{
		{
			name:     "match prefix",
			filename: "test.txt",
			prefix:   []string{"test"},
			want:     true,
		},
		{
			name:     "match suffix",
			filename: "test.txt",
			suffix:   []string{".txt"},
			want:     true,
		},
		{
			name:     "exact match",
			filename: "test.txt",
			match:    []string{"test.txt"},
			want:     true,
		},
		{
			name:     "no exact nor partial match",
			filename: "test.txt",
			prefix:   []string{"test2"},
			suffix:   []string{".txt2"},
			match:    []string{"test.txt2"},
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				forbiddenPrefixes: tt.prefix,
				forbiddenSuffixes: tt.suffix,
				forbiddenMatches:  tt.match,
			}
			if got := s.isFiltered(tt.filename); got != tt.want {
				t.Errorf("isFiltered() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsAbsolutePathForbidden(t *testing.T) {
	tests := []struct {
		name           string
		forbiddenPaths []string
		checkPath      string
		want           bool
	}{
		{
			name:           "cert in served dir is hidden",
			forbiddenPaths: []string{"/srv/www/cert.pem"},
			checkPath:      "/srv/www/cert.pem",
			want:           true,
		},
		{
			name:           "cert outside served dir is not hidden",
			forbiddenPaths: []string{"/etc/tls/cert.pem"},
			checkPath:      "/srv/www/cert.pem",
			want:           false,
		},
		{
			name:           "same name in different directory is not hidden",
			forbiddenPaths: []string{"/srv/www/cert.pem"},
			checkPath:      "/srv/www/subdir/cert.pem",
			want:           false,
		},
		{
			name:           "empty forbidden paths",
			forbiddenPaths: nil,
			checkPath:      "/srv/www/cert.pem",
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				forbiddenAbsPaths: tt.forbiddenPaths,
			}
			if got := s.isAbsolutePathForbidden(tt.checkPath); got != tt.want {
				t.Errorf("isAbsolutePathForbidden() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsAbsolutePathForbidden_PrefixBlocking(t *testing.T) {
	tests := []struct {
		name             string
		forbiddenPrefixes []string
		checkPath        string
		want             bool
	}{
		{
			name:             "certmagic dir itself is blocked",
			forbiddenPrefixes: []string{"/srv/www/.certmagic/"},
			checkPath:        "/srv/www/.certmagic",
			want:             false, // exact dir match uses forbiddenAbsPaths, not prefix
		},
		{
			name:             "file inside certmagic dir is blocked",
			forbiddenPrefixes: []string{"/srv/www/.certmagic/"},
			checkPath:        "/srv/www/.certmagic/acme/cert.pem",
			want:             true,
		},
		{
			name:             "nested deep file is blocked",
			forbiddenPrefixes: []string{"/srv/www/.certmagic/"},
			checkPath:        "/srv/www/.certmagic/acme-v02.api.letsencrypt.org-directory/sites/example.com/example.com.key",
			want:             true,
		},
		{
			name:             "unrelated path is not blocked",
			forbiddenPrefixes: []string{"/srv/www/.certmagic/"},
			checkPath:        "/srv/www/index.html",
			want:             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				forbiddenAbsPathPrefixes: tt.forbiddenPrefixes,
			}
			if got := s.isAbsolutePathForbidden(tt.checkPath); got != tt.want {
				t.Errorf("isAbsolutePathForbidden() = %v, want %v", got, tt.want)
			}
		})
	}
}
