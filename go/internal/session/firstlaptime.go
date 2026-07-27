package session

import (
	"github.com/KohanMathers/KmF1/go/internal/jolpica"
	"github.com/KohanMathers/KmF1/go/internal/livetiming/feeds"
)

func BackfillFirstLapTime(laps map[string][]feeds.LapData, results []jolpica.SessionResults, lap1Times []jolpica.LapTiming) {
	if len(lap1Times) == 0 {
		return
	}

	driverNumberByDriverId := make(map[string]string, len(results))
	for _, r := range results {
		if r.DriverId != "" {
			driverNumberByDriverId[r.DriverId] = r.DriverNumber
		}
	}

	timeByDriverNumber := make(map[string]jolpica.LapTiming, len(lap1Times))
	for _, t := range lap1Times {
		if drv, ok := driverNumberByDriverId[t.DriverId]; ok {
			timeByDriverNumber[drv] = t
		}
	}

	for drv, ls := range laps {
		t, ok := timeByDriverNumber[drv]
		if !ok {
			continue
		}
		for i := range ls {
			if ls[i].LapNo == 1 && ls[i].LapTime == nil {
				lapTime := t.Time
				ls[i].LapTime = &lapTime
			}
		}
	}
}
