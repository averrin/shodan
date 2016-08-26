package eventghost

import (
	"log"
	"time"

	"golang.org/x/net/websocket"
)

type EventGhost map[string]string

var socket *websocket.Conn

func Connect(creds map[string]string) *EventGhost {
	eg := EventGhost{}
	for k, v := range creds {
		eg[k] = v
	}
	eg.Connect()
	return &eg
}

func (eg *EventGhost) Connect() {
	creds := *eg
	var err error
	socket, err = websocket.Dial("ws://"+creds["host"], "", "http://"+creds["host"])
	if err != nil {
		go func() {
			for err != nil {
				log.Println(err)
				socket, err = websocket.Dial("ws://"+creds["host"], "", "http://"+creds["host"])
				time.Sleep(5 * time.Second)
			}
		}()
	}
}

func (eg *EventGhost) Send(msg string) {
	_, err := socket.Write([]byte(msg))
	if err != nil {
		log.Println(err)
	}
}
