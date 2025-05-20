package ctype

import (
	"net/http"
	"path"
	"unicode/utf8"

	"github.com/saintfish/chardet"
)

// DetectContentType determines the content type and charset for a file.
func DetectContentType(filePath string, data []byte, bytesRead int) (string, string) {
	// Try to determine content type by filename first
	contentType := GetContentTypeForFilename(path.Base(filePath))

	// If no content type was detected by filename and we have content to examine,
	// try to detect it from the content
	if contentType == "" && bytesRead > 0 {
		detected := http.DetectContentType(data[:bytesRead])
		if detected != "application/octet-stream" {
			contentType = detected
		}
	}

	// Detect charset if we have valid content type and data to examine
	charset := ""
	if bytesRead > 0 && contentType != "" && contentType != "application/octet-stream" {
		// Check if data is valid UTF-8
		if utf8.Valid(data[:bytesRead]) {
			charset = "utf-8"
		} else {
			// Try to detect charset with chardet
			res, err := chardet.NewTextDetector().DetectBest(data[:bytesRead])
			if err == nil && res.Confidence > 50 && res.Charset != "" {
				charset = res.Charset
			}
		}
	}

	// Add charset for text-based content types
	if charset != "" && contentType != "application/octet-stream" {
		contentType += "; charset=" + charset
	}

	return contentType, charset
}
