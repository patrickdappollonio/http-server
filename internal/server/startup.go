package server

import "fmt"

const startupPrefix = " >"

func (s *Server) PrintStartup() {
	fmt.Fprintln(s.LogOutput, "SETUP:")

	fmt.Fprintln(s.LogOutput, startupPrefix, "Configured to use port:", s.Port)
	fmt.Fprintln(s.LogOutput, startupPrefix, "Serving path:", s.Path)

	if s.PathPrefix != "" && s.PathPrefix != "/" {
		fmt.Fprintln(s.LogOutput, startupPrefix, "Path prefix:", s.PathPrefix)
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

	s.printWarnings()
}

func (s *Server) printWarnings() {
	if len(s.JWTSigningKey) < 32 {
		s.printWarning("JWT key is less than 32 characters. It can be brute forced easily.")
	}
}
