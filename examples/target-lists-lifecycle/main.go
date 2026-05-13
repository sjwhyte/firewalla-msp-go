// ⚠ This example writes to your MSP.
//
// Target-lists-lifecycle creates a global target list named
// msp-go-example-<timestamp>, fetches it, updates the targets, then
// deletes it. Any artifact left from a crashed run is identifiable by
// the msp-go-example- prefix.
//
// Run with:
//
//	FW_DOMAIN=... FW_TOKEN=... go run ./examples/target-lists-lifecycle
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sjwhyte/firewalla-msp-go"
)

func main() {
	c, err := firewalla.NewClient(os.Getenv("FW_DOMAIN"), os.Getenv("FW_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	name := fmt.Sprintf("msp-go-example-%d", time.Now().Unix())

	fmt.Println("create...")
	tl, err := c.TargetLists.Create(ctx, &firewalla.TargetListCreate{
		Name:     name,
		Targets:  []string{"example.invalid", "demo.invalid"},
		Category: "edu",
		Notes:    "Created by msp-go-example lifecycle demo",
	})
	if err != nil {
		log.Fatalf("create: %v", err)
	}
	fmt.Printf("  created %s (name=%s)\n", tl.ID, tl.Name)

	fmt.Println("get...")
	got, err := c.TargetLists.Get(ctx, tl.ID)
	if err != nil {
		log.Fatalf("get: %v", err)
	}
	fmt.Printf("  name=%s targets=%v\n", got.Name, got.Targets)

	fmt.Println("update...")
	newTargets := []string{"example.invalid", "demo.invalid", "test.invalid"}
	newNotes := "Updated by demo"
	if _, err := c.TargetLists.Update(ctx, tl.ID, &firewalla.TargetListUpdate{
		Targets: &newTargets,
		Notes:   &newNotes,
	}); err != nil {
		log.Fatalf("update: %v", err)
	}
	fmt.Println("  updated")

	fmt.Println("delete...")
	if err := c.TargetLists.Delete(ctx, tl.ID); err != nil {
		log.Fatalf("delete: %v", err)
	}
	fmt.Println("  deleted")
}
