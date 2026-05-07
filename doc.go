// Package firewalla provides a Go client for the Firewalla MSP API v2.
//
// Construct a client with a domain and Personal Access Token, then call
// methods on its sub-services:
//
//	c, err := firewalla.NewClient("mydomain.firewalla.net", os.Getenv("FW_TOKEN"))
//	if err != nil {
//	    return err
//	}
//	for alarm, err := range c.Alarms.All(ctx, &firewalla.AlarmListOptions{
//	    Query: "status:active",
//	}) {
//	    if err != nil {
//	        return err
//	    }
//	    // ... use alarm
//	}
//
// See the project README and the godoc for each sub-service for details.
package firewalla
