package firewalla

import (
	"context"
	"errors"
	"net/url"
)

// TargetList is a reusable collection of domains/IPs used in rules.
type TargetList struct {
	ID          string   `json:"id"`
	Owner       string   `json:"owner"`
	Name        string   `json:"name"`
	Category    string   `json:"category,omitempty"`
	Targets     []string `json:"targets"`
	Notes       string   `json:"notes,omitempty"`
	LastUpdated int64    `json:"lastUpdated,omitempty"`
}

type TargetListCreate struct {
	Name     string   `json:"name"`
	Targets  []string `json:"targets"`
	Category string   `json:"category,omitempty"`
	Notes    string   `json:"notes,omitempty"`
	Owner    string   `json:"owner,omitempty"`
}

type TargetListUpdate struct {
	Name     *string   `json:"name,omitempty"`
	Targets  *[]string `json:"targets,omitempty"`
	Category *string   `json:"category,omitempty"`
	Notes    *string   `json:"notes,omitempty"`
}

func (s *TargetListsService) List(ctx context.Context) ([]TargetList, error) {
	var out []TargetList
	if err := s.client.do(ctx, "GET", "/target-lists", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *TargetListsService) Get(ctx context.Context, id string) (*TargetList, error) {
	if id == "" {
		return nil, errors.New("firewalla: target list id is required")
	}
	var out TargetList
	if err := s.client.do(ctx, "GET", "/target-lists/"+url.PathEscape(id), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *TargetListsService) Create(ctx context.Context, body *TargetListCreate) (*TargetList, error) {
	if body == nil || body.Name == "" {
		return nil, errors.New("firewalla: target list Name is required")
	}
	var out TargetList
	if err := s.client.do(ctx, "POST", "/target-lists", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *TargetListsService) Update(ctx context.Context, id string, body *TargetListUpdate) (*TargetList, error) {
	if id == "" {
		return nil, errors.New("firewalla: target list id is required")
	}
	if body == nil {
		return nil, errors.New("firewalla: update body is required")
	}
	var out TargetList
	if err := s.client.do(ctx, "PATCH", "/target-lists/"+url.PathEscape(id), nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *TargetListsService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("firewalla: target list id is required")
	}
	return s.client.do(ctx, "DELETE", "/target-lists/"+url.PathEscape(id), nil, nil, nil)
}
