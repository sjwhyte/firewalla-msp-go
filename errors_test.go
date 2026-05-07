package firewalla

import (
	"errors"
	"fmt"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	e := &APIError{
		HTTPStatus: 404,
		Method:     "GET",
		URL:        "/v2/alarms/abc/1",
		Message:    "alarm not found",
	}
	got := e.Error()
	want := "firewalla: GET /v2/alarms/abc/1: 404 alarm not found"
	if got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

func TestAPIError_Error_NoMessage(t *testing.T) {
	e := &APIError{HTTPStatus: 500, Method: "GET", URL: "/v2/foo"}
	got := e.Error()
	want := "firewalla: GET /v2/foo: 500"
	if got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

func TestPredicates(t *testing.T) {
	cases := []struct {
		name    string
		err     error
		want404 bool
		want401 bool
		want429 bool
	}{
		{"nil", nil, false, false, false},
		{"plain", errors.New("boom"), false, false, false},
		{"404", &APIError{HTTPStatus: 404}, true, false, false},
		{"401", &APIError{HTTPStatus: 401}, false, true, false},
		{"429", &APIError{HTTPStatus: 429}, false, false, true},
		{"wrapped 404", fmt.Errorf("ctx: %w", &APIError{HTTPStatus: 404}), true, false, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := IsNotFound(c.err); got != c.want404 {
				t.Errorf("IsNotFound = %v, want %v", got, c.want404)
			}
			if got := IsUnauthorized(c.err); got != c.want401 {
				t.Errorf("IsUnauthorized = %v, want %v", got, c.want401)
			}
			if got := IsRateLimited(c.err); got != c.want429 {
				t.Errorf("IsRateLimited = %v, want %v", got, c.want429)
			}
		})
	}
}
