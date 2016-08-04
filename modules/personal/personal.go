package personal

import "time"

type Personal struct{}

func Connect() Personal {
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
