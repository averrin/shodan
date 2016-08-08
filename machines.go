package main

import (
	"time"

	"github.com/jinzhu/gorm"
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
		s := state.(*PlaceState)
		shodan.Say("at home")
		if !teamviewer.GetPCStatus() {
			go func() {
				time.Sleep(15 * time.Minute)
				if !teamviewer.GetPCStatus() && s.GetState() == "home" {
					shodan.Say("at home, no pc")
				}
			}()
		}
		return nil
	}).Exit(func(state interface{}, tx *gorm.DB) error {
		return nil
	})
	m.State("work").Exit(func(state interface{}, tx *gorm.DB) error {
		// s := state.(*PlaceState)
		// telegram.Send("Хорошей дороги.")
		return nil
	})
	m.Event("home").To("home").From("nowhere")
	m.Event("work").To("work").From("nowhere")
	m.Event("village").To("village").From("nowhere")
	m.Event("nowhere").To("nowhere").From("work")
	m.Event("nowhere").To("nowhere").From("home")
	m.Event("nowhere").To("nowhere").From("village")
	return m
}

type PlaceState struct {
	transition.Transition
}
