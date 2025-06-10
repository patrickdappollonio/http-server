package server

import "testing"

func TestGetCustomCSSURL(t *testing.T) {
	tests := []struct {
		name      string
		prefix    string
		customCSS string
		want      string
	}{
		{
			name:      "no prefix relative path",
			prefix:    "/",
			customCSS: "style.css",
			want:      "/style.css",
		},
		{
			name:      "prefix relative path",
			prefix:    "/blog/",
			customCSS: "style.css",
			want:      "/blog/style.css",
		},
		{
			name:      "prefix absolute path",
			prefix:    "/blog/",
			customCSS: "/style.css",
			want:      "/blog/style.css",
		},
		{
			name:      "root absolute path",
			prefix:    "/",
			customCSS: "/style.css",
			want:      "/style.css",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{PathPrefix: tt.prefix, CustomCSS: tt.customCSS}
			if got := s.getCustomCSSURL(); got != tt.want {
				t.Errorf("getCustomCSSURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
