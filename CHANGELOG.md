# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html). While the major version is `0`, breaking changes may land in minor (`0.x`) releases.

## [Unreleased]

## [v0.5.0] - 2026-06-01

### Added
- `TargetListListOptions{Owner}` filter on `TargetLists.List`, matching the documented `owner` query parameter on `GET /v2/target-lists`. Pass a box gid (or comma-separated gids) to include box-owned lists; the API otherwise returns only global and Firewalla-managed lists.

### Changed
- **Breaking:** `TargetLists.List` signature changed from `List(ctx)` to `List(ctx, opts *TargetListListOptions)`. Existing callers should pass `nil` to preserve current behavior.

## [v0.4.0] - 2026-05-13

### Added
- `examples/` directory with nine runnable programs: `tour`, `pagination`, `query`, `with-retry`, `error-handling`, `rules-lifecycle`, `target-lists-lifecycle`, `device-rename`, `alarm-delete`.
- `examples/README.md` catalog index.
- `task build:examples` target, wired into `task ci`, that compiles every example to catch API drift on every CI run.
- Root `README.md` "Examples" section pointing at the catalog.

## [v0.3.3] - 2026-05-13

### Changed
- `NewClient` now `strings.TrimSpace`s the `domain` and `token` arguments before validating, so a trailing newline from a heredoc-pasted token (or surrounding whitespace) is forgiven. Whitespace-only inputs still fail the existing empty-string checks.

### Documentation
- `WithHTTPClient` godoc now notes that the caller owns TLS policy when overriding the client and warns against `InsecureSkipVerify`.

## [v0.3.2] - 2026-05-13

### Security
- `query.Eq` / `query.Like` value escaping now triggers on `"` and `\` and escapes both correctly. Previously, caller-supplied values containing a bare double quote or backslash could break out of the quoted region in a Firewalla search expression.
- `Client.do` now reads at most `DefaultMaxResponseBodyBytes` (32 MiB) of any response body, returning an error on overflow rather than consuming unbounded memory. New `WithMaxResponseBodyBytes(n int64)` Option to override the limit.
- `paginate` detects cursor cycles and returns an error instead of looping forever when a server returns the same `NextCursor` it just consumed.

## [v0.3.1] - 2026-05-11

### Fixed
- `Box.Group` and `Device.Group` now accept either a JSON string (treated as the group id) or a JSON object with `{id, name}`. Previously, observing a string from `/v2/boxes` failed decoding.

### Changed
- `BoxGroup.ID` and `DeviceGroup.ID` types tightened from `any` to `string` with deterministic coercion for numeric ids (formatted as decimal text, no trailing `.0`).

## [v0.3.0] - 2026-05-11

### Added
- New `Timestamp` type that decodes either an integer or floating-point epoch-seconds JSON token, preserving sub-second precision via decimal-text parsing (no float64 round-trip).
- `Alarm` fields surfaced from real API responses: `TypeName` (`_type`), `ActiveTS`, `CloudAction`, `Direction`, `Count`, `Analytics` (`json.RawMessage`).
- `AlarmDevice` fields: `IP`, `MACVendor`, `Port`, `Network`, `DeviceType`, `PortInfo`.
- `AlarmRemote` fields: `IP`, `Latitude`, `Longitude`, `Port`, `RootDomain`, `PortInfo`.
- `AlarmTransfer` field: `Duration` (float seconds).
- New `AlarmPortInfo` type for device/remote port metadata.

### Changed
- **Breaking:** `Alarm.TS`, `Box.LastSeen`, `Device.LastSeen`, `Rule.CreatedAt`, `Rule.UpdatedAt`, `TargetList.LastUpdated`, `Flow.TS`, and `TrendPoint.TS` switch from `int64` to `Timestamp`. `Timestamp` embeds `time.Time`, so `.Format(...)`, `.Unix()`, etc. all work via method promotion; callers who assigned to a plain `int64` need `t.Unix()` or similar.
- `TrendPoint`'s bespoke `int64`-based `UnmarshalJSON` removed in favor of plain struct tag decoding against `Timestamp`.

## [v0.2.1] - 2026-05-11

### Fixed
- `Alarm.Status` now accepts either a JSON number (the real MSP API's behavior) or a JSON string. New `AlarmStatus` wrapper type handles both; existing string comparisons (`alarm.Status == "active"`) still work.

## [v0.2.0] - 2026-05-11

### Fixed
- `Alarm.AID` now accepts either a JSON number (the real MSP API's behavior) or a JSON string. New `AlarmID` wrapper type handles both.

### Changed
- **Breaking:** `Alarm.AID` field type: `string` → `AlarmID`.
- **Breaking:** `AlarmsService.Get` and `AlarmsService.Delete` parameter `aid`: `string` → `AlarmID`. String literals still convert implicitly; passing a `string` variable needs `firewalla.AlarmID(s)`.

## [v0.1.0] - 2026-05-08

Initial public release.

### Added
- Resource services for every endpoint in the Firewalla MSP API v2 Postman collection: `Boxes`, `Devices`, `Alarms`, `Flows`, `Rules`, `TargetLists`, `Stats`, `Trends`.
- Generic `Page[T]` and an `iter.Seq2[T, error]` auto-paginator (`Alarms.All`, `Flows.All`, `Rules.All`) for cursor-based list endpoints.
- `query` subpackage with `And`, `Or`, `Not`, `Eq`, `Like`, generic `GT`/`GTE`/`LT`/`LTE`, `Range`, and typed helpers for alarm/flow/rule qualifiers.
- `retry` subpackage exposing an opt-in `http.RoundTripper` with exponential backoff, `Retry-After` support, and safe body replay via `Request.GetBody`.
- `APIError` plus `IsNotFound`, `IsUnauthorized`, `IsRateLimited` predicates that work through `errors.As`.
- Functional options: `WithHTTPClient`, `WithBaseURL`, `WithUserAgent`.
- godoc-embedded `Example*` functions in `examples_test.go`.
- Taskfile and GitHub Actions CI running gofmt, go vet, race tests, and staticcheck.
- Release workflow that triggers on `v*` tags, validates the tag against `const Version`, re-runs CI, and creates a GitHub Release.

[Unreleased]: https://github.com/sjwhyte/firewalla-msp-go/compare/v0.5.0...HEAD
[v0.5.0]: https://github.com/sjwhyte/firewalla-msp-go/releases/tag/v0.5.0
[v0.4.0]: https://github.com/sjwhyte/firewalla-msp-go/releases/tag/v0.4.0
[v0.3.3]: https://github.com/sjwhyte/firewalla-msp-go/releases/tag/v0.3.3
[v0.3.2]: https://github.com/sjwhyte/firewalla-msp-go/releases/tag/v0.3.2
[v0.3.1]: https://github.com/sjwhyte/firewalla-msp-go/releases/tag/v0.3.1
[v0.3.0]: https://github.com/sjwhyte/firewalla-msp-go/releases/tag/v0.3.0
[v0.2.1]: https://github.com/sjwhyte/firewalla-msp-go/releases/tag/v0.2.1
[v0.2.0]: https://github.com/sjwhyte/firewalla-msp-go/releases/tag/v0.2.0
[v0.1.0]: https://github.com/sjwhyte/firewalla-msp-go/releases/tag/v0.1.0
