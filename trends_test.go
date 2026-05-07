package firewalla

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestTrendPoint_UnmarshalJSON(t *testing.T) {
	var tp TrendPoint
	if err := json.Unmarshal([]byte(`{"ts":1714867200,"value":42}`), &tp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	want := time.Unix(1714867200, 0).UTC()
	if !tp.TS.Equal(want) {
		t.Errorf("TS = %v, want %v", tp.TS, want)
	}
	if tp.Value != 42 {
		t.Errorf("Value = %d", tp.Value)
	}
}

func TestTrendsService_Flows(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()
	mux.HandleFunc("/v2/trends/flows", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("group"); got != "" {
			t.Errorf("group = %q", got)
		}
		w.WriteHeader(200)
		body, _ := os.ReadFile("internal/testdata/trends/flows.json")
		_, _ = w.Write(body)
	})
	got, err := c.Trends.Flows(context.Background(), "")
	if err != nil {
		t.Fatalf("Flows: %v", err)
	}
	if len(got) != 3 || got[0].Value != 12 {
		t.Errorf("got = %+v", got)
	}
}

func TestTrendsService_Alarms_WithGroup(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()
	mux.HandleFunc("/v2/trends/alarms", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("group"); got != "7" {
			t.Errorf("group = %q", got)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`[]`))
	})
	_, err := c.Trends.Alarms(context.Background(), "7")
	if err != nil {
		t.Fatalf("Alarms: %v", err)
	}
}

func TestTrendsService_Rules(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()
	mux.HandleFunc("/v2/trends/rules", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`[]`))
	})
	_, err := c.Trends.Rules(context.Background(), "")
	if err != nil {
		t.Fatalf("Rules: %v", err)
	}
}
