package firewalla

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
)

func TestTargetListsService_List(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()
	mux.HandleFunc("/v2/target-lists", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("method = %s", r.Method)
		}
		if got := r.URL.Query().Get("owner"); got != "" {
			t.Errorf("owner = %q, want empty", got)
		}
		w.WriteHeader(200)
		body, _ := os.ReadFile("internal/testdata/targetlists/list.json")
		_, _ = w.Write(body)
	})
	got, err := c.TargetLists.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 1 || got[0].Name != "Block Ads" {
		t.Errorf("got = %+v", got)
	}
}

func TestTargetListsService_List_OwnerFilter(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()
	mux.HandleFunc("/v2/target-lists", func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Query().Get("owner"), "box-gid-1"; got != want {
			t.Errorf("owner = %q, want %q", got, want)
		}
		w.WriteHeader(200)
		body, _ := os.ReadFile("internal/testdata/targetlists/list.json")
		_, _ = w.Write(body)
	})
	if _, err := c.TargetLists.List(context.Background(), &TargetListListOptions{Owner: "box-gid-1"}); err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestTargetListsService_Get(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()
	mux.HandleFunc("/v2/target-lists/TL-1", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		body, _ := os.ReadFile("internal/testdata/targetlists/get.json")
		_, _ = w.Write(body)
	})
	got, err := c.TargetLists.Get(context.Background(), "TL-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Owner != "global" {
		t.Errorf("Owner = %s", got.Owner)
	}
}

func TestTargetListsService_Create(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()
	mux.HandleFunc("/v2/target-lists", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s", r.Method)
		}
		var body TargetListCreate
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body.Name != "My List" || len(body.Targets) != 2 {
			t.Errorf("body = %+v", body)
		}
		w.WriteHeader(200)
		body2, _ := os.ReadFile("internal/testdata/targetlists/get.json")
		_, _ = w.Write(body2)
	})
	tl, err := c.TargetLists.Create(context.Background(), &TargetListCreate{
		Name:    "My List",
		Targets: []string{"a.com", "b.com"},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if tl == nil || tl.ID == "" {
		t.Errorf("returned tl = %+v", tl)
	}
}

func TestTargetListsService_Update(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()
	mux.HandleFunc("/v2/target-lists/TL-1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("method = %s", r.Method)
		}
		w.WriteHeader(200)
		body, _ := os.ReadFile("internal/testdata/targetlists/get.json")
		_, _ = w.Write(body)
	})
	name := "renamed"
	_, err := c.TargetLists.Update(context.Background(), "TL-1", &TargetListUpdate{Name: &name})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestTargetListsService_Delete(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()
	mux.HandleFunc("/v2/target-lists/TL-1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("method = %s", r.Method)
		}
		w.WriteHeader(204)
	})
	if err := c.TargetLists.Delete(context.Background(), "TL-1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}
