package retry

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func newTestTransport(cfg Config) http.RoundTripper {
	if cfg.BaseDelay == 0 {
		cfg.BaseDelay = 1 * time.Millisecond
	}
	return NewTransport(http.DefaultTransport, cfg)
}

func TestRetry_RetriesOn5xxThenSucceeds(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			w.WriteHeader(503)
			return
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	c := &http.Client{Transport: newTestTransport(Config{MaxAttempts: 4})}
	resp, err := c.Get(srv.URL)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 || atomic.LoadInt32(&calls) != 3 {
		t.Errorf("status=%d calls=%d", resp.StatusCode, calls)
	}
}

func TestRetry_DoesNotRetryOn4xx(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(403)
	}))
	defer srv.Close()

	c := &http.Client{Transport: newTestTransport(Config{MaxAttempts: 4})}
	resp, _ := c.Get(srv.URL)
	resp.Body.Close()
	if calls != 1 {
		t.Errorf("calls = %d, want 1", calls)
	}
}

func TestRetry_RetriesOn429(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 2 {
			w.WriteHeader(429)
			return
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	c := &http.Client{Transport: newTestTransport(Config{MaxAttempts: 4})}
	resp, _ := c.Get(srv.URL)
	resp.Body.Close()
	if calls != 2 {
		t.Errorf("calls = %d", calls)
	}
}

func TestRetry_RespectsRetryAfter(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(429)
			return
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	start := time.Now()
	c := &http.Client{Transport: newTestTransport(Config{MaxAttempts: 4, BaseDelay: 5 * time.Second})}
	resp, _ := c.Get(srv.URL)
	resp.Body.Close()
	if time.Since(start) > 1*time.Second {
		t.Errorf("Retry-After=0 should bypass BaseDelay; took %v", time.Since(start))
	}
	if calls != 2 {
		t.Errorf("calls = %d", calls)
	}
}

func TestRetry_PostRetryablyReadsBody(t *testing.T) {
	var calls int32
	var lastBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		body, _ := io.ReadAll(r.Body)
		lastBody = string(body)
		if n < 2 {
			w.WriteHeader(503)
			return
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	c := &http.Client{Transport: newTestTransport(Config{MaxAttempts: 4})}
	body := strings.NewReader(`{"k":"v"}`)
	req, _ := http.NewRequest("POST", srv.URL, body)
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	resp.Body.Close()
	if calls != 2 {
		t.Errorf("calls = %d", calls)
	}
	if lastBody != `{"k":"v"}` {
		t.Errorf("retried body = %q", lastBody)
	}
}

func TestRetry_NoRetryWhenGetBodyMissing(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(503)
	}))
	defer srv.Close()

	req, _ := http.NewRequest("POST", srv.URL, bytes.NewReader([]byte(`{"k":"v"}`)))
	req.GetBody = nil
	c := &http.Client{Transport: newTestTransport(Config{MaxAttempts: 4})}
	resp, _ := c.Do(req)
	resp.Body.Close()
	if calls != 1 {
		t.Errorf("calls = %d, want 1", calls)
	}
}

func TestRetry_GivesUpAfterMaxAttempts(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(502)
	}))
	defer srv.Close()

	c := &http.Client{Transport: newTestTransport(Config{MaxAttempts: 3})}
	resp, _ := c.Get(srv.URL)
	resp.Body.Close()
	if calls != 3 {
		t.Errorf("calls = %d, want 3", calls)
	}
}

type errBaseRoundTripper struct {
	calls *int32
	inner http.RoundTripper
}

func (e *errBaseRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	n := atomic.AddInt32(e.calls, 1)
	if n == 1 {
		return nil, errSimulatedNetwork
	}
	return e.inner.RoundTrip(req)
}

var errSimulatedNetwork = errors.New("simulated network error")

func TestRetry_NetworkErrorRetried(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()

	var calls int32
	base := &errBaseRoundTripper{calls: &calls, inner: http.DefaultTransport}
	tr := NewTransport(base, Config{MaxAttempts: 4, BaseDelay: 1 * time.Millisecond})
	c := &http.Client{Transport: tr}

	resp, err := c.Get(srv.URL)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	resp.Body.Close()
	if calls != 2 {
		t.Errorf("calls = %d, want 2 (1 failure + 1 success)", calls)
	}
}
