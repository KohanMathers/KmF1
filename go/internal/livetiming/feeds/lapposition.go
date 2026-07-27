package feeds

import "sort"

func CalculatePositions(laps map[string][]LapData, raceLike bool) {
	if !raceLike {
		return
	}

	dnfGenerated := make(map[string]bool)
	for drv, ls := range laps {
		if len(ls) == 1 && ls[0].FastF1Generated {
			dnfGenerated[drv] = true
		}
	}

	byLapNo := make(map[int][]*LapData)
	for drv, ls := range laps {
		for i := range ls {
			if dnfGenerated[drv] {
				ls[i].Position = nil
				continue
			}
			byLapNo[ls[i].LapNo] = append(byLapNo[ls[i].LapNo], &ls[i])
		}
	}

	for _, group := range byLapNo {
		sort.Slice(group, func(i, j int) bool { return group[i].Timestamp < group[j].Timestamp })
		for i, l := range group {
			pos := i + 1
			l.Position = &pos
		}
	}
}
