package query

func RuleStatus(s string) Expr      { return Eq("status", s) }
func RuleAction(a string) Expr      { return Eq("action", a) }
func BoxID(ids ...string) Expr      { return orEqValues("box.id", ids) }
func BoxGroupID(ids ...string) Expr { return orEqValues("box.group.id", ids) }
