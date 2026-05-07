// Package retry provides an opt-in http.RoundTripper that retries on transient
// failures (network errors, 429, 502, 503, 504), honors the Retry-After header,
// and re-reads the request body via http.Request.GetBody.
package retry

import (
	"io"
	"math"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"
)

type Config struct {
	MaxAttempts int           // total attempts, including the first try; default 4
	BaseDelay   time.Duration // initial backoff; default 500ms
	MaxDelay    time.Duration // backoff cap; default 30s
	Jitter      float64       // 0..1; default 0.2
}

func (c Config) withDefaults() Config {
	if c.MaxAttempts <= 0 {
		c.MaxAttempts = 4
	}
	if c.BaseDelay <= 0 {
		c.BaseDelay = 500 * time.Millisecond
	}
	if c.MaxDelay <= 0 {
		c.MaxDelay = 30 * time.Second
	}
	if c.Jitter < 0 {
		c.Jitter = 0
	}
	if c.Jitter > 1 {
		c.Jitter = 1
	}
	return c
}

type transport struct {
	base http.RoundTripper
	cfg  Config
}

func NewTransport(base http.RoundTripper, cfg Config) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	return &transport{base: base, cfg: cfg.withDefaults()}
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	canRetryBody := req.Body == nil || req.GetBody != nil

	var resp *http.Response
	var err error

	for attempt := 0; attempt < t.cfg.MaxAttempts; attempt++ {
		if attempt > 0 && req.Body != nil && req.GetBody != nil {
			body, gerr := req.GetBody()
			if gerr != nil {
				return nil, gerr
			}
			req.Body = body
		}
		resp, err = t.base.RoundTrip(req)
		if !shouldRetry(resp, err, canRetryBody) {
			return resp, err
		}
		if resp != nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}
		if attempt == t.cfg.MaxAttempts-1 {
			break
		}
		delay := t.computeDelay(attempt, resp)
		select {
		case <-time.After(delay):
		case <-req.Context().Done():
			return nil, req.Context().Err()
		}
	}
	return resp, err
}

func shouldRetry(resp *http.Response, err error, canRetryBody bool) bool {
	if !canRetryBody {
		return false
	}
	if err != nil {
		return true
	}
	switch resp.StatusCode {
	case 429, 502, 503, 504:
		return true
	}
	return false
}

func (t *transport) computeDelay(attempt int, resp *http.Response) time.Duration {
	if resp != nil {
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if d, ok := parseRetryAfter(ra); ok {
				if d > t.cfg.MaxDelay {
					d = t.cfg.MaxDelay
				}
				return d
			}
		}
	}
	d := time.Duration(math.Min(
		float64(t.cfg.MaxDelay),
		float64(t.cfg.BaseDelay)*math.Pow(2, float64(attempt)),
	))
	if t.cfg.Jitter > 0 {
		factor := 1.0 + (rand.Float64()*2-1)*t.cfg.Jitter
		d = time.Duration(float64(d) * factor)
		if d < 0 {
			d = 0
		}
	}
	return d
}

func parseRetryAfter(v string) (time.Duration, bool) {
	if secs, err := strconv.Atoi(v); err == nil {
		if secs < 0 {
			secs = 0
		}
		return time.Duration(secs) * time.Second, true
	}
	if t, err := http.ParseTime(v); err == nil {
		d := time.Until(t)
		if d < 0 {
			d = 0
		}
		return d, true
	}
	return 0, false
}
