package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/spf13/viper"
)

var mapping map[string]string

func (s *Shodan) initAPI() {
	mapping = map[string]string{
		"place/":   "imat",
		"cmd/":     "cmd",
		"display/": "phoneActivity",
		"pc/":      "pcActivity",
	}
	for route, command := range mapping {
		http.HandleFunc("/"+route, s.createHandler(route, command))
	}
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
	http.HandleFunc("/psb", func(w http.ResponseWriter, r *http.Request) {
		message, _ := ioutil.ReadAll(r.Body)
		s.Say(string(message))
		defer r.Body.Close()
	})
	http.HandleFunc("/codeship", func(w http.ResponseWriter, r *http.Request) {
		message, _ := ioutil.ReadAll(r.Body)
		s.Say(string(message))
		defer r.Body.Close()
	})
	http.HandleFunc("/alarm/", func(w http.ResponseWriter, r *http.Request) {
		sensor := strings.TrimSpace(r.URL.Path[len("/alarm/"):])
		storage.ReportEvent("alarm", sensor)
		if s.States["place"].GetState() == "home" {
			sensor = "at home"
		}
		s.Say(fmt.Sprintf("alarm %s", sensor))
	})
	go func() {
		log.Println("Start API")
		log.Println(http.ListenAndServe(":"+viper.GetString("port"), nil))
	}()
}

func (s *Shodan) createHandler(route string, command string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		tokens := strings.Split(r.URL.Path[len(route)+1:], "/")
		cmd := s.getCommand(command)
		if cmd.Cmd != "" {
			cmd.Action(tokens...)
		}
	}
}
