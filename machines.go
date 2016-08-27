package main

import (
	"time"

	ds "github.com/averrin/shodan/modules/datastream"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/qor/transition"
)

func NewBinaryMachine(state *BinaryState, shodan *Shodan) *transition.StateMachine {
	wm := transition.New(state)
	wm.Initial("good")
	wm.State("bad").Enter(func(state interface{}, tx *gorm.DB) error {
		s := state.(*BinaryState)
		shodan.Say(s.BadName)
		return nil
	}).Exit(func(state interface{}, tx *gorm.DB) error {
		s := state.(*BinaryState)
		shodan.Say(s.GoodName)
		return nil
	})
	wm.Event("to_good").To("good").From("bad")
	wm.Event("to_bad").To("bad").From("good")
	return wm
}

type BinaryState struct {
	GoodName string
	BadName  string

	transition.Transition
}

func NewPlaceMachine(state *PlaceState, shodan *Shodan) *transition.StateMachine {
	m := transition.New(state)
	m.Initial("nowhere")
	m.State("nowhere").Enter(func(state interface{}, tx *gorm.DB) error {
		s := state.(*PlaceState)
		go func() {
			time.Sleep(5 * time.Minute)
			if s.GetState() == "nowhere" {
				shodan.Say("good way")
			}
		}()
		return nil
	}).Exit(func(state interface{}, tx *gorm.DB) error {
		if shodan.LastTimes["home leave"].IsZero() {
			shodan.LastTimes["home leave"] = time.Now()
		}
		return nil
	})
	m.State("village")
	m.State("pavel")
	m.State("home").Enter(func(state interface{}, tx *gorm.DB) error {
		s := state.(*PlaceState)
		shodan.Say("at home")
		if time.Now().Sub(shodan.LastTimes["home leave"]).Minutes() > 15 {
			shodan.LastTimes["home leave"] = time.Time{}
			datastream.SendCommand(ds.Command{
				"sh:Прихожая 1:On", nil, "gideon", "Shodan",
			})
			datastream.SendCommand(ds.Command{
				"sh:Прихожая 2:On", nil, "gideon", "Shodan",
			})
			datastream.SendCommand(ds.Command{
				"sh:Alarm:Unlock", nil, "gideon", "Shodan",
			})
		}
		pcStatus := ds.Value{}
		datastream.Get("pc", &pcStatus)
		if pcStatus.Value == "off" {
			go func() {
				time.Sleep(15 * time.Minute)
				datastream.Get("pc", &pcStatus)
				if pcStatus.Value == "off" && s.GetState() == "home" {
					shodan.Say("at home, no pc")
				}
			}()
		}
		return nil
	}).Exit(func(state interface{}, tx *gorm.DB) error {
		s := state.(*PlaceState)
		shodan.LastTimes["home leave"] = time.Now()
		go func() {
			time.Sleep(5 * time.Minute)
			if s.GetState() != "home" {
				datastream.SendCommand(ds.Command{
					"sh:Alarm:Lock", nil, "gideon", "Shodan",
				})
			}
		}()
		return nil
	})
	m.State("work").Exit(func(state interface{}, tx *gorm.DB) error {
		// s := state.(*PlaceState)
		shodan.Flags["late at work"] = false
		return nil
	})
	m.Event("home").To("home").From("nowhere", "work", "village", "pavel").Before(beforePlace).After(afterPlace)
	m.Event("work").To("work").From("nowhere", "home", "village", "pavel").Before(beforePlace).After(afterPlace)
	m.Event("pavel").To("pavel").From("nowhere", "home", "village", "work").Before(beforePlace).After(afterPlace)
	m.Event("village").To("village").From("nowhere", "home", "pavel", "work").Before(beforePlace).After(afterPlace)
	m.Event("nowhere").To("nowhere").From("work", "home", "village", "pavel").Before(beforePlace).After(afterPlace)
	return m
}

func beforePlace(state interface{}, tx *gorm.DB) error {
	s := state.(*PlaceState)
	storage.ReportEvent("leave", s.GetState())
	return nil
}

func afterPlace(state interface{}, tx *gorm.DB) error {
	s := state.(*PlaceState)
	storage.ReportEvent("enter", s.GetState())
	return nil
}

func NewDayTimeMachine(state *DayTimeState, shodan *Shodan) *transition.StateMachine {
	m := transition.New(state)
	m.Initial("day")

	m.State("day")
	m.Event("day").To("day").From("morning")

	m.State("night").Enter(func(state interface{}, tx *gorm.DB) error {
		a := shodan.States["activity"]
		s := state.(*DayTimeState)
		if s.GetState() == "night" {
			go func() {
				time.Sleep(5 * time.Minute)
				if a.GetState() == "active" {
					shodan.Say("activity at night")
				}
			}()
		}
		return nil
	})
	m.Event("night").To("night").From("evening")

	m.State("evening")
	m.Event("evening").To("evening").From("day")

	m.State("morning")
	m.Event("morning").To("morning").From("night")
	return m
}

func NewActivityMachine(state *ActivityState, shodan *Shodan) *transition.StateMachine {
	shodan.Flags["late night notify"] = false
	wm := transition.New(state)
	wm.Initial("idle")
	wm.State("active").Enter(func(state interface{}, tx *gorm.DB) error {
		s := state.(*ActivityState)
		if personal.GetDaytime() == "night" {
			go func() {
				time.Sleep(5 * time.Minute)
				if s.GetState() == "active" && !shodan.Flags["late night notify"] {
					shodan.Say("activity at night")
					storage.ReportEvent("nightActivity", "")
					shodan.Flags["late night notify"] = true
					go func() {
						time.Sleep(1 * time.Hour)
						shodan.Flags["late night notify"] = false
					}()
				}
			}()
		}
		return nil
	})
	wm.Event("idle").To("idle").From("active")
	wm.Event("active").To("active").From("idle")
	return wm
}

func NewSleepMachine(state *SleepState, shodan *Shodan) *transition.StateMachine {
	wm := transition.New(state)
	wm.Initial("awake")
	wm.State("awake").Enter(func(state interface{}, tx *gorm.DB) error {
		shodan.Say("awake")
		return nil
	})
	wm.State("sleep")
	wm.State("dream").Enter(func(state interface{}, tx *gorm.DB) error {
		go func() {
			time.Sleep(3 * time.Minute)
			shodan.Say("good morning")
		}()
		s := state.(*SleepState)
		go func() {
			time.Sleep(30 * time.Minute)
			if s.GetState() == "dream" {
				shodan.Say("get up now")
				go func() {
					time.Sleep(3 * time.Minute)
					if s.GetState() == "dream" {
						shodan.Say("you were alerted")
						datastream.SendCommand(ds.Command{
							"sh:Спальня:On", nil, "gideon", "Shodan",
						})
					}
				}()
			}
		}()
		return nil
	})
	wm.Event("sleep").To("sleep").From("awake")
	wm.Event("dream").To("dream").From("sleep")
	wm.Event("awake").To("awake").From("dream", "sleep")
	return wm
}

type PlaceState struct {
	transition.Transition
}

type DayTimeState struct {
	transition.Transition
}

type ActivityState struct {
	transition.Transition
}

type SleepState struct {
	transition.Transition
}
