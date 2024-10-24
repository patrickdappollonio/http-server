package server

import (
	"html/template"
	"io"

	"github.com/patrickdappollonio/http-server/internal/redirects"
)

const repositoryURL = "https://github.com/patrickdappollonio/http-server/"

// Server is an HTTP server with optional directory listing enabled
type Server struct {
	// Core settings
	Port                 int    `flagName:"port" validate:"required,min=1,max=65535"`
	Path                 string `flagName:"path" validate:"required,dir"`
	PathPrefix           string `flagName:"pathprefix" validate:"omitempty,ispathprefix"`
	PageTitle            string `flagName:"title" validate:"omitempty,max=100"`
	BannerMarkdown       string `flagName:"banner" validate:"omitempty,max=1000"`
	cachedBannerMarkdown string
	LogOutput            io.Writer
	DisableDirectoryList bool

	// Custom NOTFOUND setting
	CustomNotFoundPage       string
	CustomNotFoundStatusCode int

	// Basic auth settings
	Username string `flagName:"username" validate:"omitempty,alphanum,excluded_with=JWTSigningKey"`
	Password string `flagName:"password" validate:"omitempty,alphanum,excluded_with=JWTSigningKey"`

	// Boolean specific settings
	CorsEnabled        bool
	HideLinks          bool
	ETagDisabled       bool
	ETagMaxSize        string
	etagMaxSizeBytes   int64
	GzipEnabled        bool
	DisableCacheBuster bool
	DisableMarkdown    bool
	MarkdownBeforeDir  bool

	// Redirection handling
	DisableRedirects bool
	redirects        *redirects.Engine

	// JWT Specific settings
	JWTSigningKey    string `flagName:"jwt-key" validate:"omitempty,excluded_with=Username,excluded_with=Password"`
	ValidateTimedJWT bool

	// Viper config settings
	ConfigFilePrefix string

	// Internal fields
	cacheBuster       string
	templates         *template.Template
	version           string
	forbiddenPrefixes []string
	forbiddenSuffixes []string
	forbiddenMatches  []string
}

// IsBasicAuthEnabled returns true if the server has been configured with
// a username and password
func (s *Server) IsBasicAuthEnabled() bool {
	return s.Username != "" && s.Password != ""
}

// SetVersion sets the server version
func (s *Server) SetVersion(version string) {
	s.version = version
}
