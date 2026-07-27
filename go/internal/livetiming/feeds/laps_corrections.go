package feeds

import "time"

func postProcess(laps []*lapBuilder, personalBests []pbEntry) ([]*lapBuilder, int) {
	laps = trimEmptyEdges(laps)
	if len(laps) == 0 {
		return laps, 0
	}

	integrityErrors := 0
	integrityErrors += sectorSumCrossCheck(laps)
	reconstructMissingValues(laps)
	integrityErrors += syncLapEndTimes(laps)

	laps = trimUnusableTrailingLap(laps)
	if len(laps) == 0 {
		return laps, integrityErrors
	}

	integrityErrors += backwardSync(laps)
	forwardSync(laps)

	reconstructDeletedPersonalBests(laps, personalBests)

	correctPitStopOffset(laps)
	correctFirstLapPitIn(laps)

	return laps, integrityErrors
}

func hasMeaningfulData(l *lapBuilder) bool {
	return l.sector1 != nil || l.sector2 != nil || l.sector3 != nil ||
		l.lapTime != nil ||
		l.speedI1 != nil || l.speedI2 != nil || l.speedFL != nil || l.speedST != nil
}

func trimEmptyEdges(laps []*lapBuilder) []*lapBuilder {
	if len(laps) == 0 {
		return laps
	}

	if last := laps[len(laps)-1]; !hasMeaningfulData(last) {
		laps = laps[:len(laps)-1]
	}
	if len(laps) == 0 {
		return laps
	}

	if first := laps[0]; !hasMeaningfulData(first) {
		laps = laps[1:]
		for _, l := range laps {
			l.numberOfLaps--
		}
	}

	return laps
}

func sectorSumCrossCheck(laps []*lapBuilder) int {
	integrityErrors := 0
	for _, l := range laps {
		sum, missing := sectorSum(l)

		if l.lapTime != nil && !missing && sum > *l.lapTime {
			l.lapTime = nil
			integrityErrors++
			continue
		}
		if l.lapTimeExplicitNone && !missing {
			s := sum
			l.lapTime = &s
		}
	}
	return integrityErrors
}

func sectorSum(l *lapBuilder) (sum time.Duration, missing bool) {
	for _, s := range []*time.Duration{l.sector1, l.sector2, l.sector3} {
		if s == nil {
			missing = true
			continue
		}
		sum += *s
	}
	return sum, missing
}

func reconstructMissingValues(laps []*lapBuilder) {
	for i := 1; i < len(laps); i++ {
		l, prev := laps[i], laps[i-1]

		if l.lapTime == nil && !l.lapTimeExplicitNone {
			sum, missing := sectorSum(l)
			if !missing && prev.lapTime != nil && sum == *prev.lapTime {
				s := sum
				l.lapTime = &s
			}
		}

		if l.lapTime == nil {
			continue
		}
		missingCount, missingSlot := 0, -1
		sectors := []**time.Duration{&l.sector1, &l.sector2, &l.sector3}
		for idx, s := range sectors {
			if *s == nil {
				missingCount++
				missingSlot = idx
			}
		}
		if missingCount != 1 {
			continue
		}
		var otherSum time.Duration
		for idx, s := range sectors {
			if idx != missingSlot && *s != nil {
				otherSum += **s
			}
		}
		backSolved := *l.lapTime - otherSum

		var prevVal *time.Duration
		switch missingSlot {
		case 0:
			prevVal = prev.sector1
		case 1:
			prevVal = prev.sector2
		case 2:
			prevVal = prev.sector3
		}
		if prevVal != nil && backSolved == *prevVal {
			*sectors[missingSlot] = &backSolved
		}
	}
}

