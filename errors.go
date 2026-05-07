package firewalla

import (
	"errors"
	"fmt"
)

// APIError is returned for any non-2xx response from the MSP API.
type APIError struct {
	HTTPStatus int
	Code       string
	Message    string
	RequestID  string
	Body       []byte
	Method     string
	URL        string
}

func (e *APIError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("firewalla: %s %s: %d", e.Method, e.URL, e.HTTPStatus)
	}
	return fmt.Sprintf("firewalla: %s %s: %d %s", e.Method, e.URL, e.HTTPStatus, e.Message)
}

func httpStatusIs(err error, status int) bool {
	var ae *APIError
	if errors.As(err, &ae) {
		return ae.HTTPStatus == status
	}
	return false
}

// IsNotFound reports whether err is an APIError with HTTP 404.
func IsNotFound(err error) bool { return httpStatusIs(err, 404) }

// IsUnauthorized reports whether err is an APIError with HTTP 401.
func IsUnauthorized(err error) bool { return httpStatusIs(err, 401) }

// IsRateLimited reports whether err is an APIError with HTTP 429.
func IsRateLimited(err error) bool { return httpStatusIs(err, 429) }
