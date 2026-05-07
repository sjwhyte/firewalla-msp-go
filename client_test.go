package firewalla

import (
	"net/http"
	"strings"
	"testing"
)

func TestNewClient_Validation(t *testing.T) {
	cases := []struct {
		name       string
		domain     string
		token      string
		wantSubstr string
	}{
		{"empty domain", "", "tok", "domain"},
		{"empty token", "dom.firewalla.net", "", "token"},
		{"protocol prefix https", "https://dom.firewalla.net", "tok", "protocol"},
		{"protocol prefix http", "http://dom.firewalla.net", "tok", "protocol"},
		{"trailing slash", "dom.firewalla.net/", "tok", "trailing"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := NewClient(c.domain, c.token)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", c.wantSubstr)
			}
			if !strings.Contains(err.Error(), c.wantSubstr) {
				t.Fatalf("error %q does not contain %q", err.Error(), c.wantSubstr)
			}
		})
	}
}

func TestNewClient_Defaults(t *testing.T) {
	c, err := NewClient("dom.firewalla.net", "tok")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c.baseURL.String() != "https://dom.firewalla.net/v2" {
		t.Errorf("baseURL = %s", c.baseURL.String())
	}
	if c.httpClient == nil {
		t.Errorf("httpClient is nil")
	}
	if !strings.HasPrefix(c.userAgent, "firewalla-msp-go/") {
		t.Errorf("userAgent = %q", c.userAgent)
	}
	if c.Boxes == nil || c.Devices == nil || c.Alarms == nil || c.Flows == nil ||
		c.Rules == nil || c.TargetLists == nil || c.Stats == nil || c.Trends == nil {
		t.Errorf("sub-service is nil")
	}
}

func TestNewClient_Options(t *testing.T) {
	hc := &http.Client{}
	c, err := NewClient("dom.firewalla.net", "tok",
		WithHTTPClient(hc),
		WithBaseURL("https://override.example/v2"),
		WithUserAgent("my-ua/1.0"),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c.httpClient != hc {
		t.Error("WithHTTPClient not applied")
	}
	if c.baseURL.String() != "https://override.example/v2" {
		t.Errorf("WithBaseURL not applied: %s", c.baseURL.String())
	}
	if c.userAgent != "my-ua/1.0" {
		t.Errorf("WithUserAgent not applied: %s", c.userAgent)
	}
}
