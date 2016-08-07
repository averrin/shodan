package main

import "math/rand"

type ShodanString []string

var ShodanStrings map[string]ShodanString

func (s ShodanString) Get() string {
	return s[rand.Intn(len(s))]
}
