package server

import (
	"testing"

	"github.com/patrickdappollonio/http-server/internal/fileutil"
)

// TestShouldForceDownload tests the ShouldForceDownload method of the Server struct
// to ensure it correctly identifies files that should be force-downloaded based on
// their extensions and the skip list.

func TestShouldForceDownload(t *testing.T) {
	tests := []struct {
		name       string
		extensions []string
		skipFiles  []string
		filename   string
		want       bool
	}{
		{
			name:       "should return false when no extensions are provided",
			extensions: []string{},
			skipFiles:  []string{},
			filename:   "test.jpg",
			want:       false,
		},
		{
			name:       "should return false for files without extension",
			extensions: []string{"jpg", "pdf"},
			skipFiles:  []string{},
			filename:   "testfile",
			want:       false,
		},
		{
			name:       "should return true for files with matching extension",
			extensions: []string{"jpg", "pdf", "zip"},
			skipFiles:  []string{},
			filename:   "test.pdf",
			want:       true,
		},
		{
			name:       "should return false for files with non-matching extension",
			extensions: []string{"jpg", "pdf", "zip"},
			skipFiles:  []string{},
			filename:   "test.txt",
			want:       false,
		},
		{
			name:       "should match extensions case-insensitively (uppercase filename)",
			extensions: []string{"jpg", "pdf"},
			skipFiles:  []string{},
			filename:   "test.PDF",
			want:       true,
		},
		{
			name:       "should match extensions case-insensitively (uppercase in list)",
			extensions: []string{"JPG", "PDF"},
			skipFiles:  []string{},
			filename:   "test.pdf",
			want:       true,
		},
		{
			name:       "should handle extensions with leading dots",
			extensions: []string{".jpg", ".pdf"},
			skipFiles:  []string{},
			filename:   "test.pdf",
			want:       true,
		},
		{
			name:       "should match exact filenames with path",
			extensions: []string{"Dockerfile"},
			skipFiles:  []string{},
			filename:   "/path/to/Dockerfile",
			want:       true,
		},
		{
			name:       "should skip files with exact match",
			extensions: []string{"jpg", "pdf"},
			skipFiles:  []string{"test.jpg"},
			filename:   "test.jpg",
			want:       false,
		},
		{
			name:       "should skip files with path",
			extensions: []string{"jpg", "pdf"},
			skipFiles:  []string{"/path/to/test.jpg"},
			filename:   "/path/to/test.jpg",
			want:       false,
		},
		{
			name:       "should skip files with case insensitive match",
			extensions: []string{"jpg", "pdf"},
			skipFiles:  []string{"test.JPG"},
			filename:   "test.jpg",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				ForceDownloadExtensions: tt.extensions,
				SkipForceDownloadFiles:  tt.skipFiles,
			}
			if got := fileutil.ShouldForceDownload(tt.filename, s.ForceDownloadExtensions, s.SkipForceDownloadFiles); got != tt.want {
				t.Errorf("ShouldForceDownload() = %v, want %v", got, tt.want)
			}
		})
	}
}
