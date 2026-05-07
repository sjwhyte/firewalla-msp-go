package query

import (
	"strconv"
	"strings"
	"time"
)

type AlarmStatusValue string

const (
	AlarmActive   AlarmStatusValue = "active"
	AlarmArchived AlarmStatusValue = "archived"
)

func AlarmStatus(s AlarmStatusValue) Expr { return Eq("status", string(s)) }

func AlarmTypeIs(types ...int) Expr {
	if len(types) == 0 {
		return rawExpr("")
	}
	parts := make([]string, 0, len(types))
	for _, t := range types {
		parts = append(parts, strconv.Itoa(t))
	}
	return rawExpr("type:" + strings.Join(parts, ","))
}

func BoxName(names ...string) Expr        { return orEqValues("box.name", names) }
func DeviceID(ids ...string) Expr         { return orEqValues("device.id", ids) }
func DeviceName(names ...string) Expr     { return orEqValues("device.name", names) }
func DeviceNameLike(glob string) Expr     { return Like("device.name", glob) }
func DeviceNetworkName(n ...string) Expr  { return orEqValues("device.network.name", n) }
func RemoteCategory(cats ...string) Expr  { return orEqValues("remote.category", cats) }
func RemoteDomain(domains ...string) Expr { return orEqValues("remote.domain", domains) }
func RemoteRegion(regions ...string) Expr { return orEqValues("remote.region", regions) }

func TransferTotalGT(v string) Expr     { return GT("transfer.total", v) }
func TransferTotalGTE(v string) Expr    { return GTE("transfer.total", v) }
func TransferDownloadGT(v string) Expr  { return GT("transfer.download", v) }
func TransferUploadGT(v string) Expr    { return GT("transfer.upload", v) }

func TSAfter(t time.Time) Expr  { return GT("ts", t.Unix()) }
func TSBefore(t time.Time) Expr { return LT("ts", t.Unix()) }
func TSBetween(lo, hi time.Time) Expr {
	return Range("ts", lo.Unix(), hi.Unix())
}

// orEqValues returns Or(Eq(q, v0), Eq(q, v1), ...). Empty input → empty expression.
func orEqValues(qualifier string, values []string) Expr {
	if len(values) == 0 {
		return rawExpr("")
	}
	exprs := make([]Expr, 0, len(values))
	for _, v := range values {
		exprs = append(exprs, Eq(qualifier, v))
	}
	return Or(exprs...)
}
