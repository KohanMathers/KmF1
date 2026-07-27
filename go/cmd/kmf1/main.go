package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	kmf1 "github.com/KohanMathers/KmF1/go"
)

type command func(ctx context.Context, client *kmf1.Client, jsonOut bool, args []string) error

var commands = map[string]command{
	"events":       cmdEvents,
	"session":      cmdSession,
	"results":      cmdResults,
	"laps":         cmdLaps,
	"drivers":      cmdDrivers,
	"weather":      cmdWeather,
	"driver-refs":  cmdDriverRefs,
	"constructors": cmdConstructors,
	"circuits":     cmdCircuits,
	"statuses":     cmdFinishingStatuses,
	"teams":        cmdTeams,
	"compounds":    cmdCompounds,
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "kmf1:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("kmf1", flag.ContinueOnError)
	cachePath := fs.String("cache", defaultCachePath(), `sqlite cache file path (":memory:" for none)`)
	jsonOut := fs.Bool("json", false, "output JSON instead of a table")
	verbose := fs.Bool("v", false, "log Load's fetch/parse progress to stderr")
	fs.Usage = printUsage
	if err := fs.Parse(args); err != nil {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		printUsage()
		return fmt.Errorf("no command given")
	}
	name, cmdArgs := rest[0], rest[1:]
	cmd, ok := commands[name]
	if !ok {
		printUsage()
		return fmt.Errorf("unknown command %q", name)
	}

	var opts []kmf1.Option
	if *verbose {
		opts = append(opts, kmf1.WithLogger(slog.New(slog.NewTextHandler(os.Stderr, nil))))
	}
	client, err := kmf1.New(*cachePath, opts...)
	if err != nil {
		return fmt.Errorf("open client (cache %q): %w", *cachePath, err)
	}
	defer client.Close()

	return cmd(context.Background(), client, *jsonOut, cmdArgs)
}

func defaultCachePath() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "kmf1-cache.sqlite"
	}
	full := filepath.Join(dir, "kmf1", "cache.sqlite")
	_ = os.MkdirAll(filepath.Dir(full), 0o755)
	return full
}

func printUsage() {
	fmt.Fprint(os.Stderr, `kmf1 — command-line access to the KmF1 F1 data library

Usage:
  kmf1 [-cache path] [-json] [-v] <command> [args...]

Commands:
  events <year>
  session <year> <round|event name> <sessionType>
  results <year> <round|event name> <sessionType>
  laps <year> <round|event name> <sessionType> [driver]
  drivers <year> <round|event name> <sessionType>
  weather <year> <round|event name> <sessionType>
  driver-refs [driverId]
  constructors [constructorId]
  circuits [circuitId]
  statuses
  teams <year> [identifier]
  compounds <year>

sessionType is one of (case-insensitive): "Practice 1", "Practice 2",
"Practice 3", "Qualifying", "Sprint Qualifying", "Sprint", "Race".

Flags:
  -cache path   sqlite cache file path (":memory:" for none; default is
                the OS cache dir, e.g. ~/.cache/kmf1/cache.sqlite)
  -json         output JSON instead of a table
  -v            log Load's fetch/parse progress to stderr
`)
}
