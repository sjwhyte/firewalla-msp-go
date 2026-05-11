package firewalla

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
)

func TestDeviceGroup_AcceptsStringOrObject(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want DeviceGroup
	}{
		{"string", `"Media"`, DeviceGroup{ID: "Media"}},
		{"object", `{"id":1,"name":"Media"}`, DeviceGroup{ID: "1", Name: "Media"}},
		{"null", `null`, DeviceGroup{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var g DeviceGroup
			if err := json.Unmarshal([]byte(c.raw), &g); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			if g != c.want {
				t.Errorf("got %+v, want %+v", g, c.want)
			}
		})
	}
}

func TestDevicesService_List(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()

	mux.HandleFunc("/v2/devices", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("box") != "" {
			t.Errorf("unexpected box filter")
		}
		w.WriteHeader(200)
		body, _ := os.ReadFile("internal/testdata/devices/list.json")
		_, _ = w.Write(body)
	})

	got, err := c.Devices.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 1 || got[0].Name != "Living Room AppleTV" {
		t.Errorf("decoded wrong: %+v", got)
	}
	if !got[0].Online || got[0].IPReserved != true {
		t.Errorf("flags wrong: %+v", got[0])
	}
}

func TestDevicesService_List_Box(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()
	mux.HandleFunc("/v2/devices", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("box") != "G1" {
			t.Errorf("box = %q", r.URL.Query().Get("box"))
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`[]`))
	})
	_, err := c.Devices.List(context.Background(), &DeviceListOptions{Box: "G1"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestDevicesService_List_Group(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()
	mux.HandleFunc("/v2/devices", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("group") != "9" {
			t.Errorf("group = %q", r.URL.Query().Get("group"))
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`[]`))
	})
	_, err := c.Devices.List(context.Background(), &DeviceListOptions{Group: "9"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestDevicesService_List_BoxAndGroupMutuallyExclusive(t *testing.T) {
	c, _, teardown := newTestServer(t)
	defer teardown()
	_, err := c.Devices.List(context.Background(), &DeviceListOptions{Box: "G1", Group: "9"})
	if err == nil {
		t.Fatalf("expected error for mutually exclusive Box+Group")
	}
}

func TestDevicesService_Update(t *testing.T) {
	c, mux, teardown := newTestServer(t)
	defer teardown()

	mux.HandleFunc("/v2/boxes/G1/devices/mac:AA:BB:CC:DD:EE:FF", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("method = %s", r.Method)
		}
		var got DeviceUpdate
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Name == nil || *got.Name != "My Updated Device" {
			t.Errorf("name = %v", got.Name)
		}
		w.WriteHeader(200)
		body, _ := os.ReadFile("internal/testdata/devices/updated.json")
		_, _ = w.Write(body)
	})

	name := "My Updated Device"
	d, err := c.Devices.Update(context.Background(), "G1", "mac:AA:BB:CC:DD:EE:FF", &DeviceUpdate{Name: &name})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if d.Name != "My Updated Device" {
		t.Errorf("name = %s", d.Name)
	}
}
