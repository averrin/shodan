package pushbullet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const baseURL = "https://api.pushbullet.com/v2"

type Pushbullet map[string]string

func Connect(creds map[string]string) Pushbullet {
	pb := Pushbullet{}
	for k, v := range creds {
		pb[k] = v
	}
	return pb
}

type Push struct {
	Active                  bool    `json:"active"`
	Body                    string  `json:"body"`
	Created                 float64 `json:"created"`
	Direction               string  `json:"direction"`
	Dismissed               bool    `json:"dismissed"`
	Iden                    string  `json:"iden"`
	Modified                float64 `json:"modified"`
	ReceiverEmail           string  `json:"receiver_email"`
	ReceiverEmailNormalized string  `json:"receiver_email_normalized"`
	ReceiverIden            string  `json:"receiver_iden"`
	SenderEmail             string  `json:"sender_email"`
	SenderEmailNormalized   string  `json:"sender_email_normalized"`
	SenderIden              string  `json:"sender_iden"`
	SenderName              string  `json:"sender_name"`
	Title                   string  `json:"title"`
	Type                    string  `json:"type"`
}

type Response struct {
	Pushes []Push `json:"pushes"`
}

func (pb Pushbullet) SendPush(title string, body string) {
	url := fmt.Sprintf("%s/%s", baseURL, "pushes")
	jsonStr := []byte(fmt.Sprintf(`{"type":"note", "title": "%s", "body": "%s"}`,
		title, body))
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Access-Token", pb["apiKey"])
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil || response.StatusCode != 200 {
		log.Print("Pushbullet error ", err)
	}
	defer response.Body.Close()
	resp, _ := ioutil.ReadAll(response.Body)
	log.Println(string(resp))
}

func (pb Pushbullet) GetPushes() []Push {
	var p []Push
	url := fmt.Sprintf("%s/%s", baseURL, "pushes")
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Access-Token", pb["apiKey"])
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil || response.StatusCode != 200 {
		log.Print("Pushbullet error ", err)
		return p
	}
	var r Response
	body, _ := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	err = json.Unmarshal(body, &r)
	if err != nil {
		log.Print(string(body))
		log.Fatal(err)
	}
	p = r.Pushes
	return p
}
