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

func (ds *DataStream) Heartbeat(key string) {
	ticker := time.NewTicker(2 * time.Second)
	channel := fmt.Sprintf("heartbeat:%s", key)
	go func() {
		for _ = range ticker.C {
			client.Publish(channel, fmt.Sprintf("%v", time.Now()))
		}
	}()
}

func (ds *DataStream) GetHeartbeat(key string) chan string {
	channel := fmt.Sprintf("heartbeat:%s", key)
	pubsub, _ := client.Subscribe(channel)
	out := make(chan string)
	go func() {
		defer pubsub.Close()
		for {
			msg, err := pubsub.ReceiveMessage()
			if err != nil {
				close(out)
				break
			}
			out <- msg.Payload
		}
	}()
	return out
}

type Point struct {
	Name      string
	Timestamp time.Time
}

type Online struct {
	Name      string
	Online    bool
	Timestamp time.Time
}

type RoomTemp struct {
	Hum       string
	Temp      string
	Timestamp time.Time
}

func (ds *DataStream) Get(key string, value interface{}) {
	raw, err := client.Get(key).Bytes()
	if err != nil {
		log.Print(err)
	}
	json.Unmarshal(raw, value)
}

func (ds *DataStream) Set(key string, value interface{}) {
	raw, _ := json.Marshal(value)
	err := client.Set(key, raw, 0).Err()
	if err != nil {
		log.Println(err)
	}
}

func (ds *DataStream) GetWhereIAm() (point Point) {
	ds.Get("whereiam", &point)
	return point
}

func (ds *DataStream) GetRoomTemp() (temp RoomTemp) {
	ds.Get("roomtemp", &temp)
	return temp
}

func (ds *DataStream) SetWhereIAm(place string) {
	point := Point{
		Name:      place,
		Timestamp: time.Now(),
	}
	ds.Set("whereiam", point)
}

func (ds *DataStream) SetRoomTemp(temp string, hum string) {
	point := RoomTemp{
		Temp:      temp,
		Hum:       hum,
		Timestamp: time.Now(),
	}
	ds.Set("roomtemp", point)
}

func (ds *DataStream) SetOnline(key string, online bool) {
	point := Online{
		Name:      key,
		Online:    online,
		Timestamp: time.Now(),
	}
	ds.Set(key, point)
}