package server

import "testing"

func Test_getContentTypeForExtension(t *testing.T) {
	tests := []struct {
		extension string
		want      string
	}{
		{
			extension: ".css",
			want:      "text/css",
		},
		{
			extension: ".html",
			want:      "text/html",
		},
		{
			extension: ".js",
			want:      "text/javascript",
		},
		{
			extension: ".json",
			want:      "application/json",
		},
		{
			extension: ".jpg",
			want:      "image/jpeg",
		},
		{
			extension: ".jpeg",
			want:      "image/jpeg",
		},
		{
			extension: ".png",
			want:      "image/png",
		},
		{
			extension: ".svg",
			want:      "image/svg+xml",
		},
		{
			extension: ".gif",
			want:      "image/gif",
		},
		{
			extension: ".webp",
			want:      "image/webp",
		},
	}
	for _, tt := range tests {
		t.Run(tt.extension, func(t *testing.T) {
			if got := getContentTypeForExtension(tt.extension); got != tt.want {
				t.Errorf("getContentTypeForExtension() = %v, want %v", got, tt.want)
			}
		})
	}
}
