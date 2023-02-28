package server

func (s *Server) isFiltered(filename string) bool {
	for _, p := range append(s.forbiddenPrefixes, s.ConfigFilePrefix) {
		if p == "" {
			continue
		}

		if len(filename) >= len(p) && filename[:len(p)] == p {
			return true
		}
	}

	for _, s := range s.forbiddenSuffixes {
		if s == "" {
			continue
		}

		if len(filename) >= len(s) && filename[len(filename)-len(s):] == s {
			return true
		}
	}

	for _, m := range s.forbiddenMatches {
		if m == "" {
			continue
		}

		if filename == m {
			return true
		}
	}

	return false
}
