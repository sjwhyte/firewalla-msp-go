// ⚠ This example writes to your MSP.
//
// Rules-lifecycle creates a block rule targeting example.invalid (RFC 2606
// reserved, no real network impact), pauses it, resumes it, and deletes
// it. The rule exists only for the duration of this program.
//
// Run with:
//
//	FW_DOMAIN=... FW_TOKEN=... \
//	  go run ./examples/rules-lifecycle -gid <box-gid>
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sjwhyte/firewalla-msp-go"
)

func main() {
	gid := flag.String("gid", "", "Firewalla box GID this rule applies to (required)")
	flag.Parse()
	if *gid == "" {
		flag.Usage()
		os.Exit(2)
	}

	c, err := firewalla.NewClient(os.Getenv("FW_DOMAIN"), os.Getenv("FW_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	fmt.Println("create...")
	r, err := c.Rules.Create(ctx, &firewalla.RuleCreate{
		Action:    "block",
		Direction: "bidirection",
		GID:       *gid,
		Target:    &firewalla.RuleTarget{Type: "domain", Value: "example.invalid", DNSOnly: true},
		Notes:     "msp-go-example: rules-lifecycle demo",
	})
	if err != nil {
		log.Fatalf("create: %v", err)
	}
	fmt.Printf("  created %s\n", r.ID)

	fmt.Println("pause...")
	if err := c.Rules.Pause(ctx, r.ID, &firewalla.RulePause{Duration: 60, Box: *gid}); err != nil {
		log.Fatalf("pause: %v", err)
	}
	fmt.Println("  paused")

	fmt.Println("resume...")
	if err := c.Rules.Resume(ctx, r.ID, &firewalla.RuleResume{Box: *gid}); err != nil {
		log.Fatalf("resume: %v", err)
	}
	fmt.Println("  resumed")

	fmt.Println("delete...")
	if err := c.Rules.Delete(ctx, r.ID); err != nil {
		log.Fatalf("delete: %v", err)
	}
	fmt.Println("  deleted")
}
