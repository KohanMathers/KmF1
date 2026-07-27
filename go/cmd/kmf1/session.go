package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	kmf1 "github.com/KohanMathers/KmF1/go"
)

var sessionTypes = []string{
	kmf1.SessionPractice1,
	kmf1.SessionPractice2,
	kmf1.SessionPractice3,
	kmf1.SessionQualifying,
	kmf1.SessionSprintQualifying,
	kmf1.SessionSprint,
	kmf1.SessionRace,
}

func resolveSessionType(input string) (string, error) {
	for _, s := range sessionTypes {
		if strings.EqualFold(s, input) {
			return s, nil
		}
	}
	return "", fmt.Errorf("unknown session type %q (want one of: Practice 1, Practice 2, Practice 3, Qualifying, Sprint Qualifying, Sprint, Race)", input)
}

func eventIdentifier(raw string) any {
	if n, err := strconv.Atoi(raw); err == nil {
		return n
	}
	return raw
}

func loadSession(ctx context.Context, client *kmf1.Client, args []string) (*kmf1.Session, []string, error) {
	if len(args) < 3 {
		return nil, nil, fmt.Errorf("want <year> <round|event name> <sessionType> [extra args...], got %d args", len(args))
	}

	year, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, nil, fmt.Errorf("invalid year %q: %w", args[0], err)
	}
	sessionType, err := resolveSessionType(args[2])
	if err != nil {
		return nil, nil, err
	}

	event, err := client.Year(year).Event(ctx, eventIdentifier(args[1]))
	if err != nil {
		return nil, nil, err
	}
	ref, err := event.Session(sessionType)
	if err != nil {
		return nil, nil, err
	}
	session, err := ref.Load(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("load session: %w", err)
	}
	return session, args[3:], nil
}
