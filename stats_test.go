package firewalla

import (
	"context"
	"net/http"
	"os"
	"testing"
)

func TestStatsService_TopBoxesByBlockedFlows(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()
	mux.HandleFunc("/v2/stats/topBoxesByBlockedFlows", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "5" {
			t.Errorf("limit = %q", r.URL.Query().Get("limit"))
		}
		w.WriteHeader(200)
		body, _ := os.ReadFile("internal/testdata/stats/topBoxesByBlockedFlows.json")
		_, _ = w.Write(body)
	})
	got, err := c.Stats.TopBoxesByBlockedFlows(context.Background(), nil)
	if err != nil {
		t.Fatalf("TopBoxesByBlockedFlows: %v", err)
	}
	if len(got) != 2 || got[0].Value != 482 || got[0].Meta.Model != "gold-plus" {
		t.Errorf("got = %+v", got)
	}
}

func TestStatsService_TopBoxesBySecurityAlarms_GroupAndLimit(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()
	mux.HandleFunc("/v2/stats/topBoxesBySecurityAlarms", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("group") != "3" {
			t.Errorf("group = %q", r.URL.Query().Get("group"))
		}
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("limit = %q", r.URL.Query().Get("limit"))
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`[]`))
	})
	_, err := c.Stats.TopBoxesBySecurityAlarms(context.Background(), &StatsOptions{Group: "3", Limit: 10})
	if err != nil {
		t.Fatalf("TopBoxesBySecurityAlarms: %v", err)
	}
}

func TestStatsService_TopRegionsByBlockedFlows(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()
	mux.HandleFunc("/v2/stats/topRegionsByBlockedFlows", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		body, _ := os.ReadFile("internal/testdata/stats/topRegions.json")
		_, _ = w.Write(body)
	})
	got, err := c.Stats.TopRegionsByBlockedFlows(context.Background(), nil)
	if err != nil {
		t.Fatalf("TopRegionsByBlockedFlows: %v", err)
	}
	if len(got) != 2 || got[0].Meta.Code != "CN" || got[0].Value != 311 {
		t.Errorf("got = %+v", got)
	}
}

func TestStatsService_Simple(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()
	mux.HandleFunc("/v2/stats/simple", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		body, _ := os.ReadFile("internal/testdata/stats/simple.json")
		_, _ = w.Write(body)
	})
	got, err := c.Stats.Simple(context.Background(), "")
	if err != nil {
		t.Fatalf("Simple: %v", err)
	}
	if got.OnlineBoxes != 4 || got.Alarms != 12 {
		t.Errorf("got = %+v", got)
	}
}
