package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	redis "gopkg.in/redis.v4"

	at "github.com/averrin/shodan/modules/attendance"
	p "github.com/averrin/shodan/modules/personal"
	pb "github.com/averrin/shodan/modules/pushbullet"
	sf "github.com/averrin/shodan/modules/sparkfun"
	tg "github.com/averrin/shodan/modules/telegram"
	wu "github.com/averrin/shodan/modules/weather"
	"github.com/jinzhu/gorm"
	"github.com/qor/transition"
	"github.com/spf13/viper"
)

var pushbullet pb.Pushbullet
var telegram tg.Telegram

// var teamviewer tv.TeamViewer
var weather wu.WUnderground
var sparkfun sf.SparkFun
var personal p.Personal
var attendance *at.Info

type ShodanString []string

func (s ShodanString) String() string {
	return s[rand.Intn(len(s))]
}

type Shodan struct {
	Strings   map[string]ShodanString
	Machines  map[string]*transition.StateMachine
	States    map[string]transition.Stater
	Flags     map[string]bool
	LastPlace string
	DB        *gorm.DB
	Client    *redis.Client
}

func NewRedis() (client *redis.Client) {
	client = redis.NewClient(&redis.Options{
		Addr:     "averr.in:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)
	return client
}

func NewShodan() *Shodan {
	s := Shodan{}
	db, err := gorm.Open("sqlite3", "shodan.db")
	if err != nil {
		panic("failed to connect database")
	}
	s.DB = db
	s.Client = NewRedis()

	rand.Seed(time.Now().UnixNano())
	s.Machines = map[string]*transition.StateMachine{}
	s.States = map[string]transition.Stater{}
	s.Flags = map[string]bool{
		"late at work": false,
	}
	s.Strings = map[string]ShodanString{
		"hello": ShodanString{
			"Привет, я включилась.",
			"О, уже утро?",
			"Я уже работаю, а ты?",
			"Прямо хочется что-нить делать.",
			"Бип-бип. Бип=)",
		},
		"good weather": ShodanString{
			"Ура погода вновь отличная! Уруру.",
			"Можно идти гулять.",
			"На улице стало приличнее",
		},
		"bad weather": ShodanString{
			"Погода ухудшилась. Мне очень жаль.",
			"Что-то хрень какая-то на улице",
			"Посмотрела погоду, не понравилось",
		},
		"at home": ShodanString{
			"Ты наконец дома, ура!",
		},
		"at home, no pc": ShodanString{
			"Ты 15 минут дома, а комп не включен. Все в порядке?",
		},
		"good way": ShodanString{
			"Хорошей дороги.",
			"Веди аккуратно.",
		},
		"go home": ShodanString{
			"Ты это чего еще на работе?",
			"Эй! Марш домой!",
		},
		"wrong place": ShodanString{
			"Эй, с тобой все в порядке?",
			"Что-то ты где-то не там, где должен быть, не?",
		},
	}

	weather = wu.Connect(viper.GetStringMapString("weather"))
	personal = p.Connect(viper.Get("personal"))
	pushbullet = pb.Connect(viper.GetStringMapString("pushbullet"))
	sparkfun = sf.Connect(viper.GetStringMap("sparkfun"))
	attendance = at.Connect(viper.GetStringMapString("attendance")).GetAttendance()
	telegram = tg.Connect(viper.GetStringMapString("telegram"))

	// teamviewer = tv.Connect(viper.GetStringMapString("teamviewer"))

	weatherState := BinaryState{
		"good weather", "bad weather",
		transition.Transition{},
	}
	s.States["weather"] = &weatherState
	s.Machines["weather"] = NewBinaryMachine(&weatherState, &s)

	ps := PlaceState{}
	s.States["place"] = &ps
	s.Machines["place"] = NewPlaceMachine(&ps, &s)

	dts := DayTimeState{}
	s.States["daytime"] = &dts
	s.Machines["daytime"] = NewDayTimeMachine(&dts, &s)
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
	tchan := make(chan time.Duration)
	go func(c chan time.Duration) {
		for {
			_, _, sinceDI, _, _ := attendance.GetHomeTime()
			c <- sinceDI
			time.Sleep(3 * time.Minute)
		}
	}(tchan)

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
			if place.Name != "" {
				s.LastPlace = place.Name
				c <- place
			}
			time.Sleep(1 * time.Minute)
		}
	}(pchan)

	s.Say("hello")
	for {
		select {
		case t := <-tchan:
			s.Machines["daytime"].Trigger(personal.GetDaytime(), s.States["daytime"], s.DB)
			if t.Minutes() < 1 && s.LastPlace == "work" && s.Flags["late at work"] != true {
				go func() {
					s.Flags["late at work"] = true
					time.Sleep(10 * time.Minute)
					if s.LastPlace == "work" {
						s.Say("go home")
					}
				}()
			}
		case w := <-wchan:
			ws := personal.GetWeatherIsOk(w)
			var event string
			if ws {
				event = "to_good"
			} else {
				event = "to_bad"
			}
			err := s.Machines["weather"].Trigger(event, s.States["weather"], s.DB)
			log.Println(err)
			if err == nil {
				s.Say(fmt.Sprintf("%s - %v°", w.Weather, w.TempC))
			}
		case p := <-pchan:
			err := s.Machines["place"].Trigger(p.Name, s.States["place"], s.DB)
			if err == nil {
				s.Flags["wrong place"] = false
			}
			ps := personal.GetPlaceIsOk(p.Place)
			if !ps && s.Flags["wrong place"] != true {
				go func() {
					s.Flags["wrong place"] = true
					time.Sleep(10 * time.Minute)
					if s.LastPlace == p.Name {
						s.Say("wrong place")
					}
				}()
			}
		default:
		}
	}
}
