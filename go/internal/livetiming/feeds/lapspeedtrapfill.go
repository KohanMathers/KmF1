package feeds

import "strings"

func ForwardFillSpeedTraps(laps map[string][]LapData) {
	for _, ls := range laps {
		forwardFillSpeedTrapsForDriver(ls)
	}
}

func forwardFillSpeedTrapsForDriver(ls []LapData) {
	i1 := ffillIntSeq(ls, func(l *LapData) *int { return l.SpeedI1 })
	i2 := ffillIntSeq(ls, func(l *LapData) *int { return l.SpeedI2 })
	st := ffillIntSeq(ls, func(l *LapData) *int { return l.SpeedST })
	fl := ffillIntSeq(ls, func(l *LapData) *int { return l.SpeedFL })

	for i := range ls {
		applMask := !ls[i].FastF1Generated && !strings.Contains(ls[i].TrackStatus, "5")
		if applMask {
			ls[i].SpeedI1 = i1[i]
			ls[i].SpeedI2 = i2[i]
			ls[i].SpeedST = st[i]
		}
		if applMask && ls[i].PitInTime == nil {
			ls[i].SpeedFL = fl[i]
		}
	}
}

func ffillIntSeq(ls []LapData, get func(*LapData) *int) []*int {
	out := make([]*int, len(ls))
	var last *int
	for i := range ls {
		if v := get(&ls[i]); v != nil {
			last = v
		}
		out[i] = last
	}
	return out
}
