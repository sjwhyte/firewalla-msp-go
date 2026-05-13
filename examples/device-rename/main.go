// ⚠ This example writes to your MSP. It renames a device.
//
// Run with:
//
//	FW_DOMAIN=... FW_TOKEN=... \
//	  go run ./examples/device-rename \
//	    -gid <box-gid> \
//	    -device <device-id-or-mac> \
//	    -name "New Name"
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
	device := flag.String("device", "", "Device ID, usually a MAC like mac:AA:BB:CC:DD:EE:FF (required)")
	name := flag.String("name", "", "New device name (required)")
	flag.Parse()
	if *gid == "" || *device == "" || *name == "" {
		flag.Usage()
		os.Exit(2)
	}

	c, err := firewalla.NewClient(os.Getenv("FW_DOMAIN"), os.Getenv("FW_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	d, err := c.Devices.Update(ctx, *gid, *device, &firewalla.DeviceUpdate{Name: name})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("renamed: id=%s name=%s\n", d.ID, d.Name)
}
