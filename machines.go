package main

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/transition"
)

func NewBinaryMachine(state *BinaryState) *transition.StateMachine {
	wm := transition.New(state)
	wm.Initial("good")
	wm.State("bad").Enter(func(state interface{}, tx *gorm.DB) error {
		s := state.(*BinaryState)
		// pushbullet.SendPush(s.BadTitle, s.BadBody)
		telegram.Send(fmt.Sprintf("%s\n%s", s.BadTitle, s.BadBody))
		return nil
	}).Exit(func(state interface{}, tx *gorm.DB) error {
		s := state.(*BinaryState)
		// pushbullet.SendPush(s.GoodTitle, s.GoodBody)
		telegram.Send(fmt.Sprintf("%s\n%s", s.GoodTitle, s.GoodBody))
		return nil
	})
	wm.Event("to_good").To("good").From("bad")
	wm.Event("to_bad").To("bad").From("good")
	return wm
}

type BinaryState struct {
	GoodTitle string
	GoodBody  string
	BadTitle  string
	BadBody   string

	transition.Transition
}

func NewPlaceMachine(state *PlaceState) *transition.StateMachine {
	m := transition.New(state)
	m.Initial("nowhere")
	m.State("nowhere")
	m.State("village")
	m.State("pavel")
	m.State("home").Enter(func(state interface{}, tx *gorm.DB) error {
		s := state.(*PlaceState)
		telegram.Send("Ты наконец дома, ура!")
		if !teamviewer.GetPCStatus() {
			go func() {
				time.Sleep(15 * time.Minute)
				if !teamviewer.GetPCStatus() && s.GetState() == "home" {
					telegram.Send("Ты 15 минут дома, а комп не включен. Все в порядке?")
				}
			}()
		}
		return nil
	}).Exit(func(state interface{}, tx *gorm.DB) error {
		// s := state.(*PlaceState)
		telegram.Send("Хорошей дороги.")
		return nil
	})
	m.State("work").Exit(func(state interface{}, tx *gorm.DB) error {
		// s := state.(*PlaceState)
		telegram.Send("Хорошей дороги.")
		return nil
	})
	m.Event("home").To("home").From("nowhere")
	m.Event("work").To("work").From("nowhere")
	return m
}

type PlaceState struct {
	transition.Transition
}
