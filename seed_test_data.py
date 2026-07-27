#!/usr/bin/env python3
from __future__ import annotations

import sys
import time
import urllib.error
import urllib.request
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent
GO_ROOT = REPO_ROOT / "go"

LIVETIMING_BASE = "https://livetiming.formula1.com/static"
JOLPICA_BASE = "https://api.jolpi.ca/ergast/f1"
MULTIVIEWER_BASE = "https://api.multiviewer.app/api/v1/circuits"

RACE_SESSION = "2024/2024-06-23_Spanish_Grand_Prix/2024-06-23_Race"
QUALIFYING_SESSION = "2024/2024-06-23_Spanish_Grand_Prix/2024-06-22_Qualifying"

FEEDS_TESTDATA = GO_ROOT / "internal/livetiming/feeds/testdata"
WIRE_TESTDATA = GO_ROOT / "internal/livetiming/wire/testdata"
LIVETIMING_TESTDATA = GO_ROOT / "internal/livetiming/testdata"
SESSION_TESTDATA = GO_ROOT / "internal/session/testdata"
JOLPICA_TESTDATA = GO_ROOT / "internal/jolpica/testdata"
MULTIVIEWER_TESTDATA = GO_ROOT / "internal/multiviewer/testdata"

def take_first_records(data: bytes, n: int) -> bytes:
    """Equivalent to `head -n N` on a record-per-line wire file — the BOM
    (part of the first record's leading bytes) is left untouched."""
    parts = data.split(b"\r\n")
    return b"\r\n".join(parts[:n]) + b"\r\n"

def filter_elapsed_window(data: bytes, start_seconds: float, end_seconds: float) -> bytes:
    """Keeps only records whose frame-elapsed prefix falls in
    [start_seconds, end_seconds)."""
    bom = b"\xef\xbb\xbf"
    body = data[len(bom):] if data.startswith(bom) else data
    kept = []
    for line in body.split(b"\r\n"):
        if not line.strip():
            continue
        h, m, s = line[:12].decode("ascii").split(":")
        total = int(h) * 3600 + int(m) * 60 + float(s)
        if start_seconds <= total < end_seconds:
            kept.append(line)
    return bom + b"\r\n".join(kept) + b"\r\n"

def first(n: int):
    return ("first", n)

def window(start_minute: int):
    return ("window", start_minute * 60, start_minute * 60 + 60)

def apply_transform(data: bytes, transform) -> bytes:
    if transform is None:
        return data
    kind = transform[0]
    if kind == "first":
        return take_first_records(data, transform[1])
    if kind == "window":
        return filter_elapsed_window(data, transform[1], transform[2])
    raise ValueError(f"unknown transform {transform!r}")

