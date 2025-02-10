package utils

import (
	"path"
	"reflect"
	"time"
)

// RFC1123 returns the time formatted as RFC1123.
func RFC1123(t time.Time) string {
	return t.Format(time.RFC1123)
}

// PrettyTime returns the time formatted as "Jan 2, 2006 3:04pm MST".
func PrettyTime(t time.Time) string {
	return t.Format("Jan 2, 2006 3:04pm MST")
}

// CanonicalURL returns a canonical URL for the given path.
func CanonicalURL(isDir bool, p ...string) string {
	s := path.Join(p...)

	if isDir {
		s += "/"
	}

	return s
}

// DefaultValue returns the first non-empty value.
//
//nolint:ireturn // works with any type and used from templates
func DefaultValue(defaultValue any, given ...any) any {
	for _, g := range given {
		if !Empty(g) {
			return g
		}
	}

	return defaultValue
}

// Empty returns true if the given value has the zero value for its type.
func Empty(given interface{}) bool {
	g := reflect.ValueOf(given)
	if !g.IsValid() {
		return true
	}

	// Basically adapted from text/template.isTrue
	//nolint:exhaustive // all types are handled with default
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
