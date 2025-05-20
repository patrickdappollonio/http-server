package fileutil

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// FileExists checks if a file exists at the given path.
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

// IsMarkdownFile checks if a file is a markdown file based on its extension.
func IsMarkdownFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".md" || ext == ".markdown"
}

// IsIndexFile checks if a filename is in the allowedIndexFiles list.
func IsIndexFile(filename string, allowedIndexFiles []string) bool {
	return slices.Contains(allowedIndexFiles, filename)
}

// IsHTMLIndexFile checks if a filename is a standard HTML index file.
func IsHTMLIndexFile(filename string, htmlIndexFiles []string) bool {
	return slices.Contains(htmlIndexFiles, filename)
}
