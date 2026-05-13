// Tour is a read-only walkthrough that hits most resources in the
// Firewalla MSP API. It makes only GET requests and does not mutate
// any state.
//
// Run with:
//
//	FW_DOMAIN=... FW_TOKEN=... go run ./examples/tour
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
	c, err := firewalla.NewClient(
		os.Getenv("FW_DOMAIN"),
		os.Getenv("FW_TOKEN"),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	section("Simple stats")
	if s, err := c.Stats.Simple(ctx, ""); err == nil {
		fmt.Printf("  onlineBoxes=%d offlineBoxes=%d alarms=%d rules=%d\n",
			s.OnlineBoxes, s.OfflineBoxes, s.Alarms, s.Rules)
	} else {
		warn(err)
	}

	section("Boxes")
	if boxes, err := c.Boxes.List(ctx, nil); err == nil {
		for _, b := range boxes {
			online := "offline"
			if b.Online {
				online = "online"
			}
			fmt.Printf("  %s  %-20s  %-10s  %s  (devices=%d rules=%d alarms=%d, last seen %s)\n",
				b.GID, b.Name, b.Model, online, b.DeviceCount, b.RuleCount, b.AlarmCount,
				b.LastSeen.Format(time.RFC3339))
		}
	} else {
		warn(err)
	}

	section("Devices (first 10)")
	if devices, err := c.Devices.List(ctx, nil); err == nil {
		for i, d := range devices {
			if i == 10 {
				fmt.Printf("  ... +%d more\n", len(devices)-10)
				break
			}
			online := "off"
			if d.Online {
				online = "on "
			}
			fmt.Printf("  %s  %-25s  %-15s  %s  dl=%-10d ul=%-10d\n",
				online, d.Name, d.IP, d.MACVendor, d.TotalDownload, d.TotalUpload)
		}
	} else {
		warn(err)
	}

	section("Top 5 boxes by blocked flows")
	if rows, err := c.Stats.TopBoxesByBlockedFlows(ctx, &firewalla.StatsOptions{Limit: 5}); err == nil {
		for _, r := range rows {
			fmt.Printf("  %-30s  %d  (%s)\n", r.Meta.Name, r.Value, r.Meta.Model)
		}
	} else {
		warn(err)
	}

	section("Top 5 regions by blocked flows")
	if rows, err := c.Stats.TopRegionsByBlockedFlows(ctx, &firewalla.StatsOptions{Limit: 5}); err == nil {
		for _, r := range rows {
			fmt.Printf("  %s  %d\n", r.Meta.Code, r.Value)
		}
	} else {
		warn(err)
	}

	section("Alarms trend (last 30 days)")
	if pts, err := c.Trends.Alarms(ctx, ""); err == nil {
		for _, p := range pts {
			fmt.Printf("  %s  %d\n", p.TS.Format("2006-01-02"), p.Value)
		}
	} else {
		warn(err)
	}

	section("First 5 active alarms (newest first)")
	page, err := c.Alarms.List(ctx, &firewalla.AlarmListOptions{
		Query:  "status:active",
		SortBy: "ts:desc",
		Limit:  5,
	})
	if err != nil {
		warn(err)
	} else {
		for _, a := range page.Results {
			dev := "?"
			if a.Device != nil {
				dev = a.Device.Name
			}
			fmt.Printf("  aid=%s  %-22s  %-20s  %s\n",
				a.AID, a.TypeName, dev, a.TS.Format(time.RFC3339))
		}
		if page.NextCursor != "" {
			fmt.Printf("  (more available; next_cursor=%q)\n", page.NextCursor)
		}
	}

	section("Flows in last hour, >1MB total (first 5)")
	hourAgo := time.Now().Add(-1 * time.Hour).Unix()
	q := fmt.Sprintf("ts:>%d total:>1MB", hourAgo)
	fpage, err := c.Flows.List(ctx, &firewalla.FlowListOptions{
		Query:  q,
		SortBy: "ts:desc",
		Limit:  5,
	})
	if err != nil {
		warn(err)
	} else {
		for _, f := range fpage.Results {
			dev := "?"
			if f.Device != nil {
				dev = f.Device.Name
			}
			fmt.Printf("  %-8s  %-25s  %-25s  dl=%-10d ul=%-10d total=%d\n",
				f.Direction, dev, f.Domain, f.Download, f.Upload, f.Total)
		}
	}

	section("Rules (first 5)")
	rpage, err := c.Rules.List(ctx, &firewalla.RuleListOptions{Limit: 5})
	if err != nil {
		warn(err)
	} else {
		for _, r := range rpage.Results {
			target := "?"
			if r.Target != nil {
				target = fmt.Sprintf("%s=%s", r.Target.Type, r.Target.Value)
			}
			fmt.Printf("  %s  %-7s  %-12s  %s\n", r.ID, r.Action, r.Direction, target)
		}
	}

	section("Target lists")
	if lists, err := c.TargetLists.List(ctx); err == nil {
		for _, tl := range lists {
			fmt.Printf("  %s  owner=%-40s  %-20s  (%d targets, cat=%s)\n",
				tl.ID, tl.Owner, tl.Name, len(tl.Targets), tl.Category)
		}
	} else {
		warn(err)
	}
}

func section(name string) {
	fmt.Printf("\n=== %s ===\n", name)
}

func warn(err error) {
	switch {
	case firewalla.IsNotFound(err):
		fmt.Println("  not found")
	case firewalla.IsUnauthorized(err):
		fmt.Println("  unauthorized (bad token?)")
	case firewalla.IsRateLimited(err):
		fmt.Println("  rate-limited")
	default:
		fmt.Printf("  error: %v\n", err)
	}
}
