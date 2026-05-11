package firewalla

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const Version = "0.2.1"

type Client struct {
	domain     string
	token      string
	baseURL    *url.URL
	httpClient *http.Client
	userAgent  string

	Boxes       *BoxesService
	Devices     *DevicesService
	Alarms      *AlarmsService
	Flows       *FlowsService
	Rules       *RulesService
	TargetLists *TargetListsService
	Stats       *StatsService
	Trends      *TrendsService
}

type Option func(*Client) error

func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) error {
		if h == nil {
			return errors.New("firewalla: WithHTTPClient: nil http.Client")
		}
		c.httpClient = h
		return nil
	}
}

func WithBaseURL(raw string) Option {
	return func(c *Client) error {
		u, err := url.Parse(raw)
		if err != nil {
			return fmt.Errorf("firewalla: WithBaseURL: %w", err)
		}
		c.baseURL = u
		return nil
	}
}

func WithUserAgent(ua string) Option {
	return func(c *Client) error {
		if ua == "" {
			return errors.New("firewalla: WithUserAgent: empty user agent")
		}
		c.userAgent = ua
		return nil
	}
}

func NewClient(domain, token string, opts ...Option) (*Client, error) {
	if domain == "" {
		return nil, errors.New("firewalla: domain is required")
	}
	if token == "" {
		return nil, errors.New("firewalla: token is required")
	}
	if strings.Contains(domain, "://") {
		return nil, fmt.Errorf("firewalla: domain must not contain protocol prefix: %q", domain)
	}
	if strings.HasSuffix(domain, "/") {
		return nil, fmt.Errorf("firewalla: domain must not have trailing slash: %q", domain)
	}

	base, err := url.Parse("https://" + domain + "/v2")
	if err != nil {
		return nil, fmt.Errorf("firewalla: invalid domain: %w", err)
	}

	c := &Client{
		domain:     domain,
		token:      token,
		baseURL:    base,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		userAgent:  "firewalla-msp-go/" + Version,
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	c.Boxes = &BoxesService{client: c}
	c.Devices = &DevicesService{client: c}
	c.Alarms = &AlarmsService{client: c}
	c.Flows = &FlowsService{client: c}
	c.Rules = &RulesService{client: c}
	c.TargetLists = &TargetListsService{client: c}
	c.Stats = &StatsService{client: c}
	c.Trends = &TrendsService{client: c}
	return c, nil
}

// do executes an HTTP request against the MSP API.
//
// path is appended to baseURL (must start with "/").
// query, body, and out are all optional. If body is non-nil it is JSON-encoded.
// On non-2xx, returns *APIError. On 2xx with out != nil, decodes JSON into out.
func (c *Client) do(ctx context.Context, method, path string, query url.Values, body any, out any) error {
	u := *c.baseURL
	u.Path += path
	if query != nil {
		u.RawQuery = query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("firewalla: marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return fmt.Errorf("firewalla: build request: %w", err)
	}
	req.Header.Set("Authorization", "Token "+c.token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		ae := &APIError{
			HTTPStatus: resp.StatusCode,
			Body:       raw,
			Method:     method,
			URL:        u.Path,
			RequestID:  resp.Header.Get("X-Request-Id"),
		}
		var generic struct {
			Message string `json:"message"`
			Code    string `json:"code"`
			Error   string `json:"error"`
		}
		if json.Unmarshal(raw, &generic) == nil {
			ae.Code = generic.Code
			ae.Message = generic.Message
			if ae.Message == "" {
				ae.Message = generic.Error
			}
		}
		return ae
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("firewalla: decode response: %w", err)
		}
	}
	return nil
}

// Service placeholders. Methods and fields are added in their own files.
type BoxesService struct{ client *Client }
type DevicesService struct{ client *Client }
type AlarmsService struct{ client *Client }
type FlowsService struct{ client *Client }
type RulesService struct{ client *Client }
type TargetListsService struct{ client *Client }
type StatsService struct{ client *Client }
type TrendsService struct{ client *Client }
