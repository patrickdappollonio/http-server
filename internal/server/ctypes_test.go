package server

import "testing"

func Test_getContentTypeForExtension(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{
			filename: ".css",
			want:     "text/css",
		},
		{
			filename: ".html",
			want:     "text/html",
		},
		{
			filename: ".js",
			want:     "text/javascript",
		},
		{
			filename: ".json",
			want:     "application/json",
		},
		{
			filename: ".jpg",
			want:     "image/jpeg",
		},
		{
			filename: ".jpeg",
			want:     "image/jpeg",
		},
		{
			filename: ".png",
			want:     "image/png",
		},
		{
			filename: ".svg",
			want:     "image/svg+xml",
		},
		{
			filename: ".gif",
			want:     "image/gif",
		},
		{
			filename: ".webp",
			want:     "image/webp",
		},
		{
			filename: "Makefile",
			want:     "text/x-makefile",
		},
	}
	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			if got := getContentTypeForFilename(tt.filename); got != tt.want {
				t.Errorf("getContentTypeForFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_findNoDuplicatesInContentTypes(t *testing.T) {
	seenExt := map[string]bool{}
	seenFilenames := map[string]bool{}

	for _, ct := range ctypes {
		for _, ext := range ct.Extension {
			if seenExt[ext] {
				t.Errorf("duplicate extension: %s", ext)
			}
			seenExt[ext] = true
		}

		for _, name := range ct.ExactNames {
			if seenFilenames[name] {
				t.Errorf("duplicate filename: %s", name)
			}
			seenFilenames[name] = true
		}
	}
}
