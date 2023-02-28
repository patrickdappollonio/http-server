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
