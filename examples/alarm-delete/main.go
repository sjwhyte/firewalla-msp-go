// ⚠ This example permanently deletes an alarm from your MSP.
//
// Run with:
//
//	FW_DOMAIN=... FW_TOKEN=... \
//	  go run ./examples/alarm-delete -gid <box-gid> -aid <alarm-id>
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
	gid := flag.String("gid", "", "Box GID (required)")
	aid := flag.String("aid", "", "Alarm AID (required)")
	flag.Parse()
	if *gid == "" || *aid == "" {
		flag.Usage()
		os.Exit(2)
	}

	c, err := firewalla.NewClient(os.Getenv("FW_DOMAIN"), os.Getenv("FW_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := c.Alarms.Delete(ctx, *gid, firewalla.AlarmID(*aid)); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("deleted alarm gid=%s aid=%s\n", *gid, *aid)
}
