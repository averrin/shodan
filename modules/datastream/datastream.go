package datastream

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	redis "gopkg.in/redis.v4"
)

type DataStream map[string]string

var client *redis.Client

func Connect(creds map[string]string) *DataStream {
	ds := DataStream{}
	for k, v := range creds {
		ds[k] = v
	}
	ds.NewRedis()
	return &ds
}

func (ds *DataStream) NewRedis() {
	creds := *ds
	db, _ := strconv.Atoi(creds["db"])
	client = redis.NewClient(&redis.Options{
		Addr:     creds["host"],
		Password: creds["password"],
		DB:       db,
	})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)
}

type Point struct {
	Name      string
	Timestamp time.Time
}

func (ds *DataStream) GetWhereIAm() (point Point) {
	raw, err := client.Get("whereiam").Bytes()
	if err != nil {
		log.Print(err)
		return point
	}
	json.Unmarshal(raw, point)
	return point
}

func (ds *DataStream) SetWhereIAm(place string) {
	point := Point{
		Name:      place,
		Timestamp: time.Now(),
	}
	raw, _ := json.Marshal(point)
	client.Set("whereiam", raw, 0)
}
