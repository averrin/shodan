package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	at "github.com/averrin/shodan/modules/attendance"
	p "github.com/averrin/shodan/modules/personal"
	"github.com/fatih/color"
	// pb "github.com/averrin/shodan/modules/pushbullet"
	a "github.com/averrin/shodan/modules/auth"
	ds "github.com/averrin/shodan/modules/datastream"
	stor "github.com/averrin/shodan/modules/storage"
	tg "github.com/averrin/shodan/modules/telegram"
	wu "github.com/averrin/shodan/modules/weather"
	"github.com/jinzhu/gorm"
	"github.com/qor/transition"
	"github.com/spf13/viper"
)

// var pushbullet pb.Pushbullet
var telegram tg.Telegram

var weather *wu.WUnderground
var datastream *ds.DataStream
var storage *stor.Storage
var personal p.Personal
var attendance at.Attendance
var auth *a.Auth
var nobot *bool

// var eventghost *eg.EventGhost
// var smarthome sh.SmartHome

type ShodanString []string

func (s ShodanString) String() string {
	return s[rand.Intn(len(s))]
}

type Shodan struct {
	Strings   map[string]ShodanString
	Machines  map[string]*transition.StateMachine
	States    map[string]transition.Stater
	LastTimes map[string]time.Time
	Flags     map[string]bool
	Commands  Commands
	LastPlace string
	DB        *gorm.DB
}

func NewShodan() *Shodan {
	s := Shodan{}
	db, err := gorm.Open("sqlite3", "shodan.db")
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&transition.StateChangeLog{})
	s.DB = nil

	rand.Seed(time.Now().UnixNano())
	s.Machines = map[string]*transition.StateMachine{}
	s.States = map[string]transition.Stater{}
	s.LastTimes = map[string]time.Time{}
	s.Commands = s.getCommands()
	s.Flags = map[string]bool{
		"late at work":       false,
		"debug":              false,
		"pc activity notify": false,
	}
	s.Strings = getStrings()

	weather = wu.Connect(viper.GetStringMapString("weather"))
	personal = p.Connect(viper.Get("personal"))
	// pushbullet = pb.Connect(viper.GetStringMapString("pushbullet"))
	datastream = ds.Connect(viper.GetStringMapString("datastream"))
	auth = a.Connect(datastream)
	storage = stor.Connect(viper.GetStringMapString("storage"))
	attendance = at.Connect(viper.GetStringMapString("attendance"))
	if !*nobot {
		telegram = tg.Connect(viper.GetStringMapString("telegram"))
	}
	// smarthome = sh.Connect(viper.GetStringMapString("smarthome"))
	// eventghost = eg.Connect(viper.GetStringMapString("eventghost"))

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

	as := ActivityState{}
	s.States["activity"] = &as
	s.Machines["activity"] = NewActivityMachine(&as, &s)

	ss := SleepState{}
	s.States["sleep"] = &ss
	s.Machines["sleep"] = NewSleepMachine(&ss, &s)

	s.LastTimes["start"] = time.Now()
	return &s
}

func (s *Shodan) GetString(name string) string {
	return fmt.Sprintf("%s", s.Strings[name])
}

func (s *Shodan) Say(name string) {
	if name == "" {
		return
	}
	if s.Strings[name] != nil {
		m := s.GetString(name)
		telegram.Send(m)
		storage.ReportEvent("say", m)
	} else {
		telegram.Send(name)
		storage.ReportEvent("sayDirect", name)
	}
}

func (s *Shodan) SayOffRecord(name string) {
	if name == "" {
		return
	}
	if s.Strings[name] != nil {
		m := s.GetString(name)
		telegram.Send(m)
	} else {
		telegram.Send(name)
	}
}

