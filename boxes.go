package firewalla

import (
	"context"
	"net/url"
)

// Box represents a Firewalla box registered to the MSP.
type Box struct {
	GID         string    `json:"gid"`
	Name        string    `json:"name"`
	Model       string    `json:"model"`
	Mode        string    `json:"mode"`
	Online      bool      `json:"online"`
	Version     string    `json:"version"`
	PublicIP    string    `json:"publicIP"`
	Location    string    `json:"location"`
	LastSeen    int64     `json:"lastSeen"`
	DeviceCount int       `json:"deviceCount"`
	RuleCount   int       `json:"ruleCount"`
	AlarmCount  int       `json:"alarmCount"`
	Group       *BoxGroup `json:"group,omitempty"`
}

type BoxGroup struct {
	ID   any    `json:"id"`
	Name string `json:"name"`
}

type BoxListOptions struct {
	Group string
}

func (s *BoxesService) List(ctx context.Context, opts *BoxListOptions) ([]Box, error) {
	q := url.Values{}
	if opts != nil && opts.Group != "" {
		q.Set("group", opts.Group)
	}
	var out []Box
	if err := s.client.do(ctx, "GET", "/boxes", q, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
