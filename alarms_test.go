package firewalla

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
)

func TestAlarmID_UnmarshalJSON(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want AlarmID
	}{
		{"numeric", `1`, "1"},
		{"numeric large", `1234567890`, "1234567890"},
		{"string", `"1"`, "1"},
		{"string alphanumeric", `"a1b2"`, "a1b2"},
		{"null", `null`, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var got AlarmID
			if err := got.UnmarshalJSON([]byte(c.raw)); err != nil {
				t.Fatalf("UnmarshalJSON: %v", err)
			}
			if got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestAlarm_DecodesNumericAndStringAID(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want AlarmID
	}{
		{"numeric aid", `{"aid":42,"gid":"G","type":1,"ts":1,"status":"active"}`, "42"},
		{"string aid", `{"aid":"42","gid":"G","type":1,"ts":1,"status":"active"}`, "42"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var a Alarm
			if err := json.Unmarshal([]byte(c.raw), &a); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			if a.AID != c.want {
				t.Errorf("AID = %q, want %q", a.AID, c.want)
			}
		})
	}
}

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

func TestAlarmsService_All_WalksAllPages(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()
	mux.HandleFunc("/v2/alarms", func(w http.ResponseWriter, r *http.Request) {
		cursor := r.URL.Query().Get("cursor")
		var path string
		switch cursor {
		case "":
			path = "internal/testdata/alarms/list_page1.json"
		case "c2":
			path = "internal/testdata/alarms/list_page2.json"
		case "c3":
			path = "internal/testdata/alarms/list_page3.json"
		default:
			t.Fatalf("unexpected cursor %q", cursor)
		}
		w.WriteHeader(200)
		body, _ := os.ReadFile(path)
		_, _ = w.Write(body)
	})

	var ids []string
	for a, err := range c.Alarms.All(context.Background(), nil) {
		if err != nil {
			t.Fatalf("All: %v", err)
		}
		ids = append(ids, string(a.AID))
	}
	want := []string{"1", "2", "3", "4", "5"}
	if len(ids) != len(want) {
		t.Fatalf("ids = %v, want %v", ids, want)
	}
	for i, id := range ids {
		if id != want[i] {
			t.Errorf("ids[%d] = %s, want %s", i, id, want[i])
		}
	}
}

func TestAlarmsService_All_IgnoresOptsCursor(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()
	first := true
	mux.HandleFunc("/v2/alarms", func(w http.ResponseWriter, r *http.Request) {
		if first {
			if got := r.URL.Query().Get("cursor"); got != "" {
				t.Errorf("first call cursor = %q, want empty (opts.Cursor must be ignored)", got)
			}
			first = false
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"count":0,"results":[],"next_cursor":""}`))
	})

	for _, err := range c.Alarms.All(context.Background(), &AlarmListOptions{Cursor: "should-be-ignored"}) {
		if err != nil {
			t.Fatalf("All: %v", err)
		}
	}
}
