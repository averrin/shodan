package teamviewer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
)

type TeamViewer map[string]string

func Connect(creds map[string]string) TeamViewer {
	tv := TeamViewer{}
	for k, v := range creds {
		tv[k] = v
	}
	c := tv.Auth()
	tv["access_token"] = c.AccessToken
	tv["refresh_token"] = c.RefreshToken
	return tv
}

const baseURL = "https://webapi.teamviewer.com/api/v1"

func (tv TeamViewer) Auth() Creds {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	credPath := path.Join(dir, ".tv_credentials")
	if _, err := os.Stat(credPath); err == nil {
		data, err := ioutil.ReadFile(credPath)
		if err != nil {
			log.Fatal(err)
		}
		c := Creds{}
		json.Unmarshal(data, &c)
		return c
	}
	return tv.GetNewCreds()
}

func (tv TeamViewer) RefreshGetNewCreds(Creds) Creds {
}

func (tv TeamViewer) GetNewCreds() Creds {
	tokenURL := fmt.Sprintf("%s/%s", baseURL, "oauth2/token")
	args := url.Values{"code": {tv["code"]}, "grant_type": {"authorization_code"},
		"client_id": {tv["clientID"]}, "client_secret": {tv["clientSecret"]},
	}.Encode()
	log.Println(tokenURL, args)
	req, _ := http.NewRequest("POST", tokenURL,
		bytes.NewBufferString(args))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		log.Print("TeamViewer error ", err)
	}
	defer response.Body.Close()
	resp, _ := ioutil.ReadAll(response.Body)
	creds := Creds{}
	json.Unmarshal(resp, &creds)
	if creds.AccessToken == "" {
		log.Print("TeamViewer error:")
		log.Println(response.StatusCode, string(resp))
		return Creds{}
	}
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	credPath := path.Join(dir, ".tv_credentials")
	ioutil.WriteFile(credPath, resp, 0644)
	return creds
}

type Creds struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}
