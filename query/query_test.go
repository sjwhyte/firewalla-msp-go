package query

import (
	"testing"
	"time"
)

func TestEq(t *testing.T) {
	if got := Eq("status", "active").String(); got != "status:active" {
		t.Errorf("Eq = %q", got)
	}
}

func TestEq_QuotesSpecialChars(t *testing.T) {
	cases := []struct{ value, want string }{
		{"hello world", `box.name:"hello world"`},
		{"a,b", `box.name:"a,b"`},
		{"a:b", `box.name:"a:b"`},
		{`he said "hi"`, `box.name:"he said \"hi\""`},
		{"", `box.name:""`},
	}
	for _, c := range cases {
		got := Eq("box.name", c.value).String()
		if got != c.want {
			t.Errorf("Eq(%q) = %q, want %q", c.value, got, c.want)
		}
	}
}

func TestLike(t *testing.T) {
	if got := Like("device.name", "iphone").String(); got != "device.name:*iphone*" {
		t.Errorf("Like (no stars) = %q", got)
	}
	if got := Like("device.name", "iphone*").String(); got != "device.name:iphone*" {
		t.Errorf("Like (trailing) = %q", got)
	}
	if got := Like("device.name", "*iphone").String(); got != "device.name:*iphone" {
		t.Errorf("Like (leading) = %q", got)
	}
}

func TestNot(t *testing.T) {
	if got := Not(Eq("status", "ok")).String(); got != "-status:ok" {
		t.Errorf("Not = %q", got)
	}
}

func TestAnd(t *testing.T) {
	got := And(Eq("status", "active"), Eq("type", "1")).String()
	if got != "status:active type:1" {
		t.Errorf("And = %q", got)
	}
}

func TestOr_OnSameQualifier(t *testing.T) {
	got := Or(Eq("box.name", "Gold Plus"), Eq("box.name", "Purple")).String()
	want := `box.name:"Gold Plus",Purple`
	if got != want {
		t.Errorf("Or = %q, want %q", got, want)
	}
}

func TestComparators(t *testing.T) {
	cases := []struct {
		expr Expr
		want string
	}{
		{GT("transfer.total", "50MB"), "transfer.total:>50MB"},
		{GTE("ts", 1714867200), "ts:>=1714867200"},
		{LT("dport", 1024), "dport:<1024"},
		{LTE("upload", "1GB"), "upload:<=1GB"},
		{Range("ts", 1714000000, 1714867200), "ts:1714000000-1714867200"},
	}
	for _, c := range cases {
		if got := c.expr.String(); got != c.want {
			t.Errorf("expr = %q, want %q", got, c.want)
		}
	}
}

func TestEmptyExprIgnored(t *testing.T) {
	got := And(Eq("a", "1"), Eq("", "")).String()
	if got != "a:1" {
		t.Errorf("And with empty = %q", got)
	}
}

func TestAlarmHelpers(t *testing.T) {
	if got := AlarmStatus(AlarmActive).String(); got != "status:active" {
		t.Errorf("AlarmStatus = %q", got)
	}
	if got := AlarmStatus(AlarmArchived).String(); got != "status:archived" {
		t.Errorf("AlarmStatus archived = %q", got)
	}
	if got := AlarmTypeIs(1, 5).String(); got != "type:1,5" {
		t.Errorf("AlarmTypeIs = %q", got)
	}
	if got := BoxName("Gold Plus", "Purple").String(); got != `box.name:"Gold Plus",Purple` {
		t.Errorf("BoxName = %q", got)
	}
	if got := DeviceNameLike("iphone").String(); got != "device.name:*iphone*" {
		t.Errorf("DeviceNameLike = %q", got)
	}
	if got := TransferTotalGT("50MB").String(); got != "transfer.total:>50MB" {
		t.Errorf("TransferTotalGT = %q", got)
	}
	t0 := time.Unix(1714867200, 0).UTC()
	if got := TSAfter(t0).String(); got != "ts:>1714867200" {
		t.Errorf("TSAfter = %q", got)
	}
}

func TestFlowHelpers(t *testing.T) {
	if got := Direction(DirectionInbound).String(); got != "direction:inbound" {
		t.Errorf("Direction = %q", got)
	}
	if got := Category("social", "video").String(); got != "category:social,video" {
		t.Errorf("Category = %q", got)
	}
	if got := DPort(443).String(); got != "dport:443" {
		t.Errorf("DPort = %q", got)
	}
	if got := TotalGT("100MB").String(); got != "total:>100MB" {
		t.Errorf("TotalGT = %q", got)
	}
}

func TestRuleHelpers(t *testing.T) {
	if got := RuleStatus("active").String(); got != "status:active" {
		t.Errorf("RuleStatus = %q", got)
	}
	if got := RuleAction("block").String(); got != "action:block" {
		t.Errorf("RuleAction = %q", got)
	}
	if got := BoxID("G1").String(); got != "box.id:G1" {
		t.Errorf("BoxID = %q", got)
	}
}
