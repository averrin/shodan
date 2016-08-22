package personal

import (
	"time"

	ds "github.com/averrin/shodan/modules/datastream"
	wu "github.com/averrin/shodan/modules/weather"
)

type Personal struct{}

func Connect(creds interface{}) Personal {
	return Personal{}
}

type Daytime int

const (
	DAY = iota
	NIGHT
	MORNING
	EVENING
)

type Day int

const (
	WORKDAY = iota
	WEEKEND
)

type Season int

const (
	WINTER = iota
	SPRING
	SUMMER
	AUTUMN
)

func (Personal) GetDaytime() (daytime string) {
	now := time.Now()
	h := now.Hour()
	daytime = "day"
	if h < 12 && h >= 5 {
		daytime = "morning"
	} else if h >= 19 && h < 23 {
		daytime = "evening"
	} else if h >= 23 || h < 5 {
		daytime = "night"
	}
	return daytime
}

func (Personal) GetDay() (day Day) {
	now := time.Now()
	d := now.Weekday()
	day = WORKDAY
	if d >= 6 {
		day = WEEKEND
	}
	return day
}

func (Personal) GetSeason() (season Season) {
	now := time.Now()
	m := now.Month()
	if m == 12 && m <= 2 {
		season = WINTER
	} else if m >= 3 && m <= 5 {
		season = SPRING
	} else if m >= 6 && m <= 8 {
		season = SUMMER
	} else {
		season = AUTUMN
	}
	return season
}

func (Personal) GetActivity(datastream *ds.DataStream) bool {
	phone := ds.Value{}
	datastream.Get("display", &phone)
	pc := ds.Value{}
	datastream.Get("pc", &pc)
	if phone.Value.(string) == "on" || pc.Value.(string) == "unidle" {
		return true
	}
	return false
}

func (p Personal) GetPlaceIsOk(point ds.Point) bool {
	place := point.Name
	day := p.GetDay()
	daytime := p.GetDaytime()
	if place == "work" && day != WORKDAY {
		return false
	}
	if place != "home" && daytime == "night" {
		return false
	}
	return true
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (p Personal) GetWeatherIsOk(weather wu.Weather) bool {
	bad := []string{
		"rain",
		"chanceflurries",
		"chancerain",
		"chancesleet",
		"chancetstorms",
		"flurries",
		"fog",
		"hazy",
		"sleet",
		"snow",
		"tstorms",
	}
	if contains(bad, weather.Icon) {
		return false
	}
	season := p.GetSeason()
	if weather.TempC > 28 || weather.TempC < -10 {
		return false
	}
	if season == SUMMER && weather.TempC < 15 {
		return false
	}
	if (season == SUMMER || season == AUTUMN) && weather.TempC < 5 {
		return false
	}
	return true
}

type Place int

const (
	WORK = iota
	HOME
	NOWHERE
	VILLAGE
	PAVEL
)

var Places map[string]Place
