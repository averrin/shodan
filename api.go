package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	ds "github.com/averrin/shodan/modules/datastream"
	"github.com/spf13/viper"
)

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
		s.Say("sending command")
		s.Say(tokens[0])
		result := datastream.SendCommand(ds.Command{
			tokens[1], nil, tokens[0], "Shodan",
		})
		if result.Success {
			s.Say("command success")
		} else {
			s.Say("command fail")
		}
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
	http.HandleFunc("/alarm/", func(w http.ResponseWriter, r *http.Request) {
		sensor := strings.TrimSpace(r.URL.Path[len("/alarm/"):])
		storage.ReportEvent("alarm", sensor)
		if s.States["place"].GetState() == "home" {
			sensor = "at home"
		}
		s.Say(fmt.Sprintf("alarm %s", sensor))
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
