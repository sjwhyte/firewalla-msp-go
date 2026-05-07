package firewalla

import (
	"context"
	"encoding/json"
	"net/url"
	"time"
)

// TrendPoint is one daily data point in a trend series.
type TrendPoint struct {
	TS    time.Time
	Value int
}

func (p *TrendPoint) UnmarshalJSON(data []byte) error {
	var raw struct {
		TS    int64 `json:"ts"`
		Value int   `json:"value"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	p.TS = time.Unix(raw.TS, 0).UTC()
	p.Value = raw.Value
	return nil
}

func (p TrendPoint) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		TS    int64 `json:"ts"`
		Value int   `json:"value"`
	}{TS: p.TS.Unix(), Value: p.Value})
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
