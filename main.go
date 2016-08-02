package main

import (
	"flag"
	"fmt"
	"log"

	att "./modules/attendance/"
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
}
