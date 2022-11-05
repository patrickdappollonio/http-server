package server

import "fmt"

const startupPrefix = " >"

func (s *Server) PrintStartup() {
	fmt.Fprintln(s.LogOutput, "SETUP:")

	if s.PathPrefix != "" && s.PathPrefix != "/" {
		fmt.Fprintln(s.LogOutput, startupPrefix, "Path prefix:", s.PathPrefix)
	}

	if s.CorsEnabled {
		fmt.Fprintln(s.LogOutput, startupPrefix, "CORS headers enabled: adding \"Access-Control-Allow-Origin=*\" header")
	}

	if s.IsAuthEnabled() {
		fmt.Fprintln(s.LogOutput, startupPrefix, "Basic authentication enabled with username:", s.Username)
	}

	if s.PageTitle != "" {
		fmt.Fprintln(s.LogOutput, startupPrefix, "Custom page title:", s.PageTitle)
	}

	if s.HideLinks {
		fmt.Fprintln(s.LogOutput, startupPrefix, "Consider helping the project here:", repositoryURL)
	}

	fmt.Fprintln(s.LogOutput, startupPrefix, "Configured to use port:", s.Port)
	fmt.Fprintln(s.LogOutput, startupPrefix, "Serving path:", s.Path)
}
