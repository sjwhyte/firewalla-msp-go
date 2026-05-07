package query

type DirectionValue string

const (
	DirectionInbound  DirectionValue = "inbound"
	DirectionOutbound DirectionValue = "outbound"
	DirectionLocal    DirectionValue = "local"
)

func Direction(d DirectionValue) Expr { return Eq("direction", string(d)) }

func FlowStatus(s string) Expr         { return Eq("status", s) }
func Category(cats ...string) Expr     { return orEqValues("category", cats) }
func Domain(domains ...string) Expr    { return orEqValues("domain", domains) }
func Region(regions ...string) Expr    { return orEqValues("region", regions) }
func NetworkID(ids ...string) Expr     { return orEqValues("network.id", ids) }
func NetworkName(names ...string) Expr { return orEqValues("network.name", names) }

func SPort(p int) Expr { return Eq("sport", itoa(p)) }
func DPort(p int) Expr { return Eq("dport", itoa(p)) }

func DownloadGT(v string) Expr { return GT("download", v) }
func UploadGT(v string) Expr   { return GT("upload", v) }
func TotalGT(v string) Expr    { return GT("total", v) }
