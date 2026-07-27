package main

import (
	"context"
	"fmt"

	kmf1 "github.com/KohanMathers/KmF1/go"
	"github.com/KohanMathers/KmF1/go/internal/jolpica"
)

func cmdDriverRefs(ctx context.Context, client *kmf1.Client, jsonOut bool, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("usage: driver-refs [driverId]")
	}

	var refs []jolpica.DriverReference
	if len(args) == 1 {
		ref, err := client.DriverRef(ctx, args[0])
		if err != nil {
			return err
		}
		refs = []jolpica.DriverReference{ref}
	} else {
		var err error
		refs, err = client.DriverRefs(ctx)
		if err != nil {
			return err
		}
	}

	headers := []string{"ID", "ABBR", "NUMBER", "NAME", "NATIONALITY"}
	rows := make([][]string, len(refs))
	for i, r := range refs {
		rows[i] = []string{
			r.DriverId,
			strOrDash(r.Abbreviation),
			intOrDash(r.PermanentNumber),
			r.GivenName + " " + r.FamilyName,
			r.Nationality,
		}
	}
	return output(jsonOut, refs, headers, rows)
}

func cmdConstructors(ctx context.Context, client *kmf1.Client, jsonOut bool, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("usage: constructors [constructorId]")
	}

	var refs []jolpica.ConstructorReference
	if len(args) == 1 {
		ref, err := client.Constructor(ctx, args[0])
		if err != nil {
			return err
		}
		refs = []jolpica.ConstructorReference{ref}
	} else {
		var err error
		refs, err = client.Constructors(ctx)
		if err != nil {
			return err
		}
	}

	headers := []string{"ID", "NAME", "NATIONALITY"}
	rows := make([][]string, len(refs))
	for i, r := range refs {
		rows[i] = []string{r.ConstructorId, r.Name, r.Nationality}
	}
	return output(jsonOut, refs, headers, rows)
}

func cmdCircuits(ctx context.Context, client *kmf1.Client, jsonOut bool, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("usage: circuits [circuitId]")
	}

	var refs []jolpica.CircuitReference
	if len(args) == 1 {
		ref, err := client.CircuitRef(ctx, args[0])
		if err != nil {
			return err
		}
		refs = []jolpica.CircuitReference{ref}
	} else {
		var err error
		refs, err = client.CircuitRefs(ctx)
		if err != nil {
			return err
		}
	}

	headers := []string{"ID", "NAME", "LOCALITY", "COUNTRY"}
	rows := make([][]string, len(refs))
	for i, r := range refs {
		rows[i] = []string{r.CircuitId, r.Name, r.Locality, r.Country}
	}
	return output(jsonOut, refs, headers, rows)
}

func cmdFinishingStatuses(ctx context.Context, client *kmf1.Client, jsonOut bool, args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("usage: statuses")
	}

	statuses, err := client.FinishingStatuses(ctx)
	if err != nil {
		return err
	}

	headers := []string{"ID", "STATUS", "COUNT"}
	rows := make([][]string, len(statuses))
	for i, s := range statuses {
		rows[i] = []string{s.StatusId, s.Status, fmt.Sprint(s.Count)}
	}
	return output(jsonOut, statuses, headers, rows)
}
