// Error-handling demonstrates the typed error predicates. It intentionally
// calls Alarms.Get with a nonexistent GID/AID pair so the server returns
// 404; the program prints how to recognize and react to each common error
// class. Safe to run — no mutations.
//
// Run with:
//
//	FW_DOMAIN=... FW_TOKEN=... go run ./examples/error-handling
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/sjwhyte/firewalla-msp-go"
)

func main() {
	c, err := firewalla.NewClient(os.Getenv("FW_DOMAIN"), os.Getenv("FW_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	_, err = c.Alarms.Get(context.Background(),
		"00000000-0000-0000-0000-deadbeef0000",
		firewalla.AlarmID("999999999"),
	)
	classify(err)
}

func classify(err error) {
	switch {
	case err == nil:
		fmt.Println("no error (unexpected — the bogus alarm should not exist)")
	case firewalla.IsNotFound(err):
		fmt.Println("not found (expected): caller can ignore or retry with a different ID")
	case firewalla.IsUnauthorized(err):
		fmt.Println("unauthorized: token is bad or missing scope")
	case firewalla.IsRateLimited(err):
		fmt.Println("rate-limited: back off or use the retry subpackage")
	default:
		var ae *firewalla.APIError
		if errors.As(err, &ae) {
			fmt.Printf("other api error: status=%d code=%s msg=%s\n",
				ae.HTTPStatus, ae.Code, ae.Message)
			return
		}
		fmt.Printf("non-api error: %v\n", err)
	}
}
