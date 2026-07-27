# KmF1

A Go library for Formula 1 timing, telemetry, and historical data.

KmF1 pulls together F1's live-timing archive, the [Jolpica](https://api.jolpi.ca/ergast/f1) (Ergast-compatible) historical results API, and the [MultiViewer](https://api.multiviewer.app) circuit geometry API into one client, with a persistent on-disk cache so repeated loads don't re-fetch data that hasn't changed.

## Features

- Season / event / session model: `client.Year(2024).Event(...).Session(...)` walks from a season down to a single practice/qualifying/sprint/race session, loading and reconciling results, laps, telemetry, stints, intervals, weather, track status, race control messages, team radio, and pit stops.
- Historical results via Jolpica, including driver/constructor/circuit reference tables, championship standings, and finishing statuses.
- Circuit geometry via MultiViewer: corners, marshal lights, marshal sectors.
- Team and tyre-compound styling metadata (colours, short names) per season, with no charting dependency.
- Live push: connect to F1's real-time SignalR feed and stream raw messages as they arrive (`Client.Live`), decoded the same way the historical archive is.
- Persistent sqlite cache for every backend, keyed by URL and parser version so a cache built by an older version of KmF1 is invalidated automatically instead of silently misparsing.
- `kmf1` CLI: a thin command-line front end over the library for quick lookups without writing Go.

## Install

```sh
go get github.com/KohanMathers/KmF1/go
```

## Quick start

```go
package main

import (
    "context"
    "fmt"

    kmf1 "github.com/KohanMathers/KmF1/go"
)

func main() {
    ctx := context.Background()

    client, err := kmf1.New("cache.sqlite") // ":memory:" disables caching
    if err != nil {
        panic(err)
    }
    defer client.Close()

    event, err := client.Year(2024).Event(ctx, "Monza")
    if err != nil {
        panic(err)
    }

    ref, err := event.Session("Race")
    if err != nil {
        panic(err)
    }

    session, err := ref.Load(ctx)
    if err != nil {
        panic(err)
    }

    for _, result := range session.Results() {
        fmt.Println(result.PositionText, result.DriverId, result.Status)
    }
}
```

`session.Driver(driverId)` drills into a single driver's laps, telemetry, stints, and pit stops. `event.Circuit()` gets circuit geometry, and `event.Standings()` gets championship standings.

### Live push

```go
out := make(chan kmf1.LiveMessage)
go func() {
    for msg := range out {
        // msg.Payload is raw wire data; decode with the same
        // wire.Decode path the historical archive uses.
        fmt.Println(msg.Topic)
    }
}()

if err := client.Live(ctx, out, kmf1.LiveOptions{}); err != nil {
    panic(err)
}
```

`Live` blocks until `ctx` is canceled, reconnecting automatically on any drop. It's a recorder, not an in-process live session builder. There's no typed, incrementally-updated equivalent of `Session` for live data yet.

## CLI

```sh
cd go
go build -o kmf1 ./cmd/kmf1

./kmf1 events 2024
./kmf1 session 2024 Monza Race
./kmf1 results 2024 Monza Race
./kmf1 laps 2024 Monza Race VER
./kmf1 drivers 2024 Monza Race
./kmf1 weather 2024 Monza Race
./kmf1 driver-refs [driverId]
./kmf1 constructors [constructorId]
./kmf1 circuits [circuitId]
./kmf1 statuses
./kmf1 teams 2024 [identifier]
./kmf1 compounds 2024
```

Add `-json` to any command for machine-readable output, `-v` to log fetch/parse progress to stderr, and `-cache path` to point at a specific sqlite cache file (defaults to the OS cache dir, e.g. `~/.cache/kmf1/cache.sqlite`).

## Package layout

The public API (`go/*.go`) is a thin, typed façade over a handful of internal packages, each scoped to one concern:

| Package                        | Responsibility                                                        |
| ------------------------------- | ----------------------------------------------------------------------- |
| `internal/transport`            | Shared HTTP client, rate limiting, base-URL/mirror fallback            |
| `internal/cache`                | Persistent sqlite response cache, keyed by URL + parser version        |
| `internal/jolpica`               | Historical results, standings, and reference-table API client          |
| `internal/multiviewer`           | Circuit geometry API client                                            |
| `internal/schedule`              | Season calendar resolution (live-timing index, falling back to Jolpica)|
| `internal/session`               | Cross-backend reconciliation: results, laps, derived classifications   |
| `internal/styling`               | Per-season team/compound color and short-name metadata                 |
| `internal/livetiming/feeds`      | Per-topic decoders for the live-timing archive (laps, weather, etc.)   |
| `internal/livetiming/live`       | Real-time SignalR connection, reconnect policy, time anchoring         |
| `internal/livetiming/signalr`    | Minimal generic SignalR Core JSON hub protocol client                  |
| `internal/livetiming/wire`       | Raw archive payload decoding (zlib-compressed and plain JSON)          |

## Testing

Most of the Go test suite depends on real, historical fixture data that isn't committed to the repo (kept out to keep it lean). Seed it once after cloning:

```sh
python3 seed_test_data.py
cd go && go test ./...
```

## License

MIT — see [LICENSE](./LICENSE).
