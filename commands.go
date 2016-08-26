package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	ds "github.com/averrin/shodan/modules/datastream"
)

type Command struct {
	Cmd     string
	MinArgs int
	Action  func(string, []string)
}
type Commands []Command

func (s *Shodan) getCommands() Commands {
	return Commands{
		{"ds", 1,
			func(cmd string, args []string) {
				v := ds.Value{}
				datastream.Get(args[0], &v)
				if v.Value != nil {
					s.Say(v.Value.(string))
				} else {
					s.Say("wrong request")
				}
			},
		},
		{"echo", 1,
			func(cmd string, args []string) {
				s.Say(strings.Join(args, " "))
			},
		},
		{"update", 1,
			func(cmd string, args []string) {
				s.Say("update Gideon")
				s.UpdateGideon()
			},
		},
		{"lightOn", 1,
			func(cmd string, args []string) {
				s.LightOn(args[0])
			},
		},
		{"imat", 1,
			func(cmd string, args []string) {
				datastream.SetWhereIAm(args[0])
				err := s.Machines["place"].Trigger(args[0], s.States["place"], s.DB)
				if err != nil {
					log.Println(err)
				}
			},
		},
		{"time", 0,
			func(cmd string, args []string) {
				s.Say(fmt.Sprintf("%s (%s)", time.Now().Format("15:04"), personal.GetDaytime()))
			},
		},
		{"debug", 0,
			func(cmd string, args []string) {
				s.Flags["debug"] = true
				s.Say("debug on")
			},
		},
		{"w", 0,
			func(cmd string, args []string) {
				w := weather.GetWeather()
				s.Say(fmt.Sprintf("%s - %v°", w.Weather, w.TempC))
			},
		},
		{"whereiam", 0,
			func(cmd string, args []string) {
				s.Say(fmt.Sprintf("U r at %s", s.States["place"].GetState()))
			},
		},
		{"restart", 0,
			func(cmd string, args []string) {
				if len(args) == 0 {
					s.Say("Restarting...")
					os.Exit(1)
				}
				if args[0] == "gideon" {
					datastream.SendCommand(ds.Command{
						"kill", nil, "gideon", "Averrin",
					})
				}
			},
		},
		{"cmd", 2,
			func(cmd string, args []string) {
				s.Say("sending command")
				result := datastream.SendCommand(ds.Command{
					args[1], nil, args[0], "Averrin",
				})
				if result.Success {
					s.Say("command success")
				} else {
					s.Say("command fail")
				}
			},
		},
		{"status", 0,
			func(cmd string, args []string) {
				for k, v := range s.States {
					s.Say(fmt.Sprintf("%s: %s", k, v.GetState()))
				}
				for k, v := range s.Flags {
					s.Say(fmt.Sprintf("%s: %v", k, v))
				}
				for k, v := range s.LastTimes {
					s.Say(fmt.Sprintf("%s: %s", k, v))
				}
			},
		},
		{"notes", 1,
			func(cmd string, args []string) {
				switch {
				case args[0] == "list":
					notes := storage.GetNotes()
					for _, n := range notes {
						s.Say(n.Text)
					}
				case args[0] == "clear":
					storage.ClearNotes()
					s.Say("cleared")
				}
			},
		},
	}
}

func (s *Shodan) dispatchMessages(m string) {
	storage.ReportEvent("message", m)
	if strings.HasPrefix(m, "/") {
		tokens := strings.Split(m, " ")
		cmd := tokens[0][1:len(tokens[0])]
		args := tokens[1:]
		_ = args
		for _, c := range s.Commands {
			if cmd == c.Cmd && len(args) >= c.MinArgs {
				c.Action(cmd, args)
			}
		}
	} else {
		storage.SaveNote(m)
		s.Say("saved")
	}
}