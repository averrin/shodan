package weather

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const wuURL = "http://api.wunderground.com/api/%s/conditions/q/%s.json"

type CurrentObservation struct {
	// UV               string   `json:"UV"`
	// DewpointC        int      `json:"dewpoint_c"`
	// DewpointF        int      `json:"dewpoint_f"`
	// DewpointString   string   `json:"dewpoint_string"`
	// Estimated        struct{} `json:"estimated"`
	FeelslikeC string `json:"feelslike_c"`
	// FeelslikeF       string   `json:"feelslike_f"`
	// FeelslikeString  string   `json:"feelslike_string"`
	// ForecastURL      string   `json:"forecast_url"`
	// HeatIndexC       int      `json:"heat_index_c"`
	// HeatIndexF       int      `json:"heat_index_f"`
	// HeatIndexString  string   `json:"heat_index_string"`
	// HistoryURL       string   `json:"history_url"`
	Icon string `json:"icon"`
	// IconURL          string   `json:"icon_url"`
	// LocalEpoch       string   `json:"local_epoch"`
	// LocalTimeRfc822  string   `json:"local_time_rfc822"`
	// LocalTzLong      string   `json:"local_tz_long"`
	// LocalTzOffset    string   `json:"local_tz_offset"`
	// LocalTzShort     string   `json:"local_tz_short"`
	// Nowcast          string   `json:"nowcast"`
	// ObURL            string   `json:"ob_url"`
	// ObservationEpoch string   `json:"observation_epoch"`
	// ObservationLocation struct {
	// 	City           string `json:"city"`
	// 	Country        string `json:"country"`
	// 	CountryIso3166 string `json:"country_iso3166"`
	// 	Elevation      string `json:"elevation"`
	// 	Full           string `json:"full"`
	// 	Latitude       string `json:"latitude"`
	// 	Longitude      string `json:"longitude"`
	// 	State          string `json:"state"`
	// } `json:"observation_location"`
	// ObservationTime       string  `json:"observation_time"`
	// ObservationTimeRfc822 string  `json:"observation_time_rfc822"`
	// Precip1hrIn           string  `json:"precip_1hr_in"`
	// Precip1hrMetric       string  `json:"precip_1hr_metric"`
	// Precip1hrString       string  `json:"precip_1hr_string"`
	// PrecipTodayIn         string  `json:"precip_today_in"`
	// PrecipTodayMetric     string  `json:"precip_today_metric"`
	// PrecipTodayString     string  `json:"precip_today_string"`
	// PressureIn            string  `json:"pressure_in"`
	// PressureMb            string  `json:"pressure_mb"`
	// PressureTrend         string  `json:"pressure_trend"`
	RelativeHumidity string `json:"relative_humidity"`
	// Solarradiation        string  `json:"solarradiation"`
	// StationID             string  `json:"station_id"`
	TempC float32 `json:"temp_c"`
	TempF float32 `json:"temp_f"`
	// TemperatureString     string  `json:"temperature_string"`
	// VisibilityKm          string  `json:"visibility_km"`
	// VisibilityMi          string  `json:"visibility_mi"`
	Weather string `json:"weather"`
	// WindDegrees           float32 `json:"wind_degrees"`
	// WindDir               string  `json:"wind_dir"`
	// WindKph               float32 `json:"wind_kph"`
	// WindMph               float32 `json:"wind_mph"`
	// WindString            string  `json:"wind_string"`
	// WindchillC            string  `json:"windchill_c"`
	// WindchillF            string  `json:"windchill_f"`
	// WindchillString       string  `json:"windchill_string"`
}

// WeatherResponse from wunderweather json
type WeatherResponse struct {
	CurrentObservation `json:"current_observation"`
	Response           struct {
		Features struct {
			Conditions int `json:"conditions"`
		} `json:"features"`
		TermsofService string `json:"termsofService"`
		Version        string `json:"version"`
	} `json:"response"`
}

type WUnderground map[string]string

func Connect(creds map[string]string) WUnderground {
	wu := WUnderground{}
	for k, v := range creds {
		wu[k] = v
	}
	return wu
}

func (wu WUnderground) GetWeather() CurrentObservation {
	var w CurrentObservation
	url := fmt.Sprintf(wuURL, wu["apiKey"], wu["location"])
	response, err := http.Get(url)
	if err != nil || response.StatusCode != 200 {
		log.Print("Weather error")
		return w
	}

	defer response.Body.Close()
	var r WeatherResponse
	body, _ := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	err = json.Unmarshal(body, &r)
	if err != nil {
		log.Print(string(body))
		log.Fatal(err)
	}
	// fmt.Printf("%v - %v", r.CurrentObservation.Icon, r.CurrentObservation.TempC)
	w = r.CurrentObservation
	// w.TempC = rand.Int()
	return w
}
