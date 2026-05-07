package firewalla

import (
	"context"
	"net/http"
	"os"
	"testing"
)

func TestAlarmType_String(t *testing.T) {
	cases := []struct {
		t    AlarmType
		want string
	}{
		{AlarmSecurityActivity, "SecurityActivity"},
		{AlarmAbnormalUpload, "AbnormalUpload"},
		{AlarmType(99), "Unknown(99)"},
	}
	for _, c := range cases {
		if got := c.t.String(); got != c.want {
			t.Errorf("AlarmType(%d).String() = %q, want %q", c.t, got, c.want)
		}
	}
}

func TestAlarmsService_List(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()
	mux.HandleFunc("/v2/alarms", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("query"); got != "status:active" {
			t.Errorf("query = %q", got)
		}
		if got := r.URL.Query().Get("sortBy"); got != "ts:desc" {
			t.Errorf("sortBy = %q", got)
		}
		if got := r.URL.Query().Get("limit"); got != "200" {
			t.Errorf("limit = %q", got)
		}
		w.WriteHeader(200)
		body, _ := os.ReadFile("internal/testdata/alarms/list_page1.json")
		_, _ = w.Write(body)
	})
	page, err := c.Alarms.List(context.Background(), &AlarmListOptions{
		Query:  "status:active",
		SortBy: "ts:desc",
		Limit:  200,
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(page.Results) != 2 || page.Results[0].AID != "1" {
		t.Errorf("results = %+v", page.Results)
	}
	if page.Results[0].Type != AlarmSecurityActivity {
		t.Errorf("type = %d", page.Results[0].Type)
	}
}

func TestAlarmsService_List_DefaultLimitOmitted(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()
	mux.HandleFunc("/v2/alarms", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Has("limit") {
			t.Errorf("limit unexpectedly present: %q", r.URL.Query().Get("limit"))
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"count":0,"results":[],"next_cursor":""}`))
	})
	_, err := c.Alarms.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestAlarmsService_Get(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()
	mux.HandleFunc("/v2/alarms/G/1", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		body, _ := os.ReadFile("internal/testdata/alarms/get.json")
		_, _ = w.Write(body)
	})
	a, err := c.Alarms.Get(context.Background(), "G", "1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if a.AID != "1" {
		t.Errorf("AID = %s", a.AID)
	}
}

func TestAlarmsService_Delete(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()
	mux.HandleFunc("/v2/alarms/G/1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("method = %s", r.Method)
		}
		w.WriteHeader(204)
	})
	if err := c.Alarms.Delete(context.Background(), "G", "1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestAlarmsService_Delete_RequiresIDs(t *testing.T) {
	c, _, teardown := newTestServer(t)
	defer teardown()
	if err := c.Alarms.Delete(context.Background(), "", "1"); err == nil {
		t.Error("expected error for empty gid")
	}
	if err := c.Alarms.Delete(context.Background(), "G", ""); err == nil {
		t.Error("expected error for empty aid")
	}
}