func syncLapEndTimes(laps []*lapBuilder) int {
	integrityErrors := 0
	for i := 1; i < len(laps); i++ {
		l, prev := laps[i], laps[i-1]
		if !l.timeSet {
			continue
		}

		minTime := l.time
		var sectorSum time.Duration
		candidates := []struct{ t, s *time.Duration }{
			{l.sector3, l.sector3Session},
			{l.sector2, l.sector2Session},
			{l.sector1, l.sector1Session},
		}
		for _, c := range candidates {
			if c.s != nil {
				if candidate := *c.s + sectorSum; candidate < minTime {
					minTime = candidate
				}
			}
			if c.t == nil {
				break
			}
			sectorSum += *c.t
		}

		if prev.timeSet && minTime < prev.time {
			integrityErrors++
			continue
		}
		l.time = minTime
	}
	return integrityErrors
}

func trimUnusableTrailingLap(laps []*lapBuilder) []*lapBuilder {
	if len(laps) == 0 {
		return laps
	}
	last := laps[len(laps)-1]
	if last.timeSet {
		return laps
	}
	if last.pitInTime != nil {
		last.time = *last.pitInTime
		last.timeSet = true
		return laps
	}
	return laps[:len(laps)-1]
}

func backwardSync(laps []*lapBuilder) int {
	integrityErrors := 0
	for i := len(laps) - 1; i >= 1; i-- {
		l, prev := laps[i], laps[i-1]
		if l.lapTime == nil || !l.timeSet || !prev.timeSet {
			continue
		}
		candidate := l.time - *l.lapTime
		if candidate >= prev.time {
			continue
		}
		if i >= 2 && laps[i-2].timeSet && candidate < laps[i-2].time {
			integrityErrors++
			continue
		}
		prev.time = candidate
	}
	return integrityErrors
}

func forwardSync(laps []*lapBuilder) {
	for i := 0; i < len(laps)-1; i++ {
		cur, next := laps[i], laps[i+1]
		if !cur.timeSet || next.lapTime == nil || !next.timeSet {
			continue
		}
		candidate := cur.time + *next.lapTime
		if candidate >= next.time {
			continue
		}
		next.time = candidate

		var runningSum time.Duration
		sectors := []struct {
			dur *time.Duration
			ses **time.Duration
		}{
			{next.sector1, &next.sector1Session},
			{next.sector2, &next.sector2Session},
			{next.sector3, &next.sector3Session},
		}
		for _, s := range sectors {
			if s.dur == nil {
				break
			}
			runningSum += *s.dur
			candidateSession := cur.time + runningSum
			if *s.ses == nil || candidateSession < **s.ses {
				cs := candidateSession
				*s.ses = &cs
			}
		}
	}
}

func reconstructDeletedPersonalBests(laps []*lapBuilder, sightings []pbEntry) {
	if len(sightings) == 0 || len(laps) == 0 {
		return
	}

	sorted := make([]pbEntry, len(sightings))
	copy(sorted, sightings)
	for i := 1; i < len(sorted); i++ {
		for j := i; j > 0 && sorted[j].elapsed > sorted[j-1].elapsed; j-- {
			sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
		}
	}

	var corrected []time.Duration
	for _, s := range sorted {
		isMin := true
		for _, c := range corrected {
			if s.value >= c {
				isMin = false
				break
			}
		}
		if !isMin {
			continue
		}
		corrected = append(corrected, s.value)

		for _, l := range laps {
			if l.lapTime != nil && *l.lapTime == s.value {
				l.isPersonalBest = true
				break
			}
		}
	}
}

func correctPitStopOffset(laps []*lapBuilder) {
	if len(laps) == 0 {
		return
	}
	offset := laps[0].numberOfPitStops
	if offset == 0 {
		return
	}
	for _, l := range laps {
		l.numberOfPitStops -= offset
	}
}

func correctFirstLapPitIn(laps []*lapBuilder) {
	if len(laps) == 0 {
		return
	}
	first := laps[0]
	if first.pitInTime != nil && first.pitOutTime != nil && *first.pitInTime < *first.pitOutTime {
		first.pitInTime = nil
	}
}
