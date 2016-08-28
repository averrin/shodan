package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
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
	http.HandleFunc("/location/", func(w http.ResponseWriter, r *http.Request) {
		location := r.URL.Path[len("/location/"):]
		storage.ReportEvent("location", location)
		datastream.SetValue("location", location)
	})
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
		message, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
		}
		psb := string(message)
		s.Say(s.processPSB(psb))
		defer r.Body.Close()
	})
	http.HandleFunc("/codeship", func(w http.ResponseWriter, r *http.Request) {
		message, _ := ioutil.ReadAll(r.Body)
		// s.Say(string(message))
		defer r.Body.Close()
		hook := CodeshipHook{}
		json.Unmarshal(message, &hook)
		if hook.Build.Status == "testing" {
			s.Say("build started")
		} else if hook.Build.Status == "success" {
			s.Say("build success")
		} else if hook.Build.Status == "error" {
			s.Say("build failed")
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

type CodeshipHook struct {
	Build struct {
		BuildURL        string `json:"build_url"`
		CommitURL       string `json:"commit_url"`
		ProjectID       int    `json:"project_id"`
		BuildID         int    `json:"build_id"`
		Status          string `json:"status"`
		ProjectName     string `json:"project_name"`
		ProjectFullName string `json:"project_full_name"`
		CommitID        string `json:"commit_id"`
		ShortCommitID   string `json:"short_commit_id"`
		Message         string `json:"message"`
		Committer       string `json:"committer"`
		Branch          string `json:"branch"`
		StartedAt       string `json:"started_at"`
		FinishedAt      string `json:"finished_at"`
	} `json:"build"`
}

func (s *Shodan) processPSB(psb string) string {
	re := regexp.MustCompile(`Доступно ([\d ]+).*`)
	amountRaw := re.FindStringSubmatch(psb)
	log.Println(amountRaw[0], amountRaw[1])
	amount, err := strconv.Atoi(strings.Replace(amountRaw[1], " ", ""))
	log.Println(amount, err)
	storage.ReportEvent("amount", fmt.Sprintf("%d", amount))
	return fmt.Sprintf("Доступно: %d", amount)
}
