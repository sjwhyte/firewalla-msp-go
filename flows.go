package firewalla

import (
	"context"
	"iter"
	"net/url"
	"strconv"
)

type Flow struct {
	TS        Timestamp    `json:"ts"`
	Status    string       `json:"status"`
	Direction string       `json:"direction"`
	Box       *FlowBoxRef  `json:"box,omitempty"`
	Device    *AlarmDevice `json:"device,omitempty"`
	Network   *NetworkRef  `json:"network,omitempty"`
	Category  string       `json:"category,omitempty"`
	Domain    string       `json:"domain,omitempty"`
	Region    string       `json:"region,omitempty"`
	SPort     int          `json:"sport"`
	DPort     int          `json:"dport"`
	Download  int64        `json:"download"`
	Upload    int64        `json:"upload"`
	Total     int64        `json:"total"`
	Protocol  string       `json:"protocol,omitempty"`
}

type FlowBoxRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type FlowListOptions struct {
	Query   string
	GroupBy string
	SortBy  string
	Limit   int
	Cursor  string
}

func flowListQuery(opts *FlowListOptions) url.Values {
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

func (s *FlowsService) List(ctx context.Context, opts *FlowListOptions) (*Page[Flow], error) {
	var out Page[Flow]
	if err := s.client.do(ctx, "GET", "/flows", flowListQuery(opts), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *FlowsService) All(ctx context.Context, opts *FlowListOptions) iter.Seq2[Flow, error] {
	return paginate(ctx, func(cursor string) (*Page[Flow], error) {
		o := FlowListOptions{}
		if opts != nil {
			o = *opts
		}
		o.Cursor = cursor
		return s.List(ctx, &o)
	})
}
