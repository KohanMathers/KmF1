package livetiming

type Season struct {
	Year     int       `json:"Year"`
	Meetings []Meeting `json:"Meetings"`
}

type Meeting struct {
	Key          int       `json:"Key"`
	Code         string    `json:"Code"`
	Number       int       `json:"Number"`
	Location     string    `json:"Location"`
	OfficialName string    `json:"OfficialName"`
	Name         string    `json:"Name"`
	Country      Country   `json:"Country"`
	Circuit      Circuit   `json:"Circuit"`
	Sessions     []Session `json:"Sessions"`
}

type Country struct {
	Key  int    `json:"Key"`
	Code string `json:"Code"`
	Name string `json:"Name"`
}

type Circuit struct {
	Key       int    `json:"Key"`
	ShortName string `json:"ShortName"`
}

type Session struct {
	Key       int    `json:"Key"`
	Type      string `json:"Type"`
	Number    int    `json:"Number,omitempty"`
	Name      string `json:"Name"`
	StartDate string `json:"StartDate"`
	EndDate   string `json:"EndDate"`
	GmtOffset string `json:"GmtOffset"`
	Path      string `json:"Path"`
}
