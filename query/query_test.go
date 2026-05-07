package query

import "testing"

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
