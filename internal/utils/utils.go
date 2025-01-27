package utils

import (
	"path"
	"reflect"
	"time"
)

func RFC1123(t time.Time) string {
	return t.Format(time.RFC1123)
}

func PrettyTime(t time.Time) string {
	return t.Format("Jan 2, 2006 3:04pm MST")
}

func CanonicalURL(isDir bool, p ...string) string {
	s := path.Join(p...)

	if isDir {
		s = s + "/"
	}

	return s
}

// DefaultValue returns the first non-empty value.
func DefaultValue(d interface{}, given ...interface{}) interface{} {
	if Empty(given) || Empty(given[0]) {
		return d
	}
	return given[0]
}

// Empty returns true if the given value has the zero value for its type.
func Empty(given interface{}) bool {
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
