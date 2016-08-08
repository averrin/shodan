package main

import (
	"fmt"
	"math/rand"
	"time"

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
var weather wu.WUnderground
var sparkfun sf.SparkFun
var personal p.Personal

type ShodanString []string

func (s ShodanString) String() string {
	return s[rand.Intn(len(s))]
}

type Shodan struct {
	Strings  map[string]ShodanString
	Machines map[string]*transition.StateMachine
	States   map[string]transition.Stater
}

func NewShodan() *Shodan {
	rand.Seed(time.Now().UnixNano())
	s := Shodan{}
	s.Machines = map[string]*transition.StateMachine{}
	s.States = map[string]transition.Stater{}
	s.Strings = map[string]ShodanString{
		"hello": ShodanString{
			"Привет, я включилась.",
			"О, уже утро?",
			"Я уже работаю, а ты?",
		},
		"good weather": ShodanString{
			"Ура погода вновь отличная! Уруру.",
		},
		"bad weather": ShodanString{
			"Погода ухудшилась. Мне очень жаль.",
		},
		"at home": ShodanString{
			"Ты наконец дома, ура!",
		},
		"at home, no pc": ShodanString{
			"Ты 15 минут дома, а комп не включен. Все в порядке?",
		},
		"good way": ShodanString{
			"Хорошей дороги.",
		},
	}

	weather = wu.Connect(viper.GetStringMapString("weather"))
	personal = p.Connect(viper.Get("personal"))
	pushbullet = pb.Connect(viper.GetStringMapString("pushbullet"))
	sparkfun = sf.Connect(viper.GetStringMap("sparkfun"))
	// attendance := at.Connect(viper.GetStringMapString("attendance")).GetAttendance()
	telegram = tg.Connect(viper.GetStringMapString("telegram"))

	teamviewer = tv.Connect(viper.GetStringMapString("teamviewer"))

	// tchan := make(chan time.Duration)
	// go func(c chan time.Duration) {
	// 	for {
	// 		_, _, sinceDI, _, _ := attendance.GetHomeTime()
	// 		c <- sinceDI
	// 		time.Sleep(1 * time.Minute)
	// 	}
	// }(tchan)

	weatherState := BinaryState{
		"good weather", "bad weather",
		transition.Transition{},
	}
	s.States["weather"] = &weatherState
	wm := NewBinaryMachine(&weatherState, &s)
	s.Machines["weather"] = wm
	// placeState := BinaryState{
	// 	"Фух, я волновалась.", "Уруру. Shodan",
	// 	"Эй, с тобой все в порядке?", "Твоя Shodan",
	// 	transition.Transition{},
	// }
	// pm := NewBinaryMachine(&placeState)
	ps := PlaceState{}
	s.States["place"] = &ps
	m := NewPlaceMachine(&ps, &s)
	s.Machines["place"] = m
	return &s
}

func (s *Shodan) GetString(name string) string {
	return fmt.Sprintf("%s", s.Strings[name])
}

func (s *Shodan) Say(name string) {
	if s.Strings[name] != nil {
		telegram.Send(s.GetString(name))
	} else {
		telegram.Send(name)
	}
}

func (s *Shodan) Serve() {
	wchan := make(chan wu.Weather)
	go func(c chan wu.Weather) {
		for {
			c <- weather.GetWeather()
			time.Sleep(8 * time.Minute)
		}
	}(wchan)

	pchan := make(chan sf.Point)
	go func(c chan sf.Point) {
		for {
			place := sparkfun.GetWhereIAm()
			c <- place
			time.Sleep(1 * time.Minute)
		}
	}(pchan)

	s.Say("hello")
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
			ws := personal.GetWeatherIsOk(w)
			var event string
			if ws {
				event = "to_good"
			} else {
				event = "to_bad"
			}
			err := s.Machines["weather"].Trigger(event, s.States["weather"], nil)
			if err == nil {
				s.Say(fmt.Sprintf("%s - %v°", w.Weather, w.TempC))
			}
		case p := <-pchan:
			s.Machines["place"].Trigger(p.Name, s.States["place"], nil)
			// ps := personal.GetPlaceIsOk(p.Place)
			// log.Println("Im on my place: ", ps)
			// var event string
			// if ps {
			// 	event = "to_good"
			// } else {
			// 	event = "to_bad"
			// }
			// pm.Trigger(event, &placeState, nil)
		default:
		}
	}
}
