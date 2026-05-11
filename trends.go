package firewalla

import (
	"context"
	"net/url"
)

// TrendPoint is one daily data point in a trend series.
type TrendPoint struct {
	TS    Timestamp `json:"ts"`
	Value int       `json:"value"`
}

func (s *TrendsService) trend(ctx context.Context, kind, group string) ([]TrendPoint, error) {
	q := url.Values{}
	if group != "" {
		q.Set("group", group)
	}
	var out []TrendPoint
	if err := s.client.do(ctx, "GET", "/trends/"+kind, q, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *TrendsService) Flows(ctx context.Context, group string) ([]TrendPoint, error) {
	return s.trend(ctx, "flows", group)
}
func (s *TrendsService) Alarms(ctx context.Context, group string) ([]TrendPoint, error) {
	return s.trend(ctx, "alarms", group)
}
func (s *TrendsService) Rules(ctx context.Context, group string) ([]TrendPoint, error) {
	return s.trend(ctx, "rules", group)
}
