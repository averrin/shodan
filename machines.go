package main

import (
	"time"

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
	})
	m.State("village")
	m.State("pavel")
	m.State("home").Enter(func(state interface{}, tx *gorm.DB) error {
		// s := state.(*PlaceState)
		shodan.Say("at home")
		// if !teamviewer.GetPCStatus() {
		// 	go func() {
		// 		time.Sleep(15 * time.Minute)
		// 		if !teamviewer.GetPCStatus() && s.GetState() == "home" {
		// 			shodan.Say("at home, no pc")
		// 		}
		// 	}()
		// }
		return nil
	}).Exit(func(state interface{}, tx *gorm.DB) error {
		return nil
	})
	m.State("work").Exit(func(state interface{}, tx *gorm.DB) error {
		// s := state.(*PlaceState)
		shodan.Flags["late at work"] = false
		return nil
	})
	m.Event("home").To("home").From("nowhere", "work", "village", "pavel")
	m.Event("work").To("work").From("nowhere", "home", "village", "pavel")
	m.Event("pavel").To("pavel").From("nowhere", "home", "village", "work")
	m.Event("village").To("village").From("nowhere", "home", "pavel", "work")
	m.Event("nowhere").To("nowhere").From("work", "home", "village", "pavel")
	return m
}

func NewDayTimeMachine(state *DayTimeState, shodan *Shodan) *transition.StateMachine {
	m := transition.New(state)
	m.Initial("day")

	m.State("day")
	m.Event("day").To("day").From("morning")

	m.State("night")
	m.Event("night").To("night").From("evening")

	m.State("evening")
	m.Event("evening").To("evening").From("day")

	m.State("morning")
	m.Event("morning").To("morning").From("night")
	return m
}

func NewActivityMachine(state *ActivityState, shodan *Shodan) *transition.StateMachine {
	wm := transition.New(state)
	wm.Initial("idle")
	wm.State("active")
	wm.Event("idle").To("active").From("idle")
	wm.Event("active").To("idle").From("active")
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
