package redirects

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// RedirectRule represents a single redirect rule.
type RedirectRule struct {
	FromPath        string            // The path part of the 'From' pattern
	FromParams      map[string]string // Query parameters with optional placeholders
	To              string            // The 'To' path
	StatusCode      int
	KeepQueryParams bool // Whether to keep original query parameters
}

// Engine holds the parsed redirect rules.
type Engine struct {
	Rules []RedirectRule
}

const colonPlaceholder = "\x00"

// New parses the redirect rules content and returns an Engine instance.
func New(content string) (*Engine, error) {
	rules, err := parseRedirectRules(content)
	if err != nil {
		return nil, err
	}
	return &Engine{Rules: rules}, nil
}

// Middleware is an HTTP middleware that applies redirect rules.
func (e *Engine) Middleware(logger io.Writer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			destination, statusCode, err := e.DereferenceDestination(r.URL.RequestURI())
			if err != nil {
				if errors.Is(err, ErrNoMatchingRule) {
					next.ServeHTTP(w, r)
					return
				}

				http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
				return
			}

			fmt.Fprintf(logger, "REDIR %q -> %q (status: %d)\n", r.URL.RequestURI(), destination, statusCode)
			http.Redirect(w, r, destination, statusCode)
		})
	}
}

var ErrNoMatchingRule = fmt.Errorf("no matching rule")

// DereferenceDestination returns the destination URL and status code for a given request URI.
func (e *Engine) DereferenceDestination(requestURI string) (string, int, error) {
	u, err := url.ParseRequestURI(requestURI)
	if err != nil {
		return "", 0, err
	}

	for _, rule := range e.Rules {
		// Copy of request query parameters to avoid modifying the original
		requestQueryParams := u.RawQuery

		if params, ok := rule.Match(u.Path, requestQueryParams); ok {
			destination := rule.buildDestination(params, requestQueryParams, rule.KeepQueryParams)
			return destination, rule.StatusCode, nil
		}
	}
	return "", 0, ErrNoMatchingRule
}

