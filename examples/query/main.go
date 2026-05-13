// Query demonstrates building a non-trivial Firewalla search expression
// with the query subpackage and using it to fetch alarms.
//
// Run with:
//
//	FW_DOMAIN=... FW_TOKEN=... go run ./examples/query
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sjwhyte/firewalla-msp-go"
	"github.com/sjwhyte/firewalla-msp-go/query"
)

func main() {
	c, err := firewalla.NewClient(os.Getenv("FW_DOMAIN"), os.Getenv("FW_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	q := query.And(
		query.AlarmStatus(query.AlarmActive),
		query.TransferTotalGT("10MB"),
		query.TSAfter(time.Now().Add(-24*time.Hour)),
		query.Not(query.RemoteCategory("intel")),
	)
	fmt.Printf("Query: %s\n\n", q.String())

	page, err := c.Alarms.List(context.Background(), &firewalla.AlarmListOptions{
		Query:  q.String(),
		SortBy: "transfer.total:desc",
		Limit:  10,
	})
	if err != nil {
		log.Fatal(err)
	}
	if len(page.Results) == 0 {
		fmt.Println("(no matching alarms)")
		return
	}
	for _, a := range page.Results {
		total := int64(0)
		if a.Transfer != nil {
			total = a.Transfer.Total
		}
		fmt.Printf("  %-22s  %s  total=%d\n",
			a.TypeName, a.TS.Format(time.RFC3339), total)
	}
}
