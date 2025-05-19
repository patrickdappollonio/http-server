package fileutil

import "testing"

func TestShouldForceDownload(t *testing.T) {
	tests := []struct {
		name       string
		filename   string
		extensions []string
		skipFiles  []string
		want       bool
	}{
		{
			name:       "should return false when no extensions are provided",
			filename:   "file.js",
			extensions: []string{},
			skipFiles:  nil,
			want:       false,
		},
		{
			name:       "should match exact extension",
			filename:   "file.js",
			extensions: []string{"js"},
			skipFiles:  nil,
			want:       true,
		},
		{
			name:       "should be case insensitive",
			filename:   "file.JS",
			extensions: []string{"js"},
			skipFiles:  nil,
			want:       true,
		},
		{
			name:       "should match compound extensions",
			filename:   "file.min.js",
			extensions: []string{"min.js"},
			skipFiles:  nil,
			want:       true,
		},
		{
			name:       "should return false when no match is found",
			filename:   "file.txt",
			extensions: []string{"js", "css"},
			skipFiles:  nil,
			want:       false,
		},
		{
			name:       "should handle extensions with leading dots",
			filename:   "file.js",
			extensions: []string{".js"},
			skipFiles:  nil,
			want:       true,
		},
		{
			name:       "should match exact filenames without extension",
			filename:   "Dockerfile",
			extensions: []string{"Dockerfile"},
			skipFiles:  nil,
			want:       true,
		},
		{
			name:       "should match any of multiple extensions",
			filename:   "file.min.js",
			extensions: []string{"js", "min.js"},
			skipFiles:  nil,
			want:       true,
		},
		{
			name:       "should match exact base filename without path",
			filename:   "Dockerfile",
			extensions: []string{"Dockerfile"},
			skipFiles:  nil,
			want:       true,
		},
		{
			name:       "should match exact base filename in subdirectory",
			filename:   "subdir/Dockerfile",
			extensions: []string{"Dockerfile"},
			skipFiles:  nil,
			want:       true,
		},
		{
			name:       "should skip file that matches exact filename in skip list",
			filename:   "file.js",
			extensions: []string{"js"},
			skipFiles:  []string{"file.js"},
			want:       false,
		},
		{
			name:       "should skip file that matches base filename in skip list",
			filename:   "/path/to/file.js",
			extensions: []string{"js"},
			skipFiles:  []string{"file.js"},
			want:       false,
		},
		{
			name:       "should skip file that matches path suffix in skip list",
			filename:   "/path/to/file.js",
			extensions: []string{"js"},
			skipFiles:  []string{"to/file.js"},
			want:       false,
		},
		{
			name:       "skip list comparison should be case insensitive",
			filename:   "file.JS",
			extensions: []string{"js"},
			skipFiles:  []string{"file.js"},
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldForceDownload(tt.filename, tt.extensions, tt.skipFiles); got != tt.want {
				t.Errorf("ShouldForceDownload() = %v, want %v", got, tt.want)
			}
		})
	}
}
