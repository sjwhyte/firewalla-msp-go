// With-retry wraps the client's HTTP transport with the retry middleware
// from the retry subpackage. Retries: network errors, 429, 502, 503, 504.
// Honors Retry-After. The library never retries unless you opt in like
// this.
//
// Run with:
//
//	FW_DOMAIN=... FW_TOKEN=... go run ./examples/with-retry
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/sjwhyte/firewalla-msp-go"
	"github.com/sjwhyte/firewalla-msp-go/retry"
)

func main() {
	hc := &http.Client{
		Timeout: 30 * time.Second,
		Transport: retry.NewTransport(http.DefaultTransport, retry.Config{
			MaxAttempts: 4,
			BaseDelay:   500 * time.Millisecond,
			MaxDelay:    30 * time.Second,
			Jitter:      0.2,
		}),
	}
	c, err := firewalla.NewClient(
		os.Getenv("FW_DOMAIN"),
		os.Getenv("FW_TOKEN"),
		firewalla.WithHTTPClient(hc),
	)
	if err != nil {
		log.Fatal(err)
	}

	s, err := c.Stats.Simple(context.Background(), "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("onlineBoxes=%d offlineBoxes=%d alarms=%d rules=%d\n",
		s.OnlineBoxes, s.OfflineBoxes, s.Alarms, s.Rules)
}
