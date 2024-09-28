package server

import (
	"fmt"
	"os"
	"path"

	"github.com/patrickdappollonio/http-server/internal/redirects"
)

const redirectionsPath = "_redirections"

func (s *Server) getPathToRedirectionsFile() string {
	return path.Join(s.Path, redirectionsPath)
}

func (s *Server) LoadRedirectionsIfEnabled() error {
	// If redirections are disabled, return immediately
	if s.DisableRedirects {
		return nil
	}

	// Load the redirections file
	b, err := os.ReadFile(s.getPathToRedirectionsFile())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("unable to read redirections file at %q: %w", s.getPathToRedirectionsFile(), err)
	}

	// Parse the redirections file
	engine, err := redirects.New(string(b))
	if err != nil {
		return fmt.Errorf("redirection error on file %q: %w", s.getPathToRedirectionsFile(), err)
	}

	// Set the redirections engine
	s.redirects = engine
	return nil
}
