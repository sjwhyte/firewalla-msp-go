package firewalla

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
)

func TestRulesService_List(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()
	mux.HandleFunc("/v2/rules", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("query"); got != "box.id:G1" {
			t.Errorf("query = %q", got)
		}
		w.WriteHeader(200)
		body, _ := os.ReadFile("internal/testdata/rules/list.json")
		_, _ = w.Write(body)
	})
	page, err := c.Rules.List(context.Background(), &RuleListOptions{Query: "box.id:G1"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(page.Results) != 1 || page.Results[0].Action != "block" {
		t.Errorf("results = %+v", page.Results)
	}
	if page.Results[0].Target == nil || page.Results[0].Target.Value != "example.com" {
		t.Errorf("target = %+v", page.Results[0].Target)
	}
}

func TestRulesService_All(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()
	mux.HandleFunc("/v2/rules", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		body, _ := os.ReadFile("internal/testdata/rules/list.json")
		_, _ = w.Write(body)
	})
	count := 0
	for _, err := range c.Rules.All(context.Background(), nil) {
		if err != nil {
			t.Fatalf("All: %v", err)
		}
		count++
	}
	if count != 1 {
		t.Errorf("count = %d", count)
	}
}

func TestRulesService_Create(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()
	mux.HandleFunc("/v2/rules", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s", r.Method)
		}
		var got RuleCreate
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Action != "block" || got.Target.Value != "evil.example" {
			t.Errorf("body = %+v", got)
		}
		w.WriteHeader(200)
		body, _ := os.ReadFile("internal/testdata/rules/created.json")
		_, _ = w.Write(body)
	})
	r, err := c.Rules.Create(context.Background(), &RuleCreate{
		Action:    "block",
		Direction: "bidirection",
		GID:       "G1",
		Target:    &RuleTarget{Type: "domain", Value: "evil.example", DNSOnly: true},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if r.ID == "" {
		t.Errorf("ID empty")
	}
}

func TestRulesService_Pause(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()
	mux.HandleFunc("/v2/rules/R1/pause", func(w http.ResponseWriter, r *http.Request) {
		var got RulePause
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Duration != 60 || got.Box != "G1" {
			t.Errorf("body = %+v", got)
		}
		w.WriteHeader(204)
	})
	if err := c.Rules.Pause(context.Background(), "R1", &RulePause{Duration: 60, Box: "G1"}); err != nil {
		t.Fatalf("Pause: %v", err)
	}
}

func TestRulesService_Resume(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()
	mux.HandleFunc("/v2/rules/R1/resume", func(w http.ResponseWriter, r *http.Request) {
		var got RuleResume
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Box != "G1" {
			t.Errorf("box = %s", got.Box)
		}
		w.WriteHeader(204)
	})
	if err := c.Rules.Resume(context.Background(), "R1", &RuleResume{Box: "G1"}); err != nil {
		t.Fatalf("Resume: %v", err)
	}
}

func TestRulesService_Delete(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()
	mux.HandleFunc("/v2/rules/R1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("method = %s", r.Method)
		}
		w.WriteHeader(204)
	})
	if err := c.Rules.Delete(context.Background(), "R1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}
