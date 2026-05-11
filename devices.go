package firewalla

import (
	"context"
	"errors"
	"net/url"
)

// Device represents a network device tracked by a Firewalla box.
type Device struct {
	ID            string       `json:"id"`
	GID           string       `json:"gid"`
	Name          string       `json:"name"`
	IP            string       `json:"ip"`
	IPReserved    bool         `json:"ipReserved"`
	MACVendor     string       `json:"macVendor"`
	Online        bool         `json:"online"`
	LastSeen      Timestamp    `json:"lastSeen"`
	Network       *NetworkRef  `json:"network,omitempty"`
	Group         *DeviceGroup `json:"group,omitempty"`
	TotalDownload int64        `json:"totalDownload"`
	TotalUpload   int64        `json:"totalUpload"`
}

type NetworkRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// DeviceGroup identifies a device group. The MSP API has been observed to
// return this as either a JSON string (the group id) or a JSON object with
// {id, name}; DeviceGroup.UnmarshalJSON handles both.
type DeviceGroup struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

func (g *DeviceGroup) UnmarshalJSON(data []byte) error {
	return unmarshalIDName(data, &g.ID, &g.Name)
}

type DeviceListOptions struct {
	Box   string
	Group string
}

type DeviceUpdate struct {
	Name *string `json:"name,omitempty"`
}

func (s *DevicesService) List(ctx context.Context, opts *DeviceListOptions) ([]Device, error) {
	q := url.Values{}
	if opts != nil {
		if opts.Box != "" && opts.Group != "" {
			return nil, errors.New("firewalla: DeviceListOptions.Box and Group are mutually exclusive")
		}
		if opts.Box != "" {
			q.Set("box", opts.Box)
		}
		if opts.Group != "" {
			q.Set("group", opts.Group)
		}
	}
	var out []Device
	if err := s.client.do(ctx, "GET", "/devices", q, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *DevicesService) Update(ctx context.Context, gid, deviceID string, patch *DeviceUpdate) (*Device, error) {
	if gid == "" || deviceID == "" {
		return nil, errors.New("firewalla: gid and deviceID are required")
	}
	if patch == nil {
		return nil, errors.New("firewalla: patch is required")
	}
	var out Device
	path := "/boxes/" + url.PathEscape(gid) + "/devices/" + url.PathEscape(deviceID)
	if err := s.client.do(ctx, "PATCH", path, nil, patch, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
