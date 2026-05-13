# Examples

Runnable programs that demonstrate how to use [firewalla-msp-go](../).

All examples expect these environment variables:

```bash
export FW_DOMAIN=mydomain.firewalla.net
export FW_TOKEN=...                           # MSP Personal Access Token
```

Examples marked ⚠ make real, persistent changes to your MSP.

## Read-only

| Example | What it does |
| --- | --- |
| [tour](./tour) | Multi-resource walkthrough: stats, boxes, devices, alarms, flows, rules, target lists |
| [pagination](./pagination) | Manual cursor loop and `Alarms.All` iterator side-by-side |
| [query](./query) | Build a Firewalla search expression with the `query` subpackage |
| [with-retry](./with-retry) | Wrap the client with the opt-in retry middleware |
| [error-handling](./error-handling) | `IsNotFound`, `IsUnauthorized`, `IsRateLimited`, and `*APIError` patterns |

Run any of these:

```bash
go run ./examples/tour
go run ./examples/pagination
go run ./examples/query
go run ./examples/with-retry
go run ./examples/error-handling
```

## Write (⚠)

| Example | What it does |
| --- | --- |
| [rules-lifecycle](./rules-lifecycle) | Creates a block rule for `example.invalid`, pauses, resumes, deletes |
| [target-lists-lifecycle](./target-lists-lifecycle) | Creates a `msp-go-example-<ts>` target list, gets, updates, deletes |
| [device-rename](./device-rename) | Renames one device (one-shot; takes `-gid -device -name`) |
| [alarm-delete](./alarm-delete) | Permanently deletes one alarm (one-shot; takes `-gid -aid`) |

The two `-lifecycle` examples clean up after themselves. The two one-shot examples do not — you provide the target, you own restoring it.

Examples:

```bash
go run ./examples/rules-lifecycle -gid <box-gid>
go run ./examples/target-lists-lifecycle
go run ./examples/device-rename -gid <box-gid> -device mac:AA:BB:CC:DD:EE:FF -name "New Name"
go run ./examples/alarm-delete -gid <box-gid> -aid 2408
```

## CI

Every example here is compiled by `task build:examples` (and therefore `task ci`), so the catalog can't silently rot when the library API changes.
