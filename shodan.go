package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/qor/transition"
)

type ShodanString []string

func (s ShodanString) String() string {
	return s[rand.Intn(len(s))]
}

type Shodan struct {
	Strings map[string]ShodanString
}

func NewShodan() *Shodan {
	s := Shodan{}
	s.Strings = map[string]ShodanString{
		"hello": ShodanString{
			"Привет, я включилась.",
			"О, уже утро?",
			"Я уже работаю, а ты?",
		},
	}

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
	wm := NewBinaryMachine(&weatherState, &s)
	// placeState := BinaryState{
	// 	"Фух, я волновалась.", "Уруру. Shodan",
	// 	"Эй, с тобой все в порядке?", "Твоя Shodan",
	// 	transition.Transition{},
	// }
	// pm := NewBinaryMachine(&placeState)
	ps := PlaceState{}
	m := NewPlaceMachine(&ps, &s)
	return &s
}

func (s *Shodan) Say(name string) string {
	return fmt.Sprintf("%s", s.Strings[name])
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

	telegram.Send(s.Say("hello"))
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
