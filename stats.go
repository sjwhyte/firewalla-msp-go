package firewalla

import (
	"context"
	"net/url"
	"strconv"
)

// StatRow is one entry in a top-N statistics table.
type StatRow[Meta any] struct {
	Meta  Meta `json:"meta"`
	Value int  `json:"value"`
}

type BoxMeta struct {
	GID   string `json:"gid"`
	Name  string `json:"name"`
	Model string `json:"model"`
}

type RegionMeta struct {
	Code string `json:"code"`
}

type SimpleStats struct {
	OnlineBoxes  int `json:"onlineBoxes"`
	OfflineBoxes int `json:"offlineBoxes"`
	Alarms       int `json:"alarms"`
	Rules        int `json:"rules"`
}

type StatsOptions struct {
	Group string
	Limit int
}

func statsQuery(opts *StatsOptions) url.Values {
	q := url.Values{}
	limit := 5
	if opts != nil {
		if opts.Group != "" {
			q.Set("group", opts.Group)
		}
		if opts.Limit > 0 {
			limit = opts.Limit
		}
	}
	q.Set("limit", strconv.Itoa(limit))
	return q
}

func (s *StatsService) topBoxes(ctx context.Context, kind string, opts *StatsOptions) ([]StatRow[BoxMeta], error) {
	var out []StatRow[BoxMeta]
	if err := s.client.do(ctx, "GET", "/stats/"+kind, statsQuery(opts), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *StatsService) TopBoxesByBlockedFlows(ctx context.Context, opts *StatsOptions) ([]StatRow[BoxMeta], error) {
	return s.topBoxes(ctx, "topBoxesByBlockedFlows", opts)
}

func (s *StatsService) TopBoxesBySecurityAlarms(ctx context.Context, opts *StatsOptions) ([]StatRow[BoxMeta], error) {
	return s.topBoxes(ctx, "topBoxesBySecurityAlarms", opts)
}

func (s *StatsService) TopRegionsByBlockedFlows(ctx context.Context, opts *StatsOptions) ([]StatRow[RegionMeta], error) {
	var out []StatRow[RegionMeta]
	if err := s.client.do(ctx, "GET", "/stats/topRegionsByBlockedFlows", statsQuery(opts), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *StatsService) Simple(ctx context.Context, group string) (*SimpleStats, error) {
	q := url.Values{}
	if group != "" {
		q.Set("group", group)
	}
	var out SimpleStats
	if err := s.client.do(ctx, "GET", "/stats/simple", q, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
