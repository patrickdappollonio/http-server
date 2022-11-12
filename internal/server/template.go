package server

import (
	"embed"
	"fmt"
	"html/template"
	"path"
	"reflect"
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
		"default":        dfault,
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

// dfault returns the first non-empty value.
func dfault(d interface{}, given ...interface{}) interface{} {
	if empty(given) || empty(given[0]) {
		return d
	}
	return given[0]
}

// empty returns true if the given value has the zero value for its type.
func empty(given interface{}) bool {
	g := reflect.ValueOf(given)
	if !g.IsValid() {
		return true
	}

	// Basically adapted from text/template.isTrue
	switch g.Kind() {
	default:
		return g.IsNil()
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
		return g.Len() == 0
	case reflect.Bool:
		return !g.Bool()
	case reflect.Complex64, reflect.Complex128:
		return g.Complex() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return g.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return g.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return g.Float() == 0
	case reflect.Struct:
		return false
	}
}
