package smarthome

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type Device struct {
	Renamed        bool   `json:"renamed"`
	Mac            string `json:"mac"`
	LearntCodeURL  string `json:"learntCodeUrl,omitempty"`
	Name           string `json:"name"`
	TypeCode       int    `json:"typeCode"`
	Lanaddr        string `json:"lanaddr"`
	TemperatureURL string `json:"temperatureUrl,omitempty"`
	Type           string `json:"type"`
	LearnURL       string `json:"learnUrl,omitempty"`
}

type Code struct {
	Repeat       int    `json:"repeat"`
	Order        int    `json:"order"`
	SendURL      string `json:"sendUrl"`
	DisplayName  string `json:"displayName"`
	Code         string `json:"code"`
	LearnedByMac string `json:"learnedByMac"`
	RemoteName   string `json:"remoteName"`
	CodeLength   int    `json:"codeLength"`
	ID           string `json:"id"`
	Name         string `json:"name"`
	Index        int    `json:"index"`
	RemoteType   int    `json:"remoteType"`
	Type         int    `json:"type"`
	Delay        int    `json:"delay"`
}

type Response struct {
	Msg       string `json:"msg"`
	Status    string `json:"status"`
	CodeID    string `json:"codeId"`
	DeviceMac string `json:"deviceMac"`
}

type Devices []Device
type Codes []Code

type SmartHome map[string]string

func Connect(creds map[string]string) (sh SmartHome) {
	for k, v := range creds {
		sh[k] = v
	}
	for _, d := range sh.GetDevices() {
		if d.Type == "RM2+" {
			sh["mac"] = d.Mac
			break
		}
	}
	return sh
}

func (sh SmartHome) Get(url string, value interface{}) {
	response, err := http.Get(url)
	defer response.Body.Close()
	if err != nil || response.StatusCode != 200 {
		b, _ := ioutil.ReadAll(response.Body)
		log.Println("SmartHome error", response.StatusCode, b)
		return
	}

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	err = json.Unmarshal(body, &value)
	if err != nil {
		log.Print(string(body))
		log.Fatal(err)
	}
}

func (sh SmartHome) GetDevices() (devices Devices) {
	url := fmt.Sprintf("http://%s/devices", sh["gateway"])
	sh.Get(url, devices)
	return devices
}

func (sh SmartHome) GetCodes() (codes Codes) {
	url := fmt.Sprintf("http://%s/codes", sh["gateway"])
	sh.Get(url, codes)
	return codes
}

func (sh SmartHome) SendCode(code Code) (r Response) {
	url := fmt.Sprintf("http://%s/send?deviceMac=%s&codeId=%s", sh["gateway"], code.LearnedByMac, code.ID)
	sh.Get(url, r)
	return r
}

func (sh SmartHome) GetCode(remote string, name string) (code Code) {
	codes := sh.GetCodes()
	for _, code = range codes {
		if code.RemoteName == remote && code.Name == name {
			return code
		}
	}
	return code
}