func (s *Shodan) Serve() {
	ticker := time.NewTicker(1 * time.Hour)
	notifications := s.getNotifications()
	go func() {
		for _ = range ticker.C {
			// log.Println("Start testing notifications")
			for _, n := range notifications {
				// log.Println(fmt.Sprintf("Test %v", n))
				if n.Test() {
					log.Println(n.Text)
					s.Say(n.Text)
					s.GideonNotify(s.GetString(n.Text))
				}
			}
		}
	}()

	tchan := make(chan time.Duration)
	go func(c chan time.Duration) {
		for {
			if s.LastPlace == "work" {
				ht := attendance.GetAttendance().GetHomeTime()
				datastream.SetValue("attendance", ht)
				c <- ht.SinceIdeal
			}
			time.Sleep(3 * time.Minute)
		}
	}(tchan)

	wchan := make(chan wu.Weather)
	go func(c chan wu.Weather) {
		for {
			w, err := weather.GetWeather()
			if err != nil {
				time.Sleep(10 * time.Minute)
				continue
			}
			c <- w
			time.Sleep(10 * time.Minute)
		}
	}(wchan)

	pchan := make(chan ds.Point)
	go func(c chan ds.Point) {
		for {
			place := datastream.GetWhereIAm()
			if place.Name != "" {
				s.LastPlace = place.Name
				c <- place
			}
			time.Sleep(10 * time.Second)
		}
	}(pchan)

	ichan := make(chan string)
	if !*nobot {
		telegram.SetInbox(ichan)
	}

	s.initAPI()
	datastream.Heartbeat("shodan")
	s.Say("hello")
	storage.ReportEvent("startShodan", "")
	go s.trackGideon()

	for {
		select {
		case m := <-ichan:
			blue := color.New(color.FgBlue).SprintFunc()
			log.Println(blue("M:"), m)
			s.dispatchMessages(m)
		case t := <-tchan:
			dt := personal.GetDaytime()
			s.Machines["daytime"].Trigger(dt, s.States["daytime"], s.DB)
			WORK := "work"
			if t.Minutes() < 1 && s.LastPlace == WORK && s.Flags["late at work"] != true && time.Now().Hour() > 12 {
				go func() {
					s.Flags["late at work"] = true
					time.Sleep(10 * time.Minute)
					ht := attendance.GetAttendance().GetHomeTime()
					if s.LastPlace == WORK {
						if dt != "evening" && ht.SinceIdeal.Minutes() < 1 {
							s.Flags["late at work"] = false
							s.Say("attendance glitch")
							s.Say(fmt.Sprintf("Debug: %v", t))
							storage.ReportEvent("attendanceGlitch", "")
						} else {
							s.Say("go home")
							s.Say(fmt.Sprintf("Debug: %v", t))
							storage.ReportEvent("lateAtWork", "")
						}
					} else {
						s.Flags["late at work"] = false
					}
				}()
			}
		case w := <-wchan:
			if w.Weather == "" {
				s.Say("no weather")
				return
			}
			ws := personal.GetWeatherIsOk(w)
			datastream.SetValue("weatherisok", fmt.Sprintf("%v", ws))
			datastream.SetValue("weather", w)
			datastream.SetValue("weatherLocation", weather.GetLocation())
			var event string
			if ws {
				event = "to_good"
			} else {
				event = "to_bad"
			}
			err := s.Machines["weather"].Trigger(event, s.States["weather"], s.DB)
			if err == nil {
				wea := fmt.Sprintf("%s - %v°", w.Weather, w.TempC)
				s.Say(wea)
			}
		case p := <-pchan:
			err := s.Machines["place"].Trigger(p.Name, s.States["place"], s.DB)
			if err == nil {
				s.Flags["wrong place"] = false
				if s.Flags["debug"] {
					s.Say(fmt.Sprintf("New place: %s", p.Name))
				}
			}
			ps := personal.GetPlaceIsOk(p)
			if !ps && s.Flags["wrong place"] != true {
				go func() {
					s.Flags["wrong place"] = true
					time.Sleep(10 * time.Minute)
					if s.LastPlace == p.Name {
						s.Say("wrong place")
						storage.ReportEvent("wrongPlace", p.Name)
					}
				}()
			}
		default:
		}
	}
}

func (s *Shodan) LightOn(name string) {
	datastream.SendCommand(ds.Command{
		"light", map[string]interface{}{
			"name": name,
			"code": "On",
		}, "gideon", "Shodan",
	})
}

func (s *Shodan) GideonNotify(msg string) {
	result := datastream.SendCommand(ds.Command{
		"notify", map[string]interface{}{
			"message": msg,
		}, "gideon", "Shodan",
	})
	if result.Success {
		log.Println("command success")
	} else {
		log.Println("command fail")
		if result.Result != nil {
			s.Say(result.Result.(string))
		}
	}
}

func (s *Shodan) UpdateGideon() {
	result := datastream.SendCommand(ds.Command{
		"update", nil, "gideon", "Shodan",
	})
	if result.Success {
		s.Say("command success")
	} else {
		s.Say("command fail")
		if result.Result != nil {
			s.Say(result.Result.(string))
		}
	}
}

func (s *Shodan) trackGideon() {
	gideon := datastream.GetHeartbeat("gideon")
	s.Flags["gideon online"] = true
	s.LastTimes["gideon seen"] = time.Time{}
	notified := false
	for {
		select {
		case ping, ok := <-gideon:
			if ok {
				if !ping {
					if s.Flags["gideon online"] == true && (s.LastTimes["gideon seen"].IsZero() || time.Now().Sub(s.LastTimes["gideon seen"]) > time.Duration(5*time.Minute)) {
						s.Say("gideon away")
						notified = true
					}
					s.Flags["gideon online"] = false
				} else {
					if !s.Flags["gideon online"] && notified {
						s.Say("gideon started")
						notified = false
					}
					s.Flags["gideon online"] = true
					s.LastTimes["gideon seen"] = time.Now()
				}
			} else {
				if s.Flags["gideon online"] == true {
					s.Say("gideon away")
				}
				s.Flags["gideon online"] = false
			}
		default:
		}
	}
}
