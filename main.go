package main

import (
	"flag"
	"fmt"
	"log"

	p "./modules/personal/"
	sf "./modules/sparkfun/"
	wu "./modules/weather/"
	"github.com/spf13/viper"
)

func main() {
	flag.Parse()
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	// attendance := att.Connect(viper.GetStringMapString("attendance"))
	// info := attendance.GetAttendance()
	// log.Println(info.AvgWorkingTime)
	// log.Println(info.Days[1].WorkingTime)
	// log.Println(info.GetHomeTime())
	// trackstudio := ts.Connect(viper.GetStringMapString("trackstudio"))
	// log.Println(trackstudio.GetReportedYesterday())

	weather := wu.Connect(viper.GetStringMapString("weather"))
	w := weather.GetWeather()
	log.Println(fmt.Sprintf("%s - %v°", w.Weather, w.TempC))

	personal := p.Connect(viper.Get("personal"))
	log.Println(personal.GetDaytime() == p.DAY, personal.GetDay() == p.WORKDAY, personal.GetSeason() == p.SUMMER)

	// pushbullet := pb.Connect(viper.GetStringMapString("pushbullet"))
	// pushbullet.SendPush("Hi from Shodan", "Hello, insect")
	// log.Println(pushbullet.GetPushes())

	sparkfun := sf.Connect(viper.GetStringMap("sparkfun"))
	place := sparkfun.GetWhereIAm().Place
	log.Println(place == sf.WORK)
	log.Println(sparkfun.GetRoomTemp().Temp)
	log.Println(personal.GetPlaceIsOk(place))
	log.Println(personal.GetWeatherIsOk(w))

}
