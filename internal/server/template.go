package server

import (
	"embed"
	"fmt"
	"html/template"
	"path"

	"github.com/patrickdappollonio/http-server/internal/common"
)

// walkTemplatesFS embeds the templates used to render the directory listing
//
//go:embed templates/*
var walkTemplatesFS embed.FS

// generateTemplates generates the templates used to render the directory listing
func (s *Server) generateTemplates() (*template.Template, error) {
	tplfuncs := template.FuncMap{
		"assetpath":      s.assetpath,
		"rfc1123":        common.RFC1123,
		"prettytime":     common.PrettyTime,
		"humansize":      common.Humansize,
		"canonicalURL":   common.CanonicalURL,
		"getIconForFile": getIconForFile,
		"unsafeHTML":     func(s string) template.HTML { return template.HTML(s) },
		"default":        common.DefaultValue[any],
		"serverVersion":  func() string { return s.version },
		"bannerMessage":  s.generateBannerMarkdown,
	}

	wtfs, err := template.New("").Funcs(tplfuncs).ParseFS(walkTemplatesFS, "templates/*")
	if err != nil {
		return nil, fmt.Errorf("unable to parse internal templates: this is likely a development error: %w", err)
	}

	return wtfs, nil
}

// assetpath returns the path to an asset
func (s *Server) assetpath(p string) string {
	return path.Join(s.PathPrefix, specialPath, s.cacheBuster, "assets", p)
}
