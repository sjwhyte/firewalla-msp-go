package firewalla

import (
	"context"
	"net/http"
	"os"
	"testing"
)

func TestBoxesService_List_All(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()

	mux.HandleFunc("/v2/boxes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("method = %s", r.Method)
		}
		if got := r.URL.Query().Get("group"); got != "" {
			t.Errorf("unexpected group: %q", got)
		}
		w.WriteHeader(200)
		body, _ := os.ReadFile("internal/testdata/boxes/list.json")
		_, _ = w.Write(body)
	})

	got, err := c.Boxes.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d", len(got))
	}
	if got[0].GID != "00000000-0000-0000-0000-000000000001" {
		t.Errorf("GID = %s", got[0].GID)
	}
	if got[0].Name != "Office Box" || got[0].Model != "gold-plus" {
		t.Errorf("name/model wrong")
	}
	if !got[0].Online {
		t.Errorf("online = false, want true")
	}
}

func TestBoxesService_List_Group(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()

	mux.HandleFunc("/v2/boxes", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("group"); got != "1" {
			t.Errorf("group = %q, want 1", got)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`[]`))
	})

	_, err := c.Boxes.List(context.Background(), &BoxListOptions{Group: "1"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
}
