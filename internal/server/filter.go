package server

import (
	"strings"
)

// forbiddenMatches is a list of filenames that are forbidden to be served.
// This list is used to prevent sensitive files from being
// served.
var forbiddenMatches = []string{
	"_redirects",
}

// forbiddenPrefixes and forbiddenSuffixes are a list of prefixes that are
// forbidden to be served. This list is used to prevent sensitive
// files from being served.
var (
	forbiddenPrefixes = []string{}
	forbiddenSuffixes = []string{}
)

// isFiltered returns true if the filename is forbidden to be served.
func (s *Server) isFiltered(filename string) bool {
	// Adds the config prefix to the list of forbidden prefixes
	allPrefixes := append(s.forbiddenPrefixes, s.ConfigFilePrefix)

	// Adds the well known prefixes from this project
	allPrefixes = append(allPrefixes, forbiddenPrefixes...)
	allSuffixes := append(s.forbiddenSuffixes, forbiddenSuffixes...)
	allMatches := append(s.forbiddenMatches, forbiddenMatches...)

	for _, p := range allPrefixes {
		if p == "" {
			continue
		}

		if strings.HasPrefix(filename, p) {
			return true
		}
	}

	for _, s := range allSuffixes {
		if s == "" {
			continue
		}

		if strings.HasSuffix(filename, s) {
			return true
		}
	}

	for _, m := range allMatches {
		if m == "" {
			continue
		}

		if filename == m {
			return true
		}
	}

	return false
}
