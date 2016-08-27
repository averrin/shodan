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
	Action  func(...string)
	// URL     string
}
type Commands []Command

func (s *Shodan) getCommand(name string) Command {
	for _, c := range s.Commands {
		if c.Cmd == name {
			return c
		}
	}
	return Command{}
}

func (s *Shodan) getCommands() Commands {
	return Commands{
		{"ds", 1,
			func(args ...string) {
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
			func(args ...string) {
				s.Say(strings.Join(args, " "))
			},
		},
		{"update", 1,
			func(args ...string) {
				s.Say("update Gideon")
				s.UpdateGideon()
			},
		},
		{"lightOn", 1,
			func(args ...string) {
				s.LightOn(args[0])
			},
		},
		{"imat", 1,
			func(args ...string) {
				datastream.SetWhereIAm(args[0])
				err := s.Machines["place"].Trigger(args[0], s.States["place"], s.DB)
				if err != nil {
					log.Println(err)
				}
			},
		},
		{"time", 0,
			func(args ...string) {
				s.Say(fmt.Sprintf("%s (%s)", time.Now().Format("15:04"), personal.GetDaytime()))
			},
		},
		{"debug", 0,
			func(args ...string) {
				s.Flags["debug"] = true
				s.Say("debug on")
			},
		},
		{"w", 0,
			func(args ...string) {
				w := weather.GetWeather()
				s.Say(fmt.Sprintf("%s - %v°", w.Weather, w.TempC))
			},
		},
		{"whereiam", 0,
			func(args ...string) {
				s.Say(fmt.Sprintf("U r at %s", s.States["place"].GetState()))
			},
		},
		{"restart", 0,
			func(args ...string) {
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
			func(args ...string) {
				sign := fmt.Sprintf("%s.%s(%s)", args[0], args[1], strings.Join(args[2:], ", "))
				s.Say(sign)
				s.Say("sending command")
				storage.ReportEvent("command", sign)
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
			func(args ...string) {
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
			func(args ...string) {
				switch {
				case args[0] == "list":
					notes := storage.GetNotes()
					if len(notes) > 0 {
						s.Say("notes:")
						for _, n := range notes {
							s.Say(n.Text)
						}
					} else {
						s.Say("no notes")
					}
				case args[0] == "clear":
					storage.ClearNotes()
					s.Say("cleared")
				}
			},
		},
		{"pcActivity", 1,
			func(args ...string) {
				pc := args[0]
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
			},
		},
		{"phoneActivity", 1,
			func(args ...string) {
				display := args[0]
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
				c.Action(args...)
			}
		}
	} else {
		storage.SaveNote(m)
		s.Say("saved")
	}
}
