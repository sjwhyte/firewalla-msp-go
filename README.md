# firewalla-msp-go

Go client for the [Firewalla MSP API v2](https://help.firewalla.com/hc/en-us/articles/22813088626707).

## Install

```bash
go get github.com/sjwhyte/firewalla-msp-go
```

Requires Go 1.26+.

## Quickstart

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/sjwhyte/firewalla-msp-go"
)

func main() {
    c, err := firewalla.NewClient(
        os.Getenv("FW_DOMAIN"), // e.g. "mydomain.firewalla.net"
        os.Getenv("FW_TOKEN"),  // Personal Access Token from Account Settings
    )
    if err != nil {
        log.Fatal(err)
    }
    ctx := context.Background()

    for alarm, err := range c.Alarms.All(ctx, &firewalla.AlarmListOptions{
        Query:  "status:active",
        SortBy: "ts:desc",
    }) {
        if err != nil {
            log.Fatal(err)
        }
        fmt.Printf("[%s] %s — %s\n", alarm.Type, alarm.Status, alarm.Message)
    }
}
```

## Auth

Generate a Personal Access Token in the MSP UI under **Account Settings → Create New Token**, then pass it to `NewClient`.

## Pagination

Two layers, both backed by the same wire format `{count, results, next_cursor}`.

```go
// Manual cursor — full control.
page, _ := c.Alarms.List(ctx, &firewalla.AlarmListOptions{Query: "status:active"})
for _, a := range page.Results { /* ... */ }
next, _ := c.Alarms.List(ctx, &firewalla.AlarmListOptions{Cursor: page.NextCursor})

// Auto-paginate — Go 1.23+ iterator.
for alarm, err := range c.Alarms.All(ctx, &firewalla.AlarmListOptions{Query: "status:active"}) {
    if err != nil { return err }
    // ...
}
```

## Query builder

```go
import "github.com/sjwhyte/firewalla-msp-go/query"

q := query.And(
    query.AlarmStatus(query.AlarmActive),
    query.BoxName("Gold Plus", "Purple"),
    query.DeviceNameLike("iphone"),
    query.TransferTotalGT("50MB"),
)
c.Alarms.All(ctx, &firewalla.AlarmListOptions{Query: q.String()})
```

Raw query strings are always accepted directly:

```go
c.Alarms.All(ctx, &firewalla.AlarmListOptions{
    Query: `box.name:"Gold Plus" device.name:*iphone* transfer.total:>50MB`,
})
```

## Errors

```go
if err := c.Alarms.Delete(ctx, gid, aid); err != nil {
    switch {
    case firewalla.IsNotFound(err):
        // already gone
    case firewalla.IsUnauthorized(err):
        log.Fatal("bad token")
    case firewalla.IsRateLimited(err):
        // back off — or wrap your client with retry middleware (see below)
    default:
        return err
    }
}
```

The underlying error is a `*firewalla.APIError` with `HTTPStatus`, `Message`, `Code`, `Body`, `Method`, `URL`.

## Retries (opt-in)

```go
import "github.com/sjwhyte/firewalla-msp-go/retry"

hc := &http.Client{Transport: retry.NewTransport(http.DefaultTransport, retry.Config{})}
c, _ := firewalla.NewClient(domain, token, firewalla.WithHTTPClient(hc))
```

Retries: network errors, 429, 502, 503, 504. Honors `Retry-After`. Skips retry if a request body is non-replayable.

## Resources covered

`Boxes`, `Devices`, `Alarms`, `Flows`, `Rules`, `TargetLists`, `Stats`, `Trends` — every endpoint in the MSP v2 Postman collection.

## Development

Common workflows are wrapped in [go-task](https://taskfile.dev). Install with `brew install go-task` (macOS) or see https://taskfile.dev/installation/.

- `task` — run unit tests
- `task ci` — run the full CI pipeline locally (fmt check, vet, race tests, staticcheck)
- `task --list` — see all available tasks

## Releases

1. Bump `const Version` in `client.go`.
2. Commit. Run `task ci`.
3. Tag: `git tag vX.Y.Z`. Verify locally: `GIT_TAG=vX.Y.Z task release:check`.
4. Push: `git push origin main vX.Y.Z`.

The release workflow validates the tag, re-runs CI, and creates a GitHub Release with auto-generated notes.

## License

MIT
