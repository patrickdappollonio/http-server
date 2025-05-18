package server

import (
	"testing"
)

func TestShouldForceDownload(t *testing.T) {
	tests := []struct {
		name       string
		extensions []string
		filename   string
		want       bool
	}{
		{
			name:       "empty extensions list",
			extensions: []string{},
			filename:   "test.jpg",
			want:       false,
		},
		{
			name:       "filename without extension",
			extensions: []string{"jpg", "pdf"},
			filename:   "testfile",
			want:       false,
		},
		{
			name:       "matching extension",
			extensions: []string{"jpg", "pdf", "zip"},
			filename:   "test.pdf",
			want:       true,
		},
		{
			name:       "non-matching extension",
			extensions: []string{"jpg", "pdf", "zip"},
			filename:   "test.txt",
			want:       false,
		},
		{
			name:       "case insensitive match - lowercase in list",
			extensions: []string{"jpg", "pdf"},
			filename:   "test.PDF",
			want:       true,
		},
		{
			name:       "case insensitive match - uppercase in list",
			extensions: []string{"JPG", "PDF"},
			filename:   "test.pdf",
			want:       true,
		},
		{
			name:       "extension with dot",
			extensions: []string{".jpg", "pdf"},
			filename:   "test.jpg",
			want:       true,
		},
		{
			name:       "filename with path",
			extensions: []string{"jpg", "pdf"},
			filename:   "/path/to/file/test.jpg",
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				ForceDownloadExtensions: tt.extensions,
			}

			if got := s.ShouldForceDownload(tt.filename); got != tt.want {
				t.Errorf("ShouldForceDownload() = %v, want %v", got, tt.want)
			}
		})
	}
}