// parseRedirectRules parses the redirect file content into a slice of RedirectRule structs.
func parseRedirectRules(content string) ([]RedirectRule, error) {
	lines := strings.Split(content, "\n")
	var rules []RedirectRule

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Remove comments from the line
		if idx := strings.Index(line, "#"); idx != -1 {
			line = line[:idx]
		}

		// Split the line into parts
		parts := strings.Fields(line)
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid redirect rule on line %d: %q", lineNum+1, line)
		}

		from, to, statusStr := parts[0], parts[1], parts[2]

		statusCode, err := parseStatusCode(statusStr)
		if err != nil {
			return nil, fmt.Errorf("invalid status code on line %d: %v", lineNum+1, err)
		}

		keepQueryParams := false
		// Check if the 'From' path ends with '?!'
		if strings.HasSuffix(from, "?!") {
			keepQueryParams = true
			from = strings.TrimSuffix(from, "?!")
		}

		// Separate path and query parameters in the 'From' pattern
		var fromPath string
		fromParams := make(map[string]string)

		if idx := strings.Index(from, "?"); idx != -1 {
			fromPath = from[:idx]
			queryStr := from[idx+1:]
			fromParams, err = parseQueryParameters(queryStr)
			if err != nil {
				return nil, fmt.Errorf("invalid query parameters on line %d: %v", lineNum+1, err)
			}
		} else {
			fromPath = from
		}

		// Unescape colons in the 'FromPath' pattern
		fromPath = unescapeColons(fromPath)

		// Validate the 'FromPath' pattern
		if err := validateFromPathPattern(fromPath, lineNum+1); err != nil {
			return nil, err
		}

		rule := RedirectRule{
			FromPath:        fromPath,
			FromParams:      fromParams,
			To:              to,
			StatusCode:      statusCode,
			KeepQueryParams: keepQueryParams,
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

// unescapeColons replaces escaped colons "\:" with colonPlaceholder.
func unescapeColons(s string) string {
	s = strings.ReplaceAll(s, `\:`, colonPlaceholder)
	s = strings.ReplaceAll(s, `\\`, `\`)
	return s
}

// parseStatusCode parses 'temporary' or 'permanent' to HTTP status codes.
func parseStatusCode(s string) (int, error) {
	switch strings.ToLower(s) {
	case "permanent":
		return http.StatusMovedPermanently, nil // 301
	case "temporary":
		return http.StatusFound, nil // 302
	default:
		return 0, fmt.Errorf("unsupported redirection status: %q", s)
	}
}

// parseQueryParameters parses query parameters into a map.
func parseQueryParameters(queryStr string) (map[string]string, error) {
	params := make(map[string]string)
	pairs := strings.Split(queryStr, "&")
	for _, pair := range pairs {
		if pair == "" {
			continue
		}
		keyValue := strings.SplitN(pair, "=", 2)
		key, value := "", ""
		key = keyValue[0]
		if len(keyValue) > 1 {
			value = keyValue[1]
		}
		// Unescape colons in key and value
		key = unescapeColons(key)
		value = unescapeColons(value)
		params[key] = value
	}
	return params, nil
}

// Match checks if the request path and query parameters match the rule.
func (rule *RedirectRule) Match(requestPath string, requestRawQuery string) (map[string]string, bool) {
	params := make(map[string]string)

	// Match the path
	if !pathMatch(rule.FromPath, requestPath, params) {
		return nil, false
	}

	// Parse request query parameters
	requestQueryParams, _ := url.ParseQuery(requestRawQuery)

	// Match query parameters
	if !queryParamsMatch(rule.FromParams, requestQueryParams, params) {
		return nil, false
	}

	return params, true
}

// pathMatch checks if the request path matches the 'From' pattern.
func pathMatch(pattern, path string, params map[string]string) bool {
	// Split the pattern and the path into segments
	patternSegments := splitPathSegments(pattern)
	pathSegments := splitPathSegments(path)

	i, j := 0, 0
	for i < len(patternSegments) {
		if j >= len(pathSegments) {
			// Not enough segments in the path
			return false
		}

		patternSegment := patternSegments[i]
		pathSegment := pathSegments[j]

		// Replace colonPlaceholder back to ':'
		patternSegment = strings.ReplaceAll(patternSegment, colonPlaceholder, ":")
		pathSegment = strings.ReplaceAll(pathSegment, colonPlaceholder, ":")

		if patternSegment == "*" {
			// Splat matches the rest of the path
			params["splat"] = strings.Join(pathSegments[j:], "/")
			return true
		} else if strings.HasPrefix(patternSegment, ":") {
			paramName := patternSegment[1:]
			if paramName == "splat" && i == len(patternSegments)-1 {
				// ':splat' at the end captures the rest of the path
				params["splat"] = strings.Join(pathSegments[j:], "/")
				return true
			} else if paramName == "splat" {
				// ':splat' used not at the end (should have been caught during parsing)
				return false
			} else {
				// Regular placeholder
				params[paramName] = pathSegment
			}
		} else if patternSegment == pathSegment {
			// Exact match
		} else {
			// No match
			return false
		}
		i++
		j++
	}

	// Check if all path segments have been matched
	return j == len(pathSegments)
}

// splitPathSegments splits a path into segments, handling escaped characters.
func splitPathSegments(path string) []string {
	segments := []string{}
	var segment strings.Builder
	escaped := false

	for i := 0; i < len(path); i++ {
		c := path[i]
		if escaped {
			segment.WriteByte(c)
			escaped = false
		} else if c == '\\' {
			escaped = true
		} else if c == '/' {
			segments = append(segments, segment.String())
			segment.Reset()
		} else {
			segment.WriteByte(c)
		}
	}
	segments = append(segments, segment.String())
	return segments
}

// queryParamsMatch checks if request query parameters match the rule's query parameters.
func queryParamsMatch(ruleParams map[string]string, requestQueryParams url.Values, params map[string]string) bool {
	for key, ruleValue := range ruleParams {
		// Replace colonPlaceholder back to ':'
		key = strings.ReplaceAll(key, colonPlaceholder, ":")
		ruleValue = strings.ReplaceAll(ruleValue, colonPlaceholder, ":")

		requestValues, ok := requestQueryParams[key]
		if !ok || len(requestValues) == 0 {
			// Query parameter not present in request
			return false
		}
		requestValue := requestValues[0] // Only consider the first value
		// Remove the key from requestQueryParams to mark it as consumed
		delete(requestQueryParams, key)

		if strings.HasPrefix(ruleValue, ":") {
			// Placeholder, extract value
			paramName := ruleValue[1:]
			params[paramName] = requestValue
		} else {
			// Literal value, compare
			if requestValue != ruleValue {
				return false
			}
		}
	}
	return true
}

func (rule *RedirectRule) buildDestination(params map[string]string, requestRawQuery string, keepQueryParams bool) string {
	// Replace placeholders in the 'To' path
	destination := rule.To

	// Replace :splat
	if splatValue, ok := params["splat"]; ok {
		destination = strings.ReplaceAll(destination, "*", splatValue)
		destination = strings.ReplaceAll(destination, ":splat", splatValue)
	}

	// Replace other placeholders
	for key, value := range params {
		if key == "splat" {
			continue
		}
		placeholder := ":" + key
		destination = strings.ReplaceAll(destination, placeholder, value)
	}

	// Replace colon placeholders back to colons
	destination = strings.ReplaceAll(destination, colonPlaceholder, ":")

	if keepQueryParams {
		// Parse the destination URL to extract existing query parameters
		destURL, err := url.Parse(destination)
		if err != nil {
			// If parsing fails, treat the entire destination as the path
			destURL = &url.URL{Path: destination}
		}

		// Extract query parameters from destination URL, preserving order
		destQueryParams := parseQueryParamsPreserveOrder(destURL.RawQuery)

		// Collect parameter names from destination path and query parameters
		destParamNames := extractParamNamesFromDestination(rule, destQueryParams)

		// Parse request query parameters, preserving order
		requestQueryParams := parseQueryParamsPreserveOrder(requestRawQuery)

		// Collect request query parameters that are not in destParamNames
		additionalQueryParams := []QueryParam{}
		for _, qp := range requestQueryParams {
			if !destParamNames[qp.Key] {
				additionalQueryParams = append(additionalQueryParams, qp)
			}
		}

		// Build final query string
		allQueryParams := append(destQueryParams, additionalQueryParams...)
		queryString := encodeQueryParams(allQueryParams)

		// Rebuild destination URL
		destURL.RawQuery = queryString
		destination = destURL.String()
	}

	return destination
}

// extractParamNamesFromDestination extracts parameter names from the destination path and query parameters.
func extractParamNamesFromDestination(rule *RedirectRule, destQueryParams []QueryParam) map[string]bool {
	destParamNames := make(map[string]bool)

	// Collect parameter names used in placeholders (parameters in the 'To' pattern)
	for key := range rule.FromParams {
		// Replace colonPlaceholder back to ':'
		key = strings.ReplaceAll(key, colonPlaceholder, ":")
		destParamNames[key] = true
	}
	for key := range rule.FromPathParams() {
		destParamNames[key] = true
	}

	// Also collect parameter names from destination query parameters
	for _, qp := range destQueryParams {
		destParamNames[qp.Key] = true
	}

	return destParamNames
}

// FromPathParams extracts parameter names from the 'FromPath' pattern.
func (rule *RedirectRule) FromPathParams() map[string]bool {
	params := make(map[string]bool)
	patternSegments := splitPathSegments(rule.FromPath)
	for _, segment := range patternSegments {
		if strings.HasPrefix(segment, ":") {
			paramName := segment[1:]
			params[paramName] = true
		}
	}
	return params
}

// QueryParam represents a single query parameter, preserving order.
type QueryParam struct {
	Key   string
	Value string
}

// parseQueryParamsPreserveOrder parses query parameters, preserving their order.
func parseQueryParamsPreserveOrder(rawQuery string) []QueryParam {
	var params []QueryParam
	if rawQuery == "" {
		return params
	}
	pairs := strings.Split(rawQuery, "&")
	for _, pair := range pairs {
		if pair == "" {
			continue
		}
		keyValue := strings.SplitN(pair, "=", 2)
		key, value := "", ""
		key = keyValue[0]
		if len(keyValue) > 1 {
			value = keyValue[1]
		}
		// Decode key and value
		decodedKey, err := url.QueryUnescape(key)
		if err != nil {
			decodedKey = key
		}
		decodedValue, err := url.QueryUnescape(value)
		if err != nil {
			decodedValue = value
		}
		params = append(params, QueryParam{Key: decodedKey, Value: decodedValue})
	}
	return params
}

// encodeQueryParams encodes query parameters from a slice of QueryParam, preserving order.
func encodeQueryParams(params []QueryParam) string {
	var parts []string
	for _, qp := range params {
		encodedKey := url.QueryEscape(qp.Key)
		if qp.Value == "" {
			// Handle parameters with empty values
			parts = append(parts, encodedKey+"=")
		} else {
			encodedValue := url.QueryEscape(qp.Value)
			parts = append(parts, fmt.Sprintf("%s=%s", encodedKey, encodedValue))
		}
	}
	return strings.Join(parts, "&")
}

// validateFromPathPattern checks that 'splat' is only used at the end of the pattern.
func validateFromPathPattern(fromPath string, lineNum int) error {
	patternSegments := splitPathSegments(fromPath)
	for i, segment := range patternSegments {
		// Do not replace colonPlaceholder back to ':'

		if segment == "*" {
			if i != len(patternSegments)-1 {
				return fmt.Errorf("invalid use of wildcard on line %d: \"*\" can only be used at the end of a path", lineNum)
			}
		}
		if strings.HasPrefix(segment, ":") {
			paramName := segment[1:]
			if paramName == "splat" {
				if i != len(patternSegments)-1 {
					return fmt.Errorf("invalid use of \":splat\" on line %d: \":splat\" can only be used at the end of a path", lineNum)
				}
			}
		}
		// Check for unescaped colons in the middle of the segment
		startIdx := 1
		if !strings.HasPrefix(segment, ":") {
			startIdx = 0
		}
		for idx := startIdx; idx < len(segment); idx++ {
			c := segment[idx]
			if c == ':' {
				return fmt.Errorf("invalid use of \":\" in segment \"%s\" on line %d: \":\" can only be used at the beginning of a path section", strings.ReplaceAll(segment, colonPlaceholder, ":"), lineNum+1)
			}
		}
	}
	return nil
}
