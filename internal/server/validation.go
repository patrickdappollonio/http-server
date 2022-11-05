package server

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// Validate checks the configuration using struct tags and validate
// if the fields are valid per those rules
func (s *Server) Validate() error {
	// Create a custom validator
	var valid = validator.New()

	// Add custom validation rules
	valid.RegisterValidation("ispathprefix", validateIsPathPrefix)

	// Attempt to validate the structure, and grab the errors
	err := valid.Struct(s)
	valerrs, ok := err.(validator.ValidationErrors)

	// If the error isn't empty, and its type is of ValidationError
	// we can provide a better error message for its validation process
	if err != nil && ok {
		var merrs MultiError

		for _, e := range valerrs {
			// Convert the error to a human-readable error
			merrs.Errors = append(merrs.Errors, FieldToValidationError(e))
		}

		return &merrs
	}

	// If the error type is unknown or there's no error
	// return the error as-is
	return err
}

var reIsPathPrefix = regexp.MustCompile(`^\/[\w\-\_]+(\/[\w\-\_]+)*\/$`)

// validateIsPathPrefix checks if the value is a valid path prefix
// which needs to start or end with a forward slash, and include
// within alphanumeric, dashes or underscores, or additional forward slashes
func validateIsPathPrefix(field validator.FieldLevel) bool {
	if field.Field().String() == "/" {
		return true
	}

	return reIsPathPrefix.MatchString(field.Field().String())
}
