package firewalla_test

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sjwhyte/firewalla-msp-go"
	"github.com/sjwhyte/firewalla-msp-go/query"
)

func ExampleNewClient() {
	c, err := firewalla.NewClient("mydomain.firewalla.net", os.Getenv("FW_TOKEN"))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	_ = c
}

func ExampleAlarmsService_List() {
	c, _ := firewalla.NewClient("mydomain.firewalla.net", "tok")
	ctx := context.Background()
	page, err := c.Alarms.List(ctx, &firewalla.AlarmListOptions{
		Query: "status:active",
		Limit: 50,
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("got", len(page.Results), "alarms; next cursor:", page.NextCursor != "")
}

func ExampleAlarmsService_All() {
	c, _ := firewalla.NewClient("mydomain.firewalla.net", "tok")
	ctx := context.Background()
	for alarm, err := range c.Alarms.All(ctx, &firewalla.AlarmListOptions{
		Query:  "status:active",
		SortBy: "ts:desc",
	}) {
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		_ = alarm
	}
}

func ExampleRulesService_Create() {
	c, _ := firewalla.NewClient("mydomain.firewalla.net", "tok")
	ctx := context.Background()
	rule, err := c.Rules.Create(ctx, &firewalla.RuleCreate{
		Action:    "block",
		Direction: "bidirection",
		GID:       "00000000-0000-0000-0000-000000000001",
		Target:    &firewalla.RuleTarget{Type: "domain", Value: "example.com", DNSOnly: true},
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	_ = rule
}

func Example_query() {
	q := query.And(
		query.AlarmStatus(query.AlarmActive),
		query.BoxName("Gold Plus", "Purple"),
		query.DeviceNameLike("iphone"),
		query.TransferTotalGT("50MB"),
		query.TSAfter(time.Now().Add(-24*time.Hour)),
	)
	fmt.Println(len(q.String()) > 0)
	// Output: true
}
