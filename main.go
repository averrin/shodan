package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	// at "./modules/attendance/"
	p "./modules/personal/"
	pb "./modules/pushbullet/"
	sf "./modules/sparkfun/"
	tv "./modules/teamviewer/"
	tg "./modules/telegram/"
	wu "./modules/weather/"
	"github.com/qor/transition"
	"github.com/spf13/viper"
)

var pushbullet pb.Pushbullet
var telegram tg.Telegram
var teamviewer tv.TeamViewer

func main() {
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
	log.Println(teamviewer.GetPCStatus())

	wchan := make(chan wu.Weather)
	go func(c chan wu.Weather) {
		for {
			c <- weather.GetWeather()
			time.Sleep(8 * time.Minute)
		}
	}(wchan)

	pchan := make(chan sf.Point)
	// pchan <- sparkfun.GetWhereIAm()
	go func(c chan sf.Point) {
		for {
			place := sparkfun.GetWhereIAm()
			c <- place
			time.Sleep(1 * time.Minute)
		}
	}(pchan)

	// tchan := make(chan time.Duration)
	// go func(c chan time.Duration) {
	// 	for {
	// 		_, _, sinceDI, _, _ := attendance.GetHomeTime()
	// 		c <- sinceDI
	// 		time.Sleep(1 * time.Minute)
	// 	}
	// }(tchan)

	weatherState := BinaryState{
		"Ура погода вновь отличная!", "Уруру. Shodan",
		"Погода ухудшилась.", "Мне очень жаль. Shodan",
		transition.Transition{},
	}
	wm := NewBinaryMachine(&weatherState)
	placeState := BinaryState{
		"Фух, я волновалась.", "Уруру. Shodan",
		"Эй, с тобой все в порядке?", "Твоя Shodan",
		transition.Transition{},
	}
	pm := NewBinaryMachine(&placeState)
	ps := PlaceState{}
	m := NewPlaceMachine(&ps)
	for {
		select {
		// case t := <-tchan:
		// p := <-pchan
		// if t.Minutes() < 1 && p.Place == sf.WORK {
		// pushbullet.SendPush("Ты это чего еще на работе?", "Марш домой!")
		// TODO: add machine for only one notifucation
		// TODO: start notification after 10 minutes after deadline
		// }
		case w := <-wchan:
			log.Println("Street weather:", fmt.Sprintf("%s - %v°", w.Weather, w.TempC))
			ws := personal.GetWeatherIsOk(w)
			log.Println("And its good:", ws)
			var event string
			if ws {
				event = "to_good"
			} else {
				event = "to_bad"
			}
			wm.Trigger(event, &weatherState, nil)
		case p := <-pchan:
			m.Trigger(p.Name, &ps, nil)
			ps := personal.GetPlaceIsOk(p.Place)
			log.Println("Im on my place: ", ps)
			var event string
			if ps {
				event = "to_good"
			} else {
				event = "to_bad"
			}
			pm.Trigger(event, &placeState, nil)
		default:
		}
	}
}
