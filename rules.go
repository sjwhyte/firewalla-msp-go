package firewalla

import (
	"context"
	"encoding/json"
	"errors"
	"iter"
	"net/url"
	"strconv"
)

type Rule struct {
	ID        string          `json:"id"`
	Action    string          `json:"action"`
	Direction string          `json:"direction,omitempty"`
	GID       string          `json:"gid,omitempty"`
	Target    *RuleTarget     `json:"target,omitempty"`
	Scope     *RuleScope      `json:"scope,omitempty"`
	Status    string          `json:"status,omitempty"`
	Schedule  json.RawMessage `json:"schedule,omitempty"`
	Notes     string          `json:"notes,omitempty"`
	Protocol  string          `json:"protocol,omitempty"`
	CreatedAt Timestamp       `json:"createdAt,omitempty"`
	UpdatedAt Timestamp       `json:"updatedAt,omitempty"`
}

type RuleTarget struct {
	Type    string `json:"type"`
	Value   string `json:"value"`
	DNSOnly bool   `json:"dnsOnly,omitempty"`
}
type RuleScope struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type RuleListOptions struct {
	Query   string
	GroupBy string
	SortBy  string
	Limit   int
	Cursor  string
}

type RuleCreate struct {
	Action    string          `json:"action"`
	Direction string          `json:"direction,omitempty"`
	GID       string          `json:"gid,omitempty"`
	Target    *RuleTarget     `json:"target"`
	Scope     *RuleScope      `json:"scope,omitempty"`
	Protocol  string          `json:"protocol,omitempty"`
	Notes     string          `json:"notes,omitempty"`
	Schedule  json.RawMessage `json:"schedule,omitempty"`
}

type RulePause struct {
	Duration int    `json:"duration,omitempty"`
	Box      string `json:"box,omitempty"`
}

type RuleResume struct {
	Box string `json:"box,omitempty"`
}

func ruleListQuery(opts *RuleListOptions) url.Values {
	q := url.Values{}
	if opts == nil {
		return q
	}
	if opts.Query != "" {
		q.Set("query", opts.Query)
	}
	if opts.GroupBy != "" {
		q.Set("groupBy", opts.GroupBy)
	}
	if opts.SortBy != "" {
		q.Set("sortBy", opts.SortBy)
	}
	if opts.Limit > 0 {
		q.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Cursor != "" {
		q.Set("cursor", opts.Cursor)
	}
	return q
}

func (s *RulesService) List(ctx context.Context, opts *RuleListOptions) (*Page[Rule], error) {
	var out Page[Rule]
	if err := s.client.do(ctx, "GET", "/rules", ruleListQuery(opts), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *RulesService) All(ctx context.Context, opts *RuleListOptions) iter.Seq2[Rule, error] {
	return paginate(ctx, func(cursor string) (*Page[Rule], error) {
		o := RuleListOptions{}
		if opts != nil {
			o = *opts
		}
		o.Cursor = cursor
		return s.List(ctx, &o)
	})
}

func (s *RulesService) Create(ctx context.Context, body *RuleCreate) (*Rule, error) {
	if body == nil || body.Action == "" {
		return nil, errors.New("firewalla: rule Action is required")
	}
	if body.Target == nil {
		return nil, errors.New("firewalla: rule Target is required")
	}
	var out Rule
	if err := s.client.do(ctx, "POST", "/rules", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *RulesService) Pause(ctx context.Context, ruleID string, body *RulePause) error {
	if ruleID == "" {
		return errors.New("firewalla: ruleID is required")
	}
	return s.client.do(ctx, "POST", "/rules/"+url.PathEscape(ruleID)+"/pause", nil, body, nil)
}

func (s *RulesService) Resume(ctx context.Context, ruleID string, body *RuleResume) error {
	if ruleID == "" {
		return errors.New("firewalla: ruleID is required")
	}
	return s.client.do(ctx, "POST", "/rules/"+url.PathEscape(ruleID)+"/resume", nil, body, nil)
}

func (s *RulesService) Delete(ctx context.Context, ruleID string) error {
	if ruleID == "" {
		return errors.New("firewalla: ruleID is required")
	}
	return s.client.do(ctx, "DELETE", "/rules/"+url.PathEscape(ruleID), nil, nil, nil)
}
