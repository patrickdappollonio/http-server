package server

var forbiddenMatches = []string{
	"_redirects",
}

var (
	forbiddenPrefixes = []string{}
	forbiddenSuffixes = []string{}
)

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

		if len(filename) >= len(p) && filename[:len(p)] == p {
			return true
		}
	}

	for _, s := range allSuffixes {
		if s == "" {
			continue
		}

		if len(filename) >= len(s) && filename[len(filename)-len(s):] == s {
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
