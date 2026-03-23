package common

import (
	"testing"
)

func TestCanonicalURL(t *testing.T) {
	tests := []struct {
		name   string
		isDir  bool
		parts  []string
		expect string
	}{
		{
			name:   "simple file",
			isDir:  false,
			parts:  []string{"/files", "readme.txt"},
			expect: "/files/readme.txt",
		},
		{
			name:   "simple directory",
			isDir:  true,
			parts:  []string{"/files", "subdir"},
			expect: "/files/subdir/",
		},
		{
			name:   "pound sign in filename",
			isDir:  false,
			parts:  []string{"/files", "track#1.mp3"},
			expect: "/files/track%231.mp3",
		},
		{
			name:   "pound sign in directory",
			isDir:  true,
			parts:  []string{"/files", "C# Projects"},
			expect: "/files/C%23%20Projects/",
		},
		{
			name:   "space in filename",
			isDir:  false,
			parts:  []string{"/files", "my file.txt"},
			expect: "/files/my%20file.txt",
		},
		{
			name:   "question mark in filename",
			isDir:  false,
			parts:  []string{"/files", "what?.txt"},
			expect: "/files/what%3F.txt",
		},
		{
			name:   "percent in filename",
			isDir:  false,
			parts:  []string{"/files", "100%.txt"},
			expect: "/files/100%25.txt",
		},
		{
			name:   "multiple special chars",
			isDir:  false,
			parts:  []string{"/docs", "file #2 (copy).txt"},
			expect: "/docs/file%20%232%20%28copy%29.txt",
		},
		{
			name:   "nested path with pound",
			isDir:  false,
			parts:  []string{"/music", "artist#name", "track#1.mp3"},
			expect: "/music/artist%23name/track%231.mp3",
		},
		{
			name:   "root directory file",
			isDir:  false,
			parts:  []string{"/", "file.txt"},
			expect: "/file.txt",
		},
		{
			name:   "single segment file",
			isDir:  false,
			parts:  []string{"file.txt"},
			expect: "file.txt",
		},
		{
			name:   "no special chars",
			isDir:  false,
			parts:  []string{"/path", "to", "file.txt"},
			expect: "/path/to/file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CanonicalURL(tt.isDir, tt.parts...)
			if got != tt.expect {
				t.Errorf("CanonicalURL(%v, %v) = %q, want %q", tt.isDir, tt.parts, got, tt.expect)
			}
		})
	}
}
