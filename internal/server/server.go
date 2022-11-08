package server

import (
	"html/template"
	"io"
)

const repositoryURL = "https://github.com/patrickdappollonio/http-server/"

// Server is an HTTP server with optional directory listing enabled
type Server struct {
	Port       int    `json:"port" validate:"required,min=1,max=65535"`
	Path       string `json:"path" validate:"required,dir"`
	PathPrefix string `json:"pathprefix" validate:"omitempty,ispathprefix"`
	PageTitle  string `validate:"omitempty,max=100"`

	Username string `json:"username" validate:"omitempty,alphanum"`
	Password string `json:"password" validate:"omitempty,alphanum"`

	CorsEnabled        bool
	HideLinks          bool
	DisableCacheBuster bool
	DisableMarkdown    bool
	MarkdownBeforeDir  bool

	LogOutput io.Writer

	cacheBuster string
	templates   *template.Template
}

// IsAuthEnabled returns true if the server has been configured with
// a username and password
func (s *Server) IsAuthEnabled() bool {
	return s.Username != "" && s.Password != ""
}
