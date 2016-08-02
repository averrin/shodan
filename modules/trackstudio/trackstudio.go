package trackstudio

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

//TimeReport is ts response struct
type TimeReport struct {
	Flags          string `json:"flags"`
	Location       string `json:"location"`
	ReportedTime   int    `json:"reported_time"`
	UserDepartment string `json:"user_department"`
	UserEmail      string `json:"user_email"`
	UserID         string `json:"user_id"`
	UserName       string `json:"user_name"`
	UserSkype      string `json:"user_skype"`
}

type TrackStudio map[string]string

func Connect(creds map[string]string) TrackStudio {
	ts := TrackStudio{}
	for k, v := range creds {
		ts[k] = v
	}
	return ts
}

func (ts TrackStudio) getReported(dateFrom string, dateTo string) int {
	creds := fmt.Sprintf("template_login=%s&template_password=%s", ts["login"], ts["password"])
	var reports []TimeReport
	url := fmt.Sprintf("%v?%v&view=person&date_sel=today&date_from=%v&date_to=%v&person_id=%v",
		ts["endpoint"], creds,
		dateFrom, dateTo,
		ts["userID"])
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	// fmt.Println(url)

	Decoder := json.NewDecoder(resp.Body)
	Decoder.Decode(&reports)
	if len(reports) == 0 {
		return 0
	}
	return reports[0].ReportedTime
}

func (ts TrackStudio) GetReportedToday() int {
	now := time.Now()
	tomorrow := now.Add(time.Duration(24 * time.Hour))
	dateFrom := now.Format("2006-01-02")
	dateTo := tomorrow.Format("2006-01-02")
	return ts.getReported(dateFrom, dateTo)
}

func (ts TrackStudio) GetReportedYesterday() int {
	now := time.Now()
	yesterday := now.Add(-time.Duration(24 * time.Hour))
	dateTo := now.Format("2006-01-02")
	dateFrom := yesterday.Format("2006-01-02")
	return ts.getReported(dateFrom, dateTo)
}

func (ts TrackStudio) GetReportedCurrentWeek() int {
	now := time.Now()
	k := int(now.Weekday())
	monday := now.Add(-time.Duration((k - 1) * 24 * int(time.Hour)))
	sunday := now.Add(time.Duration(24 * int(time.Hour) * (7 - int(now.Weekday()) - 1)))
	dateFrom := monday.Format("2006-01-02")
	dateTo := sunday.Format("2006-01-02")
	return ts.getReported(dateFrom, dateTo)
}

func (ts TrackStudio) GetReportedLastWeek() int {
	now := time.Now()
	k := int(now.Weekday())
	monday := now.Add(-time.Duration((k - 1) * 24 * int(time.Hour))).Add(-7 * 24 * time.Hour)
	sunday := now.Add(time.Duration(24 * int(time.Hour) * (7 - int(now.Weekday()) - 1))).Add(-7 * 24 * time.Hour)
	dateFrom := monday.Format("2006-01-02")
	dateTo := sunday.Format("2006-01-02")
	return ts.getReported(dateFrom, dateTo)
}
