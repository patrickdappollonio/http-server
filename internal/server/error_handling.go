package server

import (
	"bytes"
	"fmt"
	"unicode"

	"github.com/go-playground/validator/v10"
)

// MultiError is a collection of errors
type MultiError struct {
	Errors []error
}

// Error implements the error interface
func (m *MultiError) Error() string {
	// Handle not having any error
	if len(m.Errors) == 0 {
		return ""
	}

	// For a single error, return the error lowercasing
	// its first letter
	if len(m.Errors) == 1 {
		s := m.Errors[0].Error()
		s = string(unicode.ToLower(rune(s[0]))) + s[1:]
		return s
	}

	var b bytes.Buffer

	for pos, err := range m.Errors {
		b.WriteString("  - " + err.Error())

		if pos != len(m.Errors)-1 {
			b.WriteString("\n")
		}
	}

	return "multiple errors occurred:\n" + b.String()
}

// ValidationError is an error validating a flag value
type ValidationError struct {
	Field string
	Tag   string
	Value interface{}
	Param string
}

// Error implements the error interface
func (v *ValidationError) Error() string {
	var humanMsg string
	switch v.Tag {
	case "max":
		humanMsg = fmt.Sprintf("value must be bigger than %s (for numbers) or longer than %s characters (for text)", v.Param, v.Param)
	case "min":
		humanMsg = fmt.Sprintf("value must be less than %s (for numbers) or smaller than %s characters (for text)", v.Param, v.Param)
	case "ispathprefix":
		humanMsg = "must start and end with a forward slash, and include within alphanumeric, dashes or underscores, or additional forward slashes"
	case "excluded_with":
		humanMsg = fmt.Sprintf("cannot be used in conjunction with %s", v.Param)
	default:
		humanMsg = fmt.Sprintf("%s: %s", v.Tag, v.Param)
	}

	return fmt.Sprintf(
		"Parameter %q is invalid: %s (current value: \"%v\")",
		v.Field, humanMsg, v.Value,
	)
}

// FieldToValidationError converts a validator.FieldError from
// the validator v10 package to a local ValidationError
func FieldToValidationError(field validator.FieldError) *ValidationError {
	return &ValidationError{
		Field: field.Field(),
		Tag:   field.Tag(),
		Value: field.Value(),
		Param: field.Param(),
	}
}
