// Pagination demonstrates the two ways to walk a paginated list endpoint:
//  1. Manual cursor loop, where the caller drives pagination explicitly.
//  2. The Go 1.23+ range-over-func iterator returned by Alarms.All.
//
// Both halves print the same set of alarms.
//
// Run with:
//
//	FW_DOMAIN=... FW_TOKEN=... go run ./examples/pagination
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sjwhyte/firewalla-msp-go"
)

const targetCount = 50

func main() {
	c, err := firewalla.NewClient(os.Getenv("FW_DOMAIN"), os.Getenv("FW_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	fmt.Println("=== Manual cursor loop ===")
	seen := 0
	cursor := ""
	for seen < targetCount {
		page, err := c.Alarms.List(ctx, &firewalla.AlarmListOptions{
			Query:  "status:active",
			SortBy: "ts:desc",
			Limit:  20,
			Cursor: cursor,
		})
		if err != nil {
			log.Fatalf("List: %v", err)
		}
		for _, a := range page.Results {
			if seen >= targetCount {
				break
			}
			fmt.Printf("  manual  aid=%-6s  type=%s\n", a.AID, a.TypeName)
			seen++
		}
		if page.NextCursor == "" || seen >= targetCount {
			break
		}
		cursor = page.NextCursor
	}

	fmt.Println("\n=== Iterator (Alarms.All) ===")
	iSeen := 0
	for a, err := range c.Alarms.All(ctx, &firewalla.AlarmListOptions{
		Query:  "status:active",
		SortBy: "ts:desc",
		Limit:  20,
	}) {
		if err != nil {
			log.Fatalf("All: %v", err)
		}
		if iSeen >= targetCount {
			break
		}
		fmt.Printf("  iter    aid=%-6s  type=%s\n", a.AID, a.TypeName)
		iSeen++
	}
}
