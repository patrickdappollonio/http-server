package server

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/patrickdappollonio/http-server/internal/common"
)

const warnPrefix = "[WARNING] >>> "

// _valid is a global validator instance
var _valid *validator.Validate

// getValidator returns a global validator instance or
// initializes it if it's not set
func getValidator() *validator.Validate {
	if _valid == nil {
		// Create a custom validator
		_valid = validator.New()
		// Add custom validation rules
		_valid.RegisterValidation("ispathprefix", validateIsPathPrefix)
	}

	return _valid
}

// Validate checks the configuration using struct tags and validate
// if the fields are valid per those rules
func (s *Server) Validate() error {
	// Read tag names from struct fields
	getValidator().RegisterTagNameFunc(func(fld reflect.StructField) string {
		return fld.Tag.Get("flagName")
	})

	// Check that the status code is set only if the status page was also set
	if s.CustomNotFoundStatusCode != 0 && s.CustomNotFoundPage == "" {
		return errors.New("unable to set custom 404 status code if no custom page was set")
	}

	// Validate custom 404 page
	if s.CustomNotFoundPage != "" {
		if _, err := os.Stat(s.CustomNotFoundPage); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("custom not found file %q does not exist", s.CustomNotFoundPage)
			}

			return fmt.Errorf("unable to process custom 404 page at %q: %w", s.CustomNotFoundPage, err)
		}
	}

	// Validate custom not found status code if one was set
	if str := http.StatusText(s.CustomNotFoundStatusCode); s.CustomNotFoundStatusCode != 0 && str == "" {
		return fmt.Errorf("unsupported custom not found status code: %d", s.CustomNotFoundStatusCode)
	}

	// Validate max size for ETag
	if s.ETagMaxSize == "" {
		return errors.New("etag max size is required: set it with --etag-max-size")
	}

	// Validate that if custom CSS is set, that it lives in the path where
	// we're serving files
	if s.CustomCSS != "" {
		if !validateIsFileInPath(s.Path, s.CustomCSS) {
			return fmt.Errorf("css file path %q is outside the server's path %q: it must be served from the server itself", s.CustomCSS, s.Path)
		}
	}

	size, err := common.ParseSize(s.ETagMaxSize)
	if err != nil {
		return fmt.Errorf("unable to parse ETag max size: %w", err)
	}

	s.etagMaxSizeBytes = size

	// Validate TLS configuration
	if err := s.validateTLS(); err != nil {
		return err
	}

	// Attempt to validate the structure, and grab the errors
	if err := getValidator().Struct(s); err != nil {
		// If the error isn't empty, and its type is of ValidationError
		// we can provide a better error message for its validation process
		var valerrs validator.ValidationErrors
		if errors.As(err, &valerrs) {
			var merrs MultiError

			for _, e := range valerrs {
				// Convert the error to a human-readable error
				merrs.Append(FieldToValidationError(e))
			}

			return &merrs
		}

		// If the error type is unknown or there's no error
		// return the error as-is
		return fmt.Errorf("unable to validate configuration: %w", err)
	}

	return nil
}

var reIsPathPrefix = regexp.MustCompile(`^/[\w\-\_]+(/[\w\-\_]+)*/$`)

// validateIsPathPrefix checks if the value is a valid path prefix
// which needs to start or end with a forward slash, and include
// within alphanumeric, dashes or underscores, or additional forward slashes
func validateIsPathPrefix(field validator.FieldLevel) bool {
	if field.Field().String() == "/" {
		return true
	}

	return reIsPathPrefix.MatchString(field.Field().String())
}

func validateIsFileInPath(basepath, file string) bool {
	absbasepath, err := filepath.Abs(basepath)
	if err != nil {
		return false
	}

	absfile, err := filepath.Abs(file)
	if err != nil {
		return false
	}

	return strings.HasPrefix(absfile, absbasepath)
}

func (s *Server) validateTLS() error {
	hasCert := s.TLSCert != ""
	hasKey := s.TLSKey != ""
	hasHostname := s.Hostname != ""

	// Both cert+key must be set together
	if hasCert != hasKey {
		return errors.New("both --tls-cert and --tls-key must be provided together")
	}

	// Determine TLS mode:
	//   cert+key+hostname → BYO
	//   hostname alone    → Auto (certmagic)
	//   neither           → Off
	switch {
	case hasCert && hasKey:
		s.activeTLSMode = TLSModeBYO
		if !hasHostname {
			return errors.New("--hostname is required when TLS is active")
		}
	case hasHostname && !hasCert:
		s.activeTLSMode = TLSModeAuto
	default:
		s.activeTLSMode = TLSModeOff
	}

	if s.IsTLSEnabled() {
		// Cannot use --port alongside TLS flags
		if s.PortExplicitlySet {
			return errors.New("cannot use --port with TLS flags; use --http-port and --https-port instead")
		}

		// Ports must differ (unless HTTP is disabled)
		if s.HTTPPort != 0 && s.HTTPPort == s.HTTPSPort {
			return fmt.Errorf("--http-port (%d) and --https-port (%d) must differ", s.HTTPPort, s.HTTPSPort)
		}
	}

	// BYO-specific validation
	if s.activeTLSMode == TLSModeBYO {
		if _, err := os.Stat(s.TLSCert); err != nil {
			return fmt.Errorf("TLS certificate file %q: %w", s.TLSCert, err)
		}
		if _, err := os.Stat(s.TLSKey); err != nil {
			return fmt.Errorf("TLS key file %q: %w", s.TLSKey, err)
		}

		if err := s.loadAndStoreCert(); err != nil {
			return err
		}

		s.setupTLSFileHiding()
	}

	// Auto-specific validation
	if s.activeTLSMode == TLSModeAuto {
		if s.HTTPPort == 0 {
			s.printWarningf("HTTP listener is disabled (--http-port 0). ACME HTTP-01 challenges require port 80 to be reachable. Certificate renewal may fail.")
		} else if s.HTTPPort != 80 {
			s.printWarningf("ACME HTTP-01 challenges require port 80 to be externally reachable. If --http-port %d is not mapped to port 80, certificate provisioning may fail.", s.HTTPPort)
		}
	}

	// If TLS is not active, TLS-specific flags should not be set
	if !s.IsTLSEnabled() {
		if s.HTTPPortExplicitlySet {
			return errors.New("--http-port requires TLS to be active (set --hostname for auto-TLS or --tls-cert/--tls-key for BYO)")
		}
		if s.HTTPSPortExplicitlySet {
			return errors.New("--https-port requires TLS to be active (set --hostname for auto-TLS or --tls-cert/--tls-key for BYO)")
		}
		if s.TLSEmail != "" {
			return errors.New("--tls-email requires TLS to be active (set --hostname for auto-TLS)")
		}
	}

	return nil
}

func (s *Server) setupTLSFileHiding() {
	absPath, err := filepath.Abs(s.Path)
	if err != nil {
		return
	}

	for _, f := range []string{s.TLSCert, s.TLSKey} {
		absFile, err := filepath.Abs(f)
		if err != nil {
			continue
		}
		if strings.HasPrefix(absFile, absPath+string(filepath.Separator)) {
			s.forbiddenAbsPaths = append(s.forbiddenAbsPaths, absFile)
			s.printWarningf("TLS file %q is inside the served directory and will be hidden from listings. Consider moving it outside.", f)
		}
	}
}

func (s *Server) printWarningf(format string, args ...interface{}) {
	if s.LogOutput != nil {
		fmt.Fprintf(s.LogOutput, warnPrefix+format+"\n", args...)
	}
}
