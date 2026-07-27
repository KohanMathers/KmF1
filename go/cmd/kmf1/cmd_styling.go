package main

import (
	"context"
	"fmt"
	"strconv"

	kmf1 "github.com/KohanMathers/KmF1/go"
	"github.com/KohanMathers/KmF1/go/internal/styling"
)

func cmdTeams(ctx context.Context, client *kmf1.Client, jsonOut bool, args []string) error {
	if len(args) < 1 || len(args) > 2 {
		return fmt.Errorf("usage: teams <year> [identifier]")
	}
	year, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid year %q: %w", args[0], err)
	}
	season := client.Year(year)

	var teams []styling.TeamStyle
	if len(args) == 2 {
		team, ok := season.TeamStyle(args[1])
		if !ok {
			return fmt.Errorf("no curated team styling matching %q in %d", args[1], year)
		}
		teams = []styling.TeamStyle{team}
	} else {
		teams = season.TeamStyles()
		if teams == nil {
			return fmt.Errorf("no curated team styling data for %d", year)
		}
	}

	headers := []string{"KEY", "SHORT NAME", "OFFICIAL", "FASTF1"}
	rows := make([][]string, len(teams))
	for i, t := range teams {
		rows[i] = []string{t.Key, t.ShortName, t.Colors.Official, t.Colors.FastF1}
	}
	return output(jsonOut, teams, headers, rows)
}

func cmdCompounds(ctx context.Context, client *kmf1.Client, jsonOut bool, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: compounds <year>")
	}
	year, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid year %q: %w", args[0], err)
	}
	season := client.Year(year)

	names := season.Compounds()
	if names == nil {
		return fmt.Errorf("no curated compound color data for %d", year)
	}

	type compound struct {
		Name  string
		Color string
	}
	compounds := make([]compound, len(names))
	headers := []string{"COMPOUND", "COLOR"}
	rows := make([][]string, len(names))
	for i, name := range names {
		color, _ := season.CompoundColor(name)
		compounds[i] = compound{Name: name, Color: color}
		rows[i] = []string{name, color}
	}
	return output(jsonOut, compounds, headers, rows)
}
