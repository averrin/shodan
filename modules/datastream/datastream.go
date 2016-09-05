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
			now, _ := time.Now().MarshalText()
			client.Publish(channel, string(now))
		}
	}()
}

type Status struct {
	Sender    string
	Timestamp time.Time
	Success   bool
	Result    interface{}
}

func (ds *DataStream) SendStatus(status Status) {
	raw, _ := json.Marshal(status)
	client.Publish(fmt.Sprintf("status:%s", status.Sender), string(raw))
}

func (ds *DataStream) GetHeartbeat(key string) chan bool {
	channel := fmt.Sprintf("heartbeat:%s", key)
	pubsub, _ := client.Subscribe(channel)
	out := make(chan bool)
	ticker := time.NewTicker(2 * time.Second)
	var lt time.Time
	go func() {
		defer pubsub.Close()
		for {
			msg, err := pubsub.ReceiveMessage()
			if err != nil {
				close(out)
				break
			}
			lt.UnmarshalText([]byte(msg.Payload))
		}
	}()
	go func() {
		for _ = range ticker.C {
			if time.Now().Sub(lt) > time.Duration(3*time.Second) {
				out <- false
			} else {
				out <- true
			}
		}
	}()
	return out
}

type Point struct {
	Name      string
	Timestamp time.Time
}

type Value struct {
	Value     interface{}
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

func (ds *DataStream) Get(key string, value interface{}) error {
	raw, err := client.Get(key).Bytes()
	if err != nil {
		// log.Print("get ", key, err)
		return err
	}
	json.Unmarshal(raw, value)
	return nil
}

func (ds *DataStream) Set(key string, value interface{}) error {
	raw, _ := json.Marshal(value)
	err := client.Set(key, raw, 0).Err()
	if err != nil {
		log.Println("set ", key, err)
		return err
	}
	return nil
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

func (ds *DataStream) SetValue(key string, value string) {
	v := Value{
		Value:     value,
		Timestamp: time.Now(),
	}
	ds.Set(key, v)
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

type Command struct {
	Name     string
	Args     map[string]interface{}
	Reciever string
	Sender   string
}

func (ds *DataStream) GetCommands(key string) (out chan Command) {
	channel := fmt.Sprintf("commands:%s", key)
	pubsub, _ := client.Subscribe(channel)
	out = make(chan Command)
	go func() {
		defer pubsub.Close()
		for {
			msg, err := pubsub.ReceiveMessage()
			if err != nil {
				close(out)
				break
			}
			c := Command{}
			err = json.Unmarshal([]byte(msg.Payload), &c)
			if err != nil {
				log.Println(err)
				continue
			}
			out <- c
		}
	}()
	return out
}

func (ds *DataStream) SendCommand(cmd Command) Status {
	raw, _ := json.Marshal(cmd)
	channel := fmt.Sprintf("status:%s", cmd.Reciever)
	pubsub, _ := client.Subscribe(channel)
	out := make(chan Status)
	defer close(out)
	waiting := true
	go func() {
		defer pubsub.Close()
		for {
			msg, err := pubsub.ReceiveMessage()
			status := Status{}
			if err == nil {
				json.Unmarshal([]byte(msg.Payload), &status)
			} else {
				status.Success = false
			}
			if waiting {
				waiting = false
				out <- status
			}
			break
		}
	}()
	go func() {
		time.Sleep(5 * time.Minute)
		if waiting {
			waiting = false
			out <- Status{}
			pubsub.Close()
		}
	}()
	client.Publish(fmt.Sprintf("commands:%s", cmd.Reciever), string(raw))
	return <-out
}
