package firewalla

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTimestamp_UnmarshalJSON(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want time.Time
	}{
		{"int", `1778413842`, time.Unix(1778413842, 0).UTC()},
		{"float fractional", `1778413842.883`, time.Unix(1778413842, 883000000).UTC()},
		{"epoch zero", `0`, time.Unix(0, 0).UTC()},
		{"null", `null`, time.Time{}},
		{"string int", `"1778413842"`, time.Unix(1778413842, 0).UTC()},
		{"string float", `"1778413842.883"`, time.Unix(1778413842, 883000000).UTC()},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var ts Timestamp
			if err := ts.UnmarshalJSON([]byte(c.raw)); err != nil {
				t.Fatalf("UnmarshalJSON: %v", err)
			}
			if !ts.Equal(c.want) {
				t.Errorf("got %v, want %v", ts.Time, c.want)
			}
		})
	}
}

func TestTimestamp_DecodesAsStructField(t *testing.T) {
	type S struct {
		TS Timestamp `json:"ts"`
	}
	var s S
	if err := json.Unmarshal([]byte(`{"ts":1778413842.883}`), &s); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	want := time.Unix(1778413842, 883000000).UTC()
	if !s.TS.Equal(want) {
		t.Errorf("got %v, want %v", s.TS.Time, want)
	}
}

func TestTimestamp_MarshalJSON_RoundTrip(t *testing.T) {
	orig := Timestamp{time.Unix(1778413842, 883000000).UTC()}
	b, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var back Timestamp
	if err := back.UnmarshalJSON(b); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !back.Equal(orig.Time) {
		t.Errorf("round-trip lost data: got %v want %v (raw %s)", back.Time, orig.Time, string(b))
	}
}
