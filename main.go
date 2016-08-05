package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	p "./modules/personal/"
	pb "./modules/pushbullet/"
	sf "./modules/sparkfun/"
	tv "./modules/teamviewer/"
	wu "./modules/weather/"
	"github.com/jinzhu/gorm"
	"github.com/qor/transition"
	"github.com/spf13/viper"
)

var pushbullet pb.Pushbullet

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
	// w := weather.GetWeather()

	pushbullet = pb.Connect(viper.GetStringMapString("pushbullet"))
	// pushbullet.SendPush("Hi from Shodan", "Hello, insect")
	// log.Println(pushbullet.GetPushes())

	sparkfun := sf.Connect(viper.GetStringMap("sparkfun"))
	// log.Println(fmt.Sprintf("Temp in room: %v°", sparkfun.GetRoomTemp().Temp))

	teamviewer := tv.Connect(viper.GetStringMapString("teamviewer"))
	log.Println(teamviewer["access_token"])

	wchan := make(chan wu.Weather)
	go func(c chan wu.Weather) {
		for {
			c <- weather.GetWeather()
			time.Sleep(8 * time.Minute)
		}
	}(wchan)

	pchan := make(chan sf.Place)
	go func(c chan sf.Place) {
		for {
			place := sparkfun.GetWhereIAm().Place
			c <- place
			time.Sleep(1 * time.Minute)
		}
	}(pchan)

	weatherState := BinaryState{
		"Ура погода вновь отличная!", "Уруру. Shodan",
		"Погода ухудшилась.", "Мне очень жаль. Shodan",
		transition.Transition{},
	}
	wm := InitBinaryMachine(&weatherState)
	placeState := BinaryState{
		"Фух, я волновалась.", "Уруру. Shodan",
		"Эй, с тобой все в порядке?", "Твоя Shodan",
		transition.Transition{},
	}
	pm := InitBinaryMachine(&placeState)
	for {
		select {
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
			ps := personal.GetPlaceIsOk(p)
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

func InitBinaryMachine(state *BinaryState) *transition.StateMachine {
	wm := transition.New(state)
	wm.Initial("good")
	wm.State("bad").Enter(func(state interface{}, tx *gorm.DB) error {
		s := state.(*BinaryState)
		pushbullet.SendPush(s.BadTitle, s.BadBody)
		return nil
	}).Exit(func(state interface{}, tx *gorm.DB) error {
		s := state.(*BinaryState)
		pushbullet.SendPush(s.GoodTitle, s.GoodBody)
		return nil
	})
	wm.Event("to_good").To("good").From("bad")
	wm.Event("to_bad").To("bad").From("good")
	return wm
}

type BinaryState struct {
	GoodTitle string
	GoodBody  string
	BadTitle  string
	BadBody   string

	transition.Transition
}

type PlaceState struct {
	transition.Transition
}
