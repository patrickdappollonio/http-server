package main

import (
	"os"
	"strings"
)

// exists returns whether a folder exists or not in the filesystem
func exists(path string) bool {
	_, err := os.Stat(path)

	if err != nil || os.IsNotExist(err) {
		return false
	}

	return true
}

// envany returns the first non empty value of a set of
// environment variables
func envany(key ...string) string {
	for _, v := range key {
		if s := strings.TrimSpace(os.Getenv(v)); s != "" {
			return s
		}
	}

	return ""
}

// firstNonEmpty returns the first non-empty string
// and if all strings are empty, it returns the defval
func firstNonEmpty(defval string, value ...string) string {
	for _, v := range value {
		if v != "" {
			return v
		}
	}

	return defval
}

var availableColors = map[string]struct{}{
	"cyan":   {},
	"teal":   {},
	"green":  {},
	"light":  {},
	"lime":   {},
	"yellow": {},
	"amber":  {},
	"orange": {},
	"brown":  {},
	"grey":   {},
	"deep":   {},
	"red":    {},
	"purple": {},
	"blue":   {},
	"indigo": {},
	"pink":   {},
}

// isAvailableColor returns true or false depending on if the available color
// exists in getmdl.io
func isAvailableColor(colorCouple string) bool {
	items := strings.Split(colorCouple, "-")

	if len(items) != 2 {
		return false
	}

	if _, found := availableColors[items[0]]; !found {
		return false
	}

	if _, found := availableColors[items[1]]; !found {
		return false
	}

	return true
}
