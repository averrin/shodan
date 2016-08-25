package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	at "github.com/averrin/shodan/modules/attendance"
	p "github.com/averrin/shodan/modules/personal"
	// pb "github.com/averrin/shodan/modules/pushbullet"
	ds "github.com/averrin/shodan/modules/datastream"
	// eg "github.com/averrin/shodan/modules/eventghost"
	// sh "github.com/averrin/shodan/modules/smarthome"
	stor "github.com/averrin/shodan/modules/storage"
	tg "github.com/averrin/shodan/modules/telegram"
	wu "github.com/averrin/shodan/modules/weather"
	"github.com/jinzhu/gorm"
	"github.com/qor/transition"
	"github.com/spf13/viper"
)

// var pushbullet pb.Pushbullet
var telegram tg.Telegram

var weather wu.WUnderground
var datastream *ds.DataStream
var storage *stor.Storage
var personal p.Personal
var attendance *at.Info
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
	storage = stor.Connect(viper.GetStringMapString("storage"))
	attendance = at.Connect(viper.GetStringMapString("attendance")).GetAttendance()
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
	if s.Strings[name] != nil {
		telegram.Send(s.GetString(name))
	} else {
		telegram.Send(name)
	}
}

func (s *Shodan) Serve() {
	ticker := time.NewTicker(1 * time.Hour)
	notifications := s.getNotifications()
	go func() {
		for _ = range ticker.C {
			log.Println("Start testing notifications")
			for _, n := range notifications {
				log.Println(fmt.Sprintf("Test %v", n))
				if n.Test() {
					log.Println(n.Text)
					s.Say(n.Text)
				}
			}
		}
	}()

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
	for {
		select {
		case m := <-ichan:
			log.Println(m)
			s.dispatchMessages(m)
		case t := <-tchan:
			dt := personal.GetDaytime()
			s.Machines["daytime"].Trigger(dt, s.States["daytime"], s.DB)
			if t.Minutes() < 1 && s.LastPlace == "work" && s.Flags["late at work"] != true && time.Now().Hour() > 12 {
				go func() {
					time.Sleep(10 * time.Minute)
					_, _, sinceDI, _, _ := attendance.GetHomeTime()
					if s.LastPlace == "work" {
						if dt != "evening" && sinceDI.Minutes() < 1 {
							s.Say("attendance glitch")
							s.Say(fmt.Sprintf("Debug: %v", t))
							storage.ReportEvent("attendanceGlitch", "")
						} else {
							s.Flags["late at work"] = true
							s.Say("go home")
							s.Say(fmt.Sprintf("Debug: %v", t))
							storage.ReportEvent("lateAtWork", "")
						}
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

func (s *Shodan) initAPI() {
	http.HandleFunc("/battery/", func(w http.ResponseWriter, r *http.Request) {
		level := r.URL.Path[len("/battery/"):]
		datastream.SetValue("battery", level)
		if level == "low" {
			s.Say("low battery")
			storage.ReportEvent("lowBattery", "")
		}
	})
	http.HandleFunc("/dream/", func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Path[len("/dream/"):]
		datastream.SetValue("dream", status)
		err := s.Machines["sleep"].Trigger(status, s.States["sleep"], s.DB)
		if err != nil {
			log.Println(status)
			log.Println(err)
		}
		storage.ReportEvent(status, "")
	})
	http.HandleFunc("/power/", func(w http.ResponseWriter, r *http.Request) {
		status := strings.TrimSpace(r.URL.Path[len("/power/"):])
		storage.ReportEvent("power", status)
		if s.States["sleep"].GetState() != "awake" {
			err := s.Machines["sleep"].Trigger("awake", s.States["sleep"], s.DB)
			if err != nil {
				log.Println(err)
			}
		}
	})
	http.HandleFunc("/place/", func(w http.ResponseWriter, r *http.Request) {
		place := strings.TrimSpace(r.URL.Path[len("/place/"):])
		datastream.SetWhereIAm(place)
		err := s.Machines["place"].Trigger(place, s.States["place"], s.DB)
		if err != nil {
			log.Println(place)
			log.Println(err)
		}
	})
	http.HandleFunc("/cmd/", func(w http.ResponseWriter, r *http.Request) {
		tokens := strings.Split(r.URL.Path[len("/cmd/"):], "/")
		datastream.SendCommand(ds.Command{
			tokens[1], nil, tokens[0], "Shodan",
		})
		storage.ReportEvent("command", r.URL.Path[len("/cmd/"):])
	})
	http.HandleFunc("/psb/", func(w http.ResponseWriter, r *http.Request) {
		message, _ := ioutil.ReadAll(r.Body)
		s.Say(string(message))
		defer r.Body.Close()
	})
	http.HandleFunc("/display/", func(w http.ResponseWriter, r *http.Request) {
		display := strings.TrimSpace(r.URL.Path[len("/display/"):])
		datastream.SetValue("display", display)
		storage.ReportEvent("displayActivity", display)
		var err error
		if personal.GetActivity(datastream) {
			err = s.Machines["activity"].Trigger("active", s.States["activity"], s.DB)
		} else {
			err = s.Machines["activity"].Trigger("idle", s.States["activity"], s.DB)
		}
		if err != nil {
			log.Println(err)
		}
	})
	http.HandleFunc("/pc/", func(w http.ResponseWriter, r *http.Request) {
		pc := strings.TrimSpace(r.URL.Path[len("/pc/"):])
		datastream.SetValue("pc", pc)
		storage.ReportEvent("pcActivity", pc)
		var err error
		if personal.GetActivity(datastream) {
			if s.States["place"].GetState() != "home" && !s.Flags["pc activity notify"] {
				s.Say("pc without master")
				storage.ReportEvent("pcActivityWithoutMe", pc)
				s.Flags["pc activity notify"] = true
				go func() {
					time.Sleep(2 * time.Hour)
					s.Flags["pc activity notify"] = false
				}()
			}
			err = s.Machines["activity"].Trigger("active", s.States["activity"], s.DB)
		} else {
			err = s.Machines["activity"].Trigger("idle", s.States["activity"], s.DB)
		}
		if err != nil {
			log.Println(err)
		}
	})
	go func() {
		log.Println("Start API")
		log.Println(http.ListenAndServe(":"+viper.GetString("port"), nil))
	}()
}

func (s *Shodan) LightOn(name string) {
	datastream.SendCommand(ds.Command{
		"light", map[string]interface{}{
			"name": name,
			"code": "On",
		}, "gideon", "Shodan",
	})
}

func (s *Shodan) UpdateGideon() {
	datastream.SendCommand(ds.Command{
		"update", nil, "gideon", "Shodan",
	})
}

func (s *Shodan) dispatchMessages(m string) {
	storage.ReportEvent("message", m)
	if strings.HasPrefix(m, "/") {
		tokens := strings.Split(m, " ")
		cmd := tokens[0][1:len(tokens[0])]
		args := tokens[1:]
		_ = args
		switch {
		case cmd == "ds":
			v := ds.Value{}
			datastream.Get(args[0], &v)
			if v.Value != nil {
				s.Say(v.Value.(string))
			} else {
				s.Say("wrong request")
			}
		case cmd == "echo":
			s.Say(strings.Join(args, " "))
		case cmd == "update":
			s.Say("update Gideon")
			s.UpdateGideon()
		case cmd == "lightOn" && len(args) > 0:
			s.LightOn(args[0])
		case cmd == "imat" && len(args) > 0:
			datastream.SetWhereIAm(args[0])
			err := s.Machines["place"].Trigger(args[0], s.States["place"], s.DB)
			if err != nil {
				log.Println(err)
			}
		case cmd == "time":
			s.Say(fmt.Sprintf("%s (%s)", time.Now().Format("15:04"), personal.GetDaytime()))
		case cmd == "debug":
			s.Flags["debug"] = true
			s.Say("debug on")
		case cmd == "w":
			w := weather.GetWeather()
			s.Say(fmt.Sprintf("%s - %v°", w.Weather, w.TempC))
		case cmd == "whereiam":
			s.Say(fmt.Sprintf("U r at %s", s.States["place"].GetState()))
		case cmd == "restart" && len(args) > 0:
			if args[0] == "gideon" {
				datastream.SendCommand(ds.Command{
					"kill", nil, "gideon", "Averrin",
				})
			}
		case cmd == "restart" && len(args) == 0:
			s.Say("Restarting...")
			os.Exit(1)
		case cmd == "cmd" && len(args) >= 2:
			datastream.SendCommand(ds.Command{
				args[1], nil, args[0], "Averrin",
			})
		case cmd == "status":
			for k, v := range s.States {
				s.Say(fmt.Sprintf("%s: %s", k, v.GetState()))
			}
			for k, v := range s.Flags {
				s.Say(fmt.Sprintf("%s: %v", k, v))
			}
			for k, v := range s.LastTimes {
				s.Say(fmt.Sprintf("%s: %s", k, v))
			}
		case cmd == "list":
			notes := storage.GetNotes()
			for _, n := range notes {
				s.Say(n.Text)
			}
		case cmd == "clear":
			storage.ClearNotes()
			s.Say("cleared")
		}
	} else {
		storage.SaveNote(m)
		s.Say("saved")
	}
}
