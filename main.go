package main

import (
	"flag"
	"fmt"
	"log"

	// at "./modules/attendance/"
	p "./modules/personal/"
	pb "./modules/pushbullet/"
	sf "./modules/sparkfun/"
	tv "./modules/teamviewer/"
	tg "./modules/telegram/"
	wu "./modules/weather/"
	"github.com/spf13/viper"
)

var pushbullet pb.Pushbullet
var telegram tg.Telegram
var teamviewer tv.TeamViewer

func main() {
	log.Println("======")
	flag.Parse()
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	weather := wu.Connect(viper.GetStringMapString("weather"))
	personal := p.Connect(viper.Get("personal"))
	pushbullet = pb.Connect(viper.GetStringMapString("pushbullet"))
	sparkfun := sf.Connect(viper.GetStringMap("sparkfun"))
	// attendance := at.Connect(viper.GetStringMapString("attendance")).GetAttendance()
	telegram = tg.Connect(viper.GetStringMapString("telegram"))

	teamviewer = tv.Connect(viper.GetStringMapString("teamviewer"))
	// log.Println(teamviewer.GetPCStatus())
	shodan := NewShodan()

	shodan.Serve()
}
