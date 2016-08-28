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
			}, "go dinner",
		},
		{
			func() bool {
				d := time.Now()
				return d.Day() == 22 && d.Month() == 3 && d.Hour() == 12
			}, "master birthday",
		},
		{
			func() bool {
				d := time.Now()
				return d.Day() >= 23 && d.Day() <= 25 && d.Hour() == 22
			}, "Отправь счетчики!",
		},
		{
			func() bool {
				d := time.Now()
				return d.Day() == 3 && d.Month() == 12 && d.Hour() == 12
			}, "mish birthday",
		},
		{
			func() bool {
				d := time.Now()
				return d.Day() == 6 && d.Month() == 11 && d.Hour() == 12
			}, "pavel birthday",
		},
		{
			func() bool {
				d := time.Now()
				return d.Day() == 25 && d.Month() == 8 && d.Hour() == 12
			}, "ilia birthday",
		},
		{
			func() bool {
				d := time.Now()
				return d.Day() == 18 && d.Month() == 8 && d.Hour() == 12
			}, "papa birthday",
		},
		{
			func() bool {
				d := time.Now()
				return d.Day() == 1 && d.Month() == 7 && d.Hour() == 12
			}, "sister birthday",
		},
		{
			func() bool {
				d := time.Now()
				return d.Day() == 21 && d.Month() == 4 && d.Hour() == 12
			}, "mama birthday",
		},
		{
			func() bool {
				d := time.Now()
				return d.Day() == 25 && d.Month() == 12 && d.Hour() == 12
			}, "misha birthday",
		},
		{
			func() bool {
				d := time.Now()
				return d.Day() == 25 && d.Month() == 5 && d.Hour() == 12
			}, "zoya birthday",
		},
		{
			func() bool {
				d := time.Now()
				return d.Day() == 1 && d.Month() == 5 && d.Hour() == 12
			}, "elem birthday",
		},
		{
			func() bool {
				d := time.Now()
				return d.Day() == 3 && d.Month() == 7 && d.Hour() == 12
			}, "slava birthday",
		},
		{
			func() bool {
				d := time.Now()
				return d.Day() == 1 && d.Month() == 8 && d.Hour() == 12
			}, "shodan birthday",
		},
	}
}
