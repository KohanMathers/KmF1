package main

import (
	"context"
	"fmt"
	"strconv"

	kmf1 "github.com/KohanMathers/KmF1/go"
	"github.com/KohanMathers/KmF1/go/internal/jolpica"
)

func cmdEvents(ctx context.Context, client *kmf1.Client, jsonOut bool, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: events <year>")
	}
	year, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid year %q: %w", args[0], err)
	}

	events, err := client.Year(year).Events(ctx)
	if err != nil {
		return err
	}

	infos := make([]jolpica.Schedule, len(events))
	headers := []string{"ROUND", "NAME", "LOCATION", "COUNTRY", "FORMAT"}
	rows := make([][]string, len(events))
	for i, e := range events {
		infos[i] = e.Info()
		rows[i] = []string{
			strconv.Itoa(infos[i].RoundNumber),
			infos[i].EventName,
			infos[i].Location,
			infos[i].Country,
			infos[i].EventFormat,
		}
	}
	return output(jsonOut, infos, headers, rows)
}
