package server

import (
	"embed"
	"fmt"
	"html/template"
	"path"
	"time"

	"github.com/patrickdappollonio/http-server/internal/utils"
)

// walkTemplatesFS embeds the templates used to render the directory listing
//
//go:embed templates/*
var walkTemplatesFS embed.FS

// generateTemplates generates the templates used to render the directory listing
func (s *Server) generateTemplates() (*template.Template, error) {
	tplfuncs := template.FuncMap{
		"assetpath":      s.assetpath,
		"rfc1123":        rfc1123,
		"prettytime":     prettytime,
		"humansize":      utils.Humansize,
		"canonicalURL":   canonicalURL,
		"getIconForFile": getIconForFile,
		"unsafeHTML":     func(s string) template.HTML { return template.HTML(s) },
	}

	wtfs, err := template.New("").Funcs(tplfuncs).ParseFS(walkTemplatesFS, "templates/*")
	if err != nil {
		return nil, fmt.Errorf("unable to parse internal templates: this is likely a development error: %w", err)
	}

	return wtfs, nil
}

func (s *Server) assetpath(p string) string {
	return path.Join(s.PathPrefix, specialPath, s.cacheBuster, "assets", p)
}

func rfc1123(t time.Time) string {
	return t.Format(time.RFC1123)
}

func prettytime(t time.Time) string {
	return t.Format("Jan 2, 2006 3:04pm MST")
}

func canonicalURL(isDir bool, p ...string) string {
	s := path.Join(p...)

	if isDir {
		s = s + "/"
	}

	return s
}
