// Package fileutil provides utility functions for file operations.
package fileutil

import (
	"path/filepath"
	"strings"
)

// ShouldForceDownload checks if a file should be force-downloaded based on its extension.
// It first checks if the file is in the skip list (either by exact match or path),
// and if not, checks if it should be force-downloaded based on the filename or extension.
// The check is case-insensitive and also matches exact filenames (e.g., "Dockerfile").
func ShouldForceDownload(filename string, forceExtensions, skipFiles []string) bool {
	// Early return: if no force extensions are provided, nothing to force download
	if len(forceExtensions) == 0 {
		return false
	}

	// Get the base filename for comparison
	baseFilename := filepath.Base(filename)

	// Prepare lowercase versions for case-insensitive comparisons
	lowerFilename := strings.ToLower(filename)
	lowerBaseFilename := strings.ToLower(baseFilename)

	// Check if this file should be skipped from force download
	for _, skipFile := range skipFiles {
		lowerSkipFile := strings.ToLower(skipFile)

		// Check for exact match with full path or base filename
		if lowerFilename == lowerSkipFile || lowerBaseFilename == lowerSkipFile {
			return false
		}

		// Check if the skip file is a path suffix
		if strings.HasSuffix(lowerFilename, "/"+lowerSkipFile) {
			return false
		}
	}

	// Check against force download extensions
	for _, ext := range forceExtensions {
		lowerExt := strings.ToLower(ext)

		// Check for exact match with full path
		if lowerFilename == lowerExt {
			return true
		}

		// Check for exact match with base filename (e.g., "Dockerfile")
		if lowerBaseFilename == lowerExt {
			return true
		}

		// Handle extension with leading dot (e.g., ".txt")
		if strings.HasPrefix(ext, ".") && strings.HasSuffix(lowerBaseFilename, lowerExt) {
			return true
		}

		// Handle extension without leading dot (e.g., "txt")
		if !strings.HasPrefix(ext, ".") && strings.HasSuffix(lowerBaseFilename, "."+lowerExt) {
			return true
		}
	}

	return false
}
