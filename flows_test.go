package firewalla

import (
	"context"
	"net/http"
	"os"
	"testing"
)

func TestFlowsService_List(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()
	mux.HandleFunc("/v2/flows", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("query") != "category:social" {
			t.Errorf("query = %q", r.URL.Query().Get("query"))
		}
		w.WriteHeader(200)
		body, _ := os.ReadFile("internal/testdata/flows/list.json")
		_, _ = w.Write(body)
	})
	page, err := c.Flows.List(context.Background(), &FlowListOptions{Query: "category:social"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(page.Results) != 2 || page.Results[0].Category != "social" {
		t.Errorf("results = %+v", page.Results)
	}
	if page.Results[0].DPort != 443 || page.Results[0].Protocol != "tcp" {
		t.Errorf("flow fields wrong: %+v", page.Results[0])
	}
}

func TestFlowsService_All(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()
	mux.HandleFunc("/v2/flows", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		body, _ := os.ReadFile("internal/testdata/flows/list.json")
		_, _ = w.Write(body)
	})
	count := 0
	for _, err := range c.Flows.All(context.Background(), nil) {
		if err != nil {
			t.Fatalf("All: %v", err)
		}
		count++
	}
	if count != 2 {
		t.Errorf("count = %d", count)
	}
}
