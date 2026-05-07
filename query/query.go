// Package query builds Firewalla MSP search expressions in a typed, escape-safe way.
//
// Pass the result of Expr.String() to a *ListOptions.Query field on the firewalla client:
//
//	q := query.And(
//	    query.Eq("status", "active"),
//	    query.GT("transfer.total", "50MB"),
//	)
//	c.Alarms.All(ctx, &firewalla.AlarmListOptions{Query: q.String()})
package query

import (
	"cmp"
	"fmt"
	"strconv"
	"strings"
)

// Expr is a serializable query expression.
type Expr interface {
	String() string
}

type rawExpr string

func (r rawExpr) String() string { return string(r) }

// Raw wraps a literal expression string. Use when the typed helpers don't cover a qualifier.
func Raw(s string) Expr { return rawExpr(s) }

// Eq returns a `qualifier:value` term, quoting the value if it contains
// space, comma, colon, or asterisk.
func Eq(qualifier, value string) Expr {
	if qualifier == "" {
		return rawExpr("")
	}
	return rawExpr(qualifier + ":" + escapeValue(value))
}

// Like returns a `qualifier:*value*` term. If value already contains a `*` at
// either end, it is left as-is. Asterisks are wildcard meta-characters here
// and are not quoted; values containing space, comma, or colon are quoted with
// the asterisks left outside the quotes.
func Like(qualifier, glob string) Expr {
	if qualifier == "" || glob == "" {
		return rawExpr("")
	}
	v := glob
	if !strings.HasPrefix(v, "*") && !strings.HasSuffix(v, "*") {
		v = "*" + v + "*"
	}
	return rawExpr(qualifier + ":" + escapeWildcardValue(v))
}

// escapeWildcardValue quotes the body of a wildcard expression while keeping
// leading/trailing `*` outside the quotes so they retain their glob semantics.
func escapeWildcardValue(v string) string {
	prefix, suffix := "", ""
	body := v
	if strings.HasPrefix(body, "*") {
		prefix = "*"
		body = body[1:]
	}
	if strings.HasSuffix(body, "*") {
		suffix = "*"
		body = body[:len(body)-1]
	}
	if strings.ContainsAny(body, " ,:") {
		body = `"` + strings.ReplaceAll(body, `"`, `\"`) + `"`
	}
	return prefix + body + suffix
}

// Not negates an expression by prefixing "-".
func Not(e Expr) Expr {
	s := e.String()
	if s == "" {
		return rawExpr("")
	}
	return rawExpr("-" + s)
}

// And joins terms with spaces (logical AND in Firewalla syntax).
func And(exprs ...Expr) Expr {
	parts := make([]string, 0, len(exprs))
	for _, e := range exprs {
		if s := e.String(); s != "" {
			parts = append(parts, s)
		}
	}
	return rawExpr(strings.Join(parts, " "))
}

// Or joins same-qualifier Eq terms with commas (logical OR in Firewalla syntax).
// All input terms must use the same qualifier; otherwise Or returns the
// space-separated AND form (since the API does not support cross-qualifier OR).
func Or(exprs ...Expr) Expr {
	if len(exprs) == 0 {
		return rawExpr("")
	}
	type kv struct{ q, v string }
	parsed := make([]kv, 0, len(exprs))
	for _, e := range exprs {
		s := e.String()
		if s == "" {
			continue
		}
		idx := strings.IndexByte(s, ':')
		if idx <= 0 {
			return And(exprs...)
		}
		parsed = append(parsed, kv{q: s[:idx], v: s[idx+1:]})
	}
	if len(parsed) == 0 {
		return rawExpr("")
	}
	q := parsed[0].q
	values := make([]string, 0, len(parsed))
	for _, p := range parsed {
		if p.q != q {
			return And(exprs...)
		}
		values = append(values, p.v)
	}
	return rawExpr(q + ":" + strings.Join(values, ","))
}

// GT, GTE, LT, LTE produce numeric comparators.
func GT[T cmp.Ordered](qualifier string, v T) Expr  { return cmpExpr(qualifier, ">", v) }
func GTE[T cmp.Ordered](qualifier string, v T) Expr { return cmpExpr(qualifier, ">=", v) }
func LT[T cmp.Ordered](qualifier string, v T) Expr  { return cmpExpr(qualifier, "<", v) }
func LTE[T cmp.Ordered](qualifier string, v T) Expr { return cmpExpr(qualifier, "<=", v) }

// Range produces `qualifier:lo-hi`.
func Range[T cmp.Ordered](qualifier string, lo, hi T) Expr {
	if qualifier == "" {
		return rawExpr("")
	}
	return rawExpr(fmt.Sprintf("%s:%v-%v", qualifier, lo, hi))
}

func cmpExpr[T cmp.Ordered](qualifier, op string, v T) Expr {
	if qualifier == "" {
		return rawExpr("")
	}
	return rawExpr(fmt.Sprintf("%s:%s%v", qualifier, op, v))
}

func itoa(i int) string { return strconv.Itoa(i) }

// escapeValue wraps the value in quotes if it contains space/comma/colon/asterisk,
// and escapes embedded quotes. The empty string becomes "".
func escapeValue(v string) string {
	if needsQuoting(v) {
		return `"` + strings.ReplaceAll(v, `"`, `\"`) + `"`
	}
	return v
}

func needsQuoting(v string) bool {
	if v == "" {
		return true
	}
	return strings.ContainsAny(v, " ,:*")
}