DOWNLOADS: list[tuple[str, list[tuple[Path, object]]]] = [
    (f"{LIVETIMING_BASE}/{RACE_SESSION}/SessionInfo.jsonStream",
     [(FEEDS_TESTDATA / "sessioninfo.jsonstream", None)]),
    (f"{LIVETIMING_BASE}/{RACE_SESSION}/DriverList.jsonStream",
     [(FEEDS_TESTDATA / "driverlist.jsonstream", None)]),
    (f"{LIVETIMING_BASE}/{RACE_SESSION}/TrackStatus.jsonStream",
     [(FEEDS_TESTDATA / "trackstatus.jsonstream", None)]),
    (f"{LIVETIMING_BASE}/{RACE_SESSION}/SessionStatus.jsonStream",
     [(FEEDS_TESTDATA / "sessionstatus.jsonstream", None)]),
    (f"{LIVETIMING_BASE}/{RACE_SESSION}/LapCount.jsonStream",
     [(FEEDS_TESTDATA / "lapcount.jsonstream", None)]),
    (f"{LIVETIMING_BASE}/{RACE_SESSION}/TimingData.jsonStream",
     [(FEEDS_TESTDATA / "timingdata.jsonstream", first(3700))]),
    (f"{LIVETIMING_BASE}/{RACE_SESSION}/TimingAppData.jsonStream",
     [(FEEDS_TESTDATA / "timingappdata.jsonstream", None)]),
    (f"{LIVETIMING_BASE}/{RACE_SESSION}/CarData.z.jsonStream",
     [(FEEDS_TESTDATA / "cardata.jsonstream", window(59)),
      (WIRE_TESTDATA / "cardata.jsonstream", first(3))]),
    (f"{LIVETIMING_BASE}/{RACE_SESSION}/Position.z.jsonStream",
     [(FEEDS_TESTDATA / "position.jsonstream", window(59))]),
    (f"{LIVETIMING_BASE}/{RACE_SESSION}/WeatherData.jsonStream",
     [(FEEDS_TESTDATA / "weatherdata.jsonstream", None)]),
    (f"{LIVETIMING_BASE}/{RACE_SESSION}/RaceControlMessages.jsonStream",
     [(FEEDS_TESTDATA / "racecontrolmessages.jsonstream", None)]),
    (f"{LIVETIMING_BASE}/{RACE_SESSION}/TimingStats.jsonStream",
     [(FEEDS_TESTDATA / "timingstats.jsonstream", None)]),
    (f"{LIVETIMING_BASE}/{RACE_SESSION}/TeamRadio.jsonStream",
     [(FEEDS_TESTDATA / "teamradio.jsonstream", first(6)),
      (WIRE_TESTDATA / "teamradio.jsonstream", first(6))]),
    (f"{LIVETIMING_BASE}/{RACE_SESSION}/TopThree.jsonStream",
     [(FEEDS_TESTDATA / "topthree.jsonstream", first(6)),
      (WIRE_TESTDATA / "topthree.jsonstream", first(6))]),
    (f"{LIVETIMING_BASE}/{RACE_SESSION}/ChampionshipPrediction.jsonStream",
     [(FEEDS_TESTDATA / "championshipprediction.jsonstream", None)]),
    (f"{LIVETIMING_BASE}/{RACE_SESSION}/ExtrapolatedClock.jsonStream",
     [(FEEDS_TESTDATA / "extrapolatedclock.jsonstream", None),
      (WIRE_TESTDATA / "extrapolatedclock.jsonstream", None)]),
    (f"{LIVETIMING_BASE}/{RACE_SESSION}/ArchiveStatus.json",
     [(FEEDS_TESTDATA / "archivestatus.json", None),
      (WIRE_TESTDATA / "archivestatus.json", None)]),
    (f"{LIVETIMING_BASE}/{RACE_SESSION}/AudioStreams.jsonStream",
     [(FEEDS_TESTDATA / "audiostreams.jsonstream", None)]),
    (f"{LIVETIMING_BASE}/{RACE_SESSION}/ContentStreams.jsonStream",
     [(FEEDS_TESTDATA / "contentstreams.jsonstream", None)]),
    (f"{LIVETIMING_BASE}/{RACE_SESSION}/SessionData.json",
     [(WIRE_TESTDATA / "sessiondata.json", None)]),
    (f"{LIVETIMING_BASE}/{QUALIFYING_SESSION}/SessionStatus.jsonStream",
     [(SESSION_TESTDATA / "qualifying_sessionstatus.jsonstream", None)]),
    (f"{LIVETIMING_BASE}/2024/Index.json",
     [(LIVETIMING_TESTDATA / "index2024.json", None)]),
    (f"{JOLPICA_BASE}/2024.json",
     [(JOLPICA_TESTDATA / "schedule_2024.json", None)]),
    (f"{JOLPICA_BASE}/2024/drivers.json?limit=30",
     [(JOLPICA_TESTDATA / "drivers_2024.json", None)]),
    (f"{JOLPICA_BASE}/2024/constructors.json?limit=15",
     [(JOLPICA_TESTDATA / "constructors_2024.json", None)]),
    (f"{JOLPICA_BASE}/2024/circuits.json?limit=30",
     [(JOLPICA_TESTDATA / "circuits_2024.json", None)]),
    (f"{JOLPICA_BASE}/2024/status.json",
     [(JOLPICA_TESTDATA / "status_2024.json", None)]),
    (f"{JOLPICA_BASE}/2024/9/results.json",
     [(JOLPICA_TESTDATA / "results_2024_9.json", None)]),
    (f"{JOLPICA_BASE}/2024/9/qualifying.json",
     [(JOLPICA_TESTDATA / "qualifying_2024_9.json", None)]),
    (f"{JOLPICA_BASE}/2024/6/sprint.json",
     [(JOLPICA_TESTDATA / "sprint_2024_6.json", None)]),
    (f"{JOLPICA_BASE}/2024/9/driverStandings.json",
     [(JOLPICA_TESTDATA / "driverstandings_2024_9.json", None)]),
    (f"{JOLPICA_BASE}/2024/9/constructorStandings.json",
     [(JOLPICA_TESTDATA / "constructorstandings_2024_9.json", None)]),
    (f"{JOLPICA_BASE}/2024/9/pitstops.json",
     [(JOLPICA_TESTDATA / "pitstops_2024_9.json", None)]),
    (f"{JOLPICA_BASE}/2024/9/laps/1.json",
     [(JOLPICA_TESTDATA / "laps1_2024_9.json", None)]),
    (f"{JOLPICA_BASE}/drivers/max_verstappen.json",
     [(JOLPICA_TESTDATA / "driver_max_verstappen.json", None)]),
    (f"{JOLPICA_BASE}/constructors/red_bull.json",
     [(JOLPICA_TESTDATA / "constructor_red_bull.json", None)]),
    (f"{JOLPICA_BASE}/circuits/monza.json",
     [(JOLPICA_TESTDATA / "circuit_monza.json", None)]),
    (f"{JOLPICA_BASE}/status.json?limit=1000",
     [(JOLPICA_TESTDATA / "status_alltime.json", None)]),
    (f"{MULTIVIEWER_BASE}/15/2024",
     [(MULTIVIEWER_TESTDATA / "catalunya_2024.json", None)]),
    (f"{MULTIVIEWER_BASE}/4/2024",
     [(MULTIVIEWER_TESTDATA / "hungaroring_2024.json", None)]),
]


def fetch(url: str) -> bytes:
    req = urllib.request.Request(url, headers={"User-Agent": "KmF1-seed-test-data"})
    with urllib.request.urlopen(req, timeout=30) as resp:
        return resp.read()


def write(dest: Path, data: bytes) -> None:
    dest.parent.mkdir(parents=True, exist_ok=True)
    dest.write_bytes(data)
    print(f"  wrote {dest.relative_to(REPO_ROOT)} ({len(data)} bytes)")


def main() -> int:
    failures = []
    for i, (url, outputs) in enumerate(DOWNLOADS, 1):
        print(f"[{i}/{len(DOWNLOADS)}] GET {url}")
        try:
            data = fetch(url)
        except (urllib.error.URLError, TimeoutError) as e:
            print(f"  FAILED: {e}")
            failures.append(url)
            continue

        for dest, transform in outputs:
            write(dest, apply_transform(data, transform))

        time.sleep(0.2)

    if failures:
        print(f"\n{len(failures)} download(s) failed:")
        for url in failures:
            print(f"  {url}")
        return 1

    print("\nAll fixtures seeded successfully.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
