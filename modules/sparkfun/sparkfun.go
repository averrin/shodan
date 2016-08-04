package sparkfun

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const sfURL = "https://data.sparkfun.com"

type Place int

const (
	WORK = iota
	HOME
	NOWHERE
)

var places map[string]Place

type SparkFun map[string]map[string]string

func Connect(creds map[string]interface{}) SparkFun {
	sf := SparkFun{}
	for k, v := range creds {
		m := make(map[string]string)
		for key, value := range v.(map[interface{}]interface{}) {
			switch key := key.(type) {
			case string:
				switch value := value.(type) {
				case string:
					m[key] = value
				}
			}
		}
		sf[k] = m
	}

	places = map[string]Place{
		"work":    WORK,
		"home":    HOME,
		"nowhere": NOWHERE,
	}
	return sf
}

type RawPoint struct {
	Hum       string    `json:"hum"`
	Temp      string    `json:"temp"`
	Place     string    `json:"place"`
	Timestamp time.Time `json:"timestamp"`
}
type Point struct {
	Place     Place
	Timestamp time.Time
}

type Temp RawPoint
type Stream []RawPoint

func (sf SparkFun) GetWhereIAm() Point {
	stream := sf.GetStream("whereiam")
	return Point{
		Place:     places[stream[0].Place],
		Timestamp: stream[0].Timestamp,
	}
}

func (sf SparkFun) GetRoomTemp() RawPoint {
	stream := sf.GetStream("room")
	return stream[0]
}

func (sf SparkFun) SendRoomTemp(temp string, hum string) {
	sf.Send("room", map[string]string{
		"hum":  hum,
		"temp": temp,
	})
}

func (sf SparkFun) GetStream(name string) Stream {

	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/output/%s", sfURL, sf[name]["publicKey"]), nil)
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	stream := Stream{}
	json.Unmarshal(data, &stream)
	return stream
}

func (sf SparkFun) Send(name string, vars map[string]string) {
	client := &http.Client{}
	url := fmt.Sprintf("%s/input/%s?private_key=%s", sfURL, sf[name]["publicKey"], sf[name]["privateKey"])
	for k, v := range vars {
		url += fmt.Sprintf("&%s=%s", k, v)
	}
	req, err := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()
}
