package personal

import (
	"time"

	sf "../sparkfun/"
	wu "../weather/"
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

func (Personal) GetDaytime() (daytime Daytime) {
	now := time.Now()
	h := now.Hour()
	daytime = DAY
	if h < 12 && h >= 5 {
		daytime = MORNING
	} else if h >= 19 && h < 23 {
		daytime = EVENING
	} else if h >= 23 || h < 5 {
		daytime = NIGHT
	}
	return daytime
}

func (Personal) GetDay() (day Day) {
	now := time.Now()
	d := now.Day()
	day = WORKDAY
	if d >= 5 {
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

func (p Personal) GetPlaceIsOk(place sf.Place) bool {
	day := p.GetDay()
	daytime := p.GetDaytime()
	if place == sf.WORK && day != WORKDAY {
		return false
	}
	if place != sf.HOME && daytime == NIGHT {
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
