package firewalla

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
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

// newTestServer spins up an httptest.Server, returns a Client pointed at it,
// the mux for handler registration, and a cleanup func.
func newTestServer(t *testing.T) (*Client, *http.ServeMux, func()) {
	t.Helper()
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)

	c, err := NewClient("ignored.example.com", "test-token",
		WithBaseURL(srv.URL+"/v2"),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c, mux, srv.Close
}

type echoBody struct {
	OK bool `json:"ok"`
}

func TestDo_GET_Headers(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()

	mux.HandleFunc("/v2/ping", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("method = %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Token test-token" {
			t.Errorf("Authorization = %q", got)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("Accept = %q", got)
		}
		if got := r.Header.Get("User-Agent"); got == "" {
			t.Errorf("User-Agent missing")
		}
		if r.Header.Get("Content-Type") != "" {
			t.Errorf("Content-Type set on GET: %q", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(200)
		_, _ = io.WriteString(w, `{"ok":true}`)
	})

	var out echoBody
	if err := c.do(context.Background(), "GET", "/ping", nil, nil, &out); err != nil {
		t.Fatalf("do: %v", err)
	}
	if !out.OK {
		t.Errorf("decode failed: %+v", out)
	}
}

func TestDo_POST_BodyAndContentType(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()

	mux.HandleFunc("/v2/echo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s", r.Method)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q", got)
		}
		var got echoBody
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if !got.OK {
			t.Errorf("body OK = false")
		}
		w.WriteHeader(200)
		_, _ = io.WriteString(w, `{"ok":true}`)
	})

	body := echoBody{OK: true}
	var out echoBody
	if err := c.do(context.Background(), "POST", "/echo", nil, body, &out); err != nil {
		t.Fatalf("do: %v", err)
	}
}

func TestDo_QueryEncoding(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()

	mux.HandleFunc("/v2/q", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("query"); got != "status:active foo bar" {
			t.Errorf("query = %q", got)
		}
		if got := r.URL.Query().Get("limit"); got != "5" {
			t.Errorf("limit = %q", got)
		}
		w.WriteHeader(200)
		_, _ = io.WriteString(w, `{"ok":true}`)
	})

	q := url.Values{}
	q.Set("query", "status:active foo bar")
	q.Set("limit", "5")
	var out echoBody
	if err := c.do(context.Background(), "GET", "/q", q, nil, &out); err != nil {
		t.Fatalf("do: %v", err)
	}
}

func TestDo_ErrorMapping(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
		check  func(err error) bool
	}{
		{"401", 401, `{"message":"bad token"}`, IsUnauthorized},
		{"404", 404, `{"message":"not found"}`, IsNotFound},
		{"429", 429, `{"message":"slow down"}`, IsRateLimited},
		{"500", 500, `{"message":"oops"}`, func(err error) bool {
			var ae *APIError
			return errors.As(err, &ae) && ae.HTTPStatus == 500
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cl, mux, teardown := newTestServer(t)
			defer teardown()
			mux.HandleFunc("/v2/x", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(c.status)
				_, _ = io.WriteString(w, c.body)
			})
			err := cl.do(context.Background(), "GET", "/x", nil, nil, nil)
			if err == nil {
				t.Fatalf("expected error")
			}
			if !c.check(err) {
				t.Fatalf("predicate failed for %d: %v", c.status, err)
			}
			var ae *APIError
			if !errors.As(err, &ae) || ae.Body == nil {
				t.Fatalf("APIError not populated: %v", err)
			}
		})
	}
}

func TestDo_ContextCancellation(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()

	mux.HandleFunc("/v2/slow", func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := c.do(ctx, "GET", "/slow", nil, nil, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestDo_RedactsTokenInErrorURL(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()

	mux.HandleFunc("/v2/x", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	})

	err := c.do(context.Background(), "GET", "/x", nil, nil, nil)
	var ae *APIError
	if !errors.As(err, &ae) {
		t.Fatalf("expected APIError, got %v", err)
	}
	if strings.Contains(ae.URL, "test-token") {
		t.Fatalf("URL contains token: %s", ae.URL)
	}
}
