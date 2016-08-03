package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	att "./modules/attendance/"
	pb "./modules/pushbullet/"
	ts "./modules/trackstudio/"
	wu "./modules/weather/"
	"github.com/spf13/viper"
)

func main() {
	flag.Parse()
	viper.SetConfigType("yaml")
	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(".")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	attendance := att.Connect(viper.GetStringMapString("attendance"))
	info := attendance.GetAttendance()
	log.Println(info.AvgWorkingTime)
	log.Println(info.Days[1].WorkingTime)
	log.Println(info.GetHomeTime())
	trackstudio := ts.Connect(viper.GetStringMapString("trackstudio"))
	log.Println(trackstudio.GetReportedYesterday())

	weather := wu.Connect(viper.GetStringMapString("weather"))
	w := weather.GetWeather()
	log.Println(fmt.Sprintf("%s - %v°", w.Weather, w.TempC))

	log.Println(GetDaytime(), GetDay(), GetSeason())

	pushbullet := pb.Connect(viper.GetStringMapString("pushbullet"))
	// pushbullet.SendPush("Hi from Shodan", "Hello, insect")
	log.Println(pushbullet.GetPushes())
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

func GetDaytime() (daytime Daytime) {
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

func GetDay() (day Day) {
	now := time.Now()
	d := now.Day()
	day = WORKDAY
	if d >= 5 {
		day = WEEKEND
	}
	return day
}

func GetSeason() (season Season) {
	now := time.Now()
	m := now.Month()
	if m == 11 && m <= 1 {
		season = WINTER
	} else if m >= 2 && m <= 4 {
		season = SPRING
	} else if m >= 5 && m <= 7 {
		season = SUMMER
	} else {
		season = AUTUMN
	}
	return season
}
