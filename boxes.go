package firewalla

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
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
	LastSeen    Timestamp `json:"lastSeen"`
	DeviceCount int       `json:"deviceCount"`
	RuleCount   int       `json:"ruleCount"`
	AlarmCount  int       `json:"alarmCount"`
	Group       *BoxGroup `json:"group,omitempty"`
}

// BoxGroup identifies an MSP box group. The MSP API returns this as either a
// JSON string (treated as the group id) or a JSON object with {id, name}.
type BoxGroup struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

func (g *BoxGroup) UnmarshalJSON(data []byte) error {
	return unmarshalIDName(data, &g.ID, &g.Name)
}

// unmarshalIDName decodes either a JSON string (stored in id) or a JSON object
// with {id, name} fields into the given pointers. Shared by BoxGroup and
// DeviceGroup; the MSP API has been observed to return either form.
func unmarshalIDName(data []byte, id, name *string) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	if data[0] == '"' {
		return json.Unmarshal(data, id)
	}
	var obj struct {
		ID   any    `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	*name = obj.Name
	switch v := obj.ID.(type) {
	case string:
		*id = v
	case float64:
		s := strconv.FormatFloat(v, 'f', -1, 64)
		*id = strings.TrimSuffix(s, ".0")
	case nil:
		// leave id empty
	default:
		*id = fmt.Sprintf("%v", v)
	}
	return nil
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
