package kmf1

import (
	"context"
	"fmt"

	"github.com/KohanMathers/KmF1/go/internal/jolpica"
	"github.com/KohanMathers/KmF1/go/internal/livetiming/feeds"
)

// Driver is a handle to a single driver's data within a Session, obtained via Session.Driver.
type Driver struct {
	session      *Session
	driverNumber string
	exists       bool
}

// Exists reports whether the driver was actually found in the session. All other Driver methods return zero values when this is false.
func (d *Driver) Exists() bool { return d.exists }

// Info returns the driver's identifying information (name, team, number) as reported by the live-timing driver list.
func (d *Driver) Info() feeds.DriverInfo {
	if !d.exists {
		return feeds.DriverInfo{}
	}
	if di, ok := d.session.drivers[d.driverNumber]; ok {
		return *di
	}
	return feeds.DriverInfo{}
}

// Result returns the driver's official classification in the session. It reports false if no result was recorded for the driver.
func (d *Driver) Result() (jolpica.SessionResults, bool) {
	if !d.exists {
		return jolpica.SessionResults{}, false
	}
	for _, r := range d.session.results {
		if r.DriverNumber == d.driverNumber {
			return r, true
		}
	}
	return jolpica.SessionResults{}, false
}

// Reference fetches the driver's Jolpica reference record. It requires the driver to have a known Jolpica driver ID, which is only available once the session's official results have been reconciled.
func (d *Driver) Reference(ctx context.Context) (jolpica.DriverReference, error) {
	result, ok := d.Result()
	if !ok || result.DriverId == "" {
		return jolpica.DriverReference{}, fmt.Errorf("kmf1: no Jolpica driver id known for driver %q", d.driverNumber)
	}
	return d.session.client.jolpica.FetchDriverReference(ctx, result.DriverId)
}

// BestStats returns the driver's best lap and sector times recorded during the session. It reports false if no stats were recorded.
func (d *Driver) BestStats() (feeds.DriverBestStats, bool) {
	if !d.exists {
		return feeds.DriverBestStats{}, false
	}
	snapshots := d.session.bestStats[d.driverNumber]
	if len(snapshots) == 0 {
		return feeds.DriverBestStats{}, false
	}
	return snapshots[len(snapshots)-1], true
}

// Laps returns every lap the driver completed during the session.
func (d *Driver) Laps() []feeds.LapData {
	if !d.exists {
		return nil
	}
	return d.session.laps[d.driverNumber]
}

// Lap returns the driver's data for a single lap number. It reports false if the driver has no recorded lap n.
func (d *Driver) Lap(n int) (*Lap, bool) {
	if !d.exists {
		return nil, false
	}
	for _, l := range d.session.laps[d.driverNumber] {
		if l.LapNo == n {
			return &Lap{LapData: l, driver: d}, true
		}
	}
	return nil, false
}

// Telemetry returns every telemetry sample (car data and position) recorded for the driver during the session.
func (d *Driver) Telemetry() []feeds.Telemetry {
	if !d.exists {
		return nil
	}
	return d.session.telemetry[d.driverNumber]
}

// Intervals returns the gap-to-leader/gap-ahead timing intervals recorded for the driver during the session.
func (d *Driver) Intervals() []feeds.TimingInterval {
	if !d.exists {
		return nil
	}
	return d.session.intervals[d.driverNumber]
}

// Stints returns the tyre stints the driver ran during the session.
func (d *Driver) Stints() []feeds.TyreStint {
	if !d.exists {
		return nil
	}
	return d.session.stints[d.driverNumber]
}

// PitStops returns the driver's pit stops during the session, as reported by Jolpica. It requires the driver's Jolpica driver ID to be known via Result.
func (d *Driver) PitStops() []jolpica.PitStop {
	result, ok := d.Result()
	if !ok || result.DriverId == "" {
		return nil
	}
	var out []jolpica.PitStop
	for _, p := range d.session.pitStops {
		if p.DriverId == result.DriverId {
			out = append(out, p)
		}
	}
	return out
}

// TeamRadio returns the team radio clips attributed to the driver during the session.
func (d *Driver) TeamRadio() []feeds.TeamRadio {
	if !d.exists {
		return nil
	}
	var out []feeds.TeamRadio
	for _, r := range d.session.teamRadio {
		if r.DriverNumber == d.driverNumber {
			out = append(out, r)
		}
	}
	return out
}
