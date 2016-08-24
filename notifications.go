package main

import (
	"time"

	p "github.com/averrin/shodan/modules/personal"
)

type Notification struct {
	Test func() bool
	Text string
}

type Notifications []Notification

func (s *Shodan) getNotifications() Notifications {
	return Notifications{
		{
			func() bool {
				return personal.GetDay() == p.WORKDAY && time.Now().Hour() == 16
			}, "Если еще не пообедал - марш!",
		},
		{
			func() bool {
				d := time.Now()
				return d.Day() == 22 && d.Month() == 3 && d.Hour() == 12
			}, "С днем рождения, котяра!",
		},
	}
}
