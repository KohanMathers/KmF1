package styling

import (
	_ "embed"
	"encoding/json"
	"sort"
	"strconv"
	"strings"
)

//go:embed constants.json
var constantsJSON []byte

type TeamColors struct {
	Official string
	FastF1   string
}

type TeamStyle struct {
	Key       string
	ShortName string
	Colors    TeamColors
}

type SeasonConstants struct {
	compoundColors map[string]string
	teams          map[string]TeamStyle
}

type wireSeasonConstants struct {
	CompoundColors map[string]string            `json:"compound_colors"`
	Teams          map[string]wireTeamConstants `json:"teams"`
}

type wireTeamConstants struct {
	ShortName string         `json:"short_name"`
	Colors    wireTeamColors `json:"colors"`
}

type wireTeamColors struct {
	Official string `json:"official"`
	FastF1   string `json:"fastf1"`
}

var seasons map[string]SeasonConstants

func init() {
	var wire map[string]wireSeasonConstants
	if err := json.Unmarshal(constantsJSON, &wire); err != nil {
		panic("styling: embedded constants.json is invalid: " + err.Error())
	}

	seasons = make(map[string]SeasonConstants, len(wire))
	for year, w := range wire {
		teams := make(map[string]TeamStyle, len(w.Teams))
		for key, t := range w.Teams {
			teams[key] = TeamStyle{
				Key:       key,
				ShortName: t.ShortName,
				Colors:    TeamColors{Official: t.Colors.Official, FastF1: t.Colors.FastF1},
			}
		}
		seasons[year] = SeasonConstants{compoundColors: w.CompoundColors, teams: teams}
	}
}

func ForYear(year int) (SeasonConstants, bool) {
	s, ok := seasons[strconv.Itoa(year)]
	return s, ok
}

func (s SeasonConstants) CompoundColor(compound string) (string, bool) {
	for name, color := range s.compoundColors {
		if strings.EqualFold(name, compound) {
			return color, true
		}
	}
	return "", false
}

func (s SeasonConstants) Compounds() []string {
	out := make([]string, 0, len(s.compoundColors))
	for name := range s.compoundColors {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func (s SeasonConstants) Team(identifier string) (TeamStyle, bool) {
	needle := strings.ToLower(strings.TrimSpace(identifier))
	if needle == "" {
		return TeamStyle{}, false
	}

	if t, ok := s.teams[needle]; ok {
		return t, true
	}
	for _, t := range s.teams {
		if strings.EqualFold(t.ShortName, identifier) {
			return t, true
		}
	}

	for key, t := range s.teams {
		short := strings.ToLower(t.ShortName)
		if strings.Contains(key, needle) || strings.Contains(needle, key) {
			return t, true
		}
		if strings.Contains(short, needle) || strings.Contains(needle, short) {
			return t, true
		}
	}

	return TeamStyle{}, false
}

func (s SeasonConstants) Teams() []TeamStyle {
	out := make([]TeamStyle, 0, len(s.teams))
	for _, t := range s.teams {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ShortName < out[j].ShortName })
	return out
}
