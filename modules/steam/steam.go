package steam

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type Steam map[string]string

func Connect(creds map[string]string) Steam {
	st := Steam{}
	for k, v := range creds {
		st[k] = v
	}
	return st
}

const HeaderURL = "http://cdn.akamai.steamstatic.com/steam/apps/%v/header.jpg"
const AppsListURL = "http://api.steampowered.com/IPlayerService/GetOwnedGames/v0001/?key=%s&steamid=%s&include_appinfo=1&include_played_free_games=1&format=json"

type Game struct {
	Appid                    int    `json:"appid"`
	HasCommunityVisibleStats bool   `json:"has_community_visible_stats"`
	ImgIconURL               string `json:"img_icon_url"`
	ImgLogoURL               string `json:"img_logo_url"`
	Name                     string `json:"name"`
	PlaytimeForever          int    `json:"playtime_forever"`
}

type SteamResponse struct {
	Response struct {
		GameCount int    `json:"game_count"`
		Games     []Game `json:"games"`
	} `json:"response"`
}

func (st Steam) GetGames() []Game {
	url := fmt.Sprintf(AppsListURL, st["apiKey"], st["steamID"])
	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	defer response.Body.Close()
	var r SteamResponse
	body, _ := ioutil.ReadAll(response.Body)
	json.Unmarshal(body, &r)
	return r.Response.Games
}
