// Package pathutil provides utility functions for path and URL operations.
package pathutil

import (
	"path"
	"strings"
)

// GetParentURL returns the parent URL for the given location.
func GetParentURL(base, loc string) string {
	if loc == base {
		return ""
	}

	// Handle directory paths, ensuring they end with a slash
	dir := path.Dir(strings.TrimSuffix(loc, "/"))
	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}

	return dir
}
