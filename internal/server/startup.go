package server

import (
	"fmt"
	"net/http"
)

const startupPrefix = " >"

func (s *Server) PrintStartup() {
	fmt.Fprintln(s.LogOutput, "SETUP:")

	fmt.Fprintln(s.LogOutput, startupPrefix, "Configured to use port:", s.Port)
	fmt.Fprintln(s.LogOutput, startupPrefix, "Serving path:", s.Path)

	if s.PathPrefix != "" && s.PathPrefix != "/" {
		fmt.Fprintln(s.LogOutput, startupPrefix, "Path prefix:", s.PathPrefix)
	}

	if s.DisableDirectoryList {
		fmt.Fprintln(s.LogOutput, startupPrefix, "Directory listing disabled (including markdown rendering)")
	}

	if s.CustomNotFoundPage != "" {
		fmt.Fprintln(s.LogOutput, startupPrefix, "Using custom 404 page:", s.CustomNotFoundPage)
	}

	if s.CustomNotFoundStatusCode != 0 {
		fmt.Fprintf(s.LogOutput, "%s Using custom 404 status code: \"%d %s\"\n", startupPrefix, s.CustomNotFoundStatusCode, http.StatusText(s.CustomNotFoundStatusCode))
	}

	if s.GzipEnabled {
		fmt.Fprintln(s.LogOutput, startupPrefix, "Gzip compression enabled for supported content types")
	}

	if s.ETagDisabled {
		fmt.Fprintln(s.LogOutput, startupPrefix, "ETag headers disabled")
	} else {
		fmt.Fprintf(s.LogOutput, "%s ETag headers enabled for files smaller than %s\n", startupPrefix, s.ETagMaxSize)
	}

	if s.CorsEnabled {
		fmt.Fprintln(s.LogOutput, startupPrefix, "CORS headers enabled: adding \"Access-Control-Allow-Origin=*\" header")
	}

	if s.IsBasicAuthEnabled() {
		fmt.Fprintln(s.LogOutput, startupPrefix, "Basic authentication enabled with username:", s.Username)
	}

	if s.JWTSigningKey != "" {
		fmt.Fprintln(s.LogOutput, startupPrefix, "JWT authentication enabled with given key")

		if s.ValidateTimedJWT {
			fmt.Fprintln(s.LogOutput, startupPrefix, "JWT claims \"exp\" and \"nbf\" will be validated")
		}
	}

	if s.PageTitle != "" {
		fmt.Fprintln(s.LogOutput, startupPrefix, "Custom page title:", s.PageTitle)
	}

	if s.HideLinks {
		fmt.Fprintln(s.LogOutput, startupPrefix, "Consider helping the project here:", repositoryURL)
	}

	if s.DisableCacheBuster {
		fmt.Fprintln(s.LogOutput, startupPrefix, "Cache busting for static assets disabled")
	}

	if s.DisableMarkdown {
		fmt.Fprintln(s.LogOutput, startupPrefix, "Markdown rendering disabled")
	}

	if s.MarkdownBeforeDir {
		fmt.Fprintln(s.LogOutput, startupPrefix, "Markdown rendering before directory listing enabled")
	}

	if !s.DisableRedirects {
		if s.redirects != nil {
			fmt.Fprintf(s.LogOutput, "%s Redirections enabled from %q (found %d redirections)\n", startupPrefix, s.getPathToRedirectionsFile(), len(s.redirects.Rules))
		}
	}

	s.printWarnings()
}

func (s *Server) printWarnings() {
	if s.JWTSigningKey != "" && len(s.JWTSigningKey) < 32 {
		s.printWarning("JWT key is less than 32 characters. It can be brute forced easily.")
	}
}
