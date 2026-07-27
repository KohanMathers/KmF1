package kmf1

// Session type identifiers, used with Event.Session and matched against Session.Info().Type. They correspond to the session names as they appear in the Jolpica and live-timing schedule data.
const (
	SessionPractice1        = "Practice 1"
	SessionPractice2        = "Practice 2"
	SessionPractice3        = "Practice 3"
	SessionQualifying       = "Qualifying"
	SessionSprintQualifying = "Sprint Qualifying"
	SessionSprint           = "Sprint"
	SessionRace             = "Race"
)

func isRaceLike(sessionType string) bool {
	return sessionType == SessionRace || sessionType == SessionSprint
}

func isQualiLike(sessionType string) bool {
	return sessionType == SessionQualifying || sessionType == SessionSprintQualifying
}

func ergastListAccurate(sessionType string) bool {
	return isRaceLike(sessionType) || sessionType == SessionQualifying
}
