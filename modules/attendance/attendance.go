package attendance

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Jeffail/gabs"
)

type Attendance map[string]string

func Connect(creds map[string]string) Attendance {
	ts := Attendance{}
	for k, v := range creds {
		ts[k] = v
	}
	return ts
}

func (att Attendance) parseAttendance(raw *gabs.Container) *Info {

	ch, _ := raw.Children()
	var data *gabs.Container
	for _, c := range ch {
		ch, _ := c.Children()
		for _, d := range ch {
			data = d
			break
		}
		break
	}
	info := new(Info)
	// log.Print(info.FillStruct(data.Data().(map[string]interface{})))
	json.Unmarshal([]byte(data.String()), &info)
	info.Days = make([]*Day, 32)
	d, _ := data.ChildrenMap()
	for k, v := range d {
		if i, err := strconv.Atoi(k); err == nil {
			day := new(Day)
			json.Unmarshal([]byte(v.String()), &day)
			info.Days[i] = day
		}
	}
	return info
}

func (att Attendance) GetAttendance() *Info {
	data, err := att.getAttendanceJSON()
	if err != nil {
		return new(Info)
	}
	return att.parseAttendance(data)
}

func (att Attendance) getAttendanceJSON() (*gabs.Container, error) {
	cookie := &http.Cookie{
		Name:   "remember_token",
		Value:  att["cookie"],
		Path:   "/",
		Domain: att["domain"],
	}
	// https://github.com/golang/go/issues/4800
	checkRedirect := func(req *http.Request, via []*http.Request) error {
		for attr, val := range via[0].Header {
			if _, ok := req.Header[attr]; !ok {
				req.Header[attr] = val
			}
		}
		return nil
	}
	client := &http.Client{
		CheckRedirect: checkRedirect,
	}
	req, err := http.NewRequest("GET", "http://"+att["domain"], nil)
	req.Header.Set("Accept", "application/json")
	req.AddCookie(cookie)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	return gabs.ParseJSON(data)
}

type Day struct {
	Enter  time.Time `json:"enter"`
	Exit   time.Time `json:"exit"`
	Status struct {
		CreatedAt   time.Time   `json:"created_at"`
		DisplayName string      `json:"display_name"`
		ID          int         `json:"id"`
		IsPaid      bool        `json:"is_paid"`
		IsShow      bool        `json:"is_show"`
		Name        string      `json:"name"`
		NumberTask  interface{} `json:"number_task"`
		UpdatedAt   time.Time   `json:"updated_at"`
	} `json:"status"`
	RawStatus struct {
		CreatedAt   time.Time   `json:"created_at"`
		DisplayName string      `json:"display_name"`
		ID          int         `json:"id"`
		IsPaid      bool        `json:"is_paid"`
		IsShow      bool        `json:"is_show"`
		Name        string      `json:"name"`
		NumberTask  interface{} `json:"number_task"`
		UpdatedAt   time.Time   `json:"updated_at"`
	} `json:"raw_status"`
	Day         string `json:"day"`
	IsReason    bool   `json:"is_reason"`
	WorkingTime string `json:"working_time"`
	WorkingHour int    `json:"working_hour"`
	WorkingMin  int    `json:"working_min"`
	TimeForWeek string `json:"time_for_week"`
}

type Info struct {
	Days       []*Day
	Total      string `json:"total"`
	WorkingDay int    `json:"working_day"`
	NeedTotal  int    `json:"need_total"`
	Person     struct {
		ADDOMAINID        int         `json:"AD_DOMAIN_ID"`
		ADENABLED         bool        `json:"AD_ENABLED"`
		ADSYNCPENDING     int         `json:"AD_SYNC_PENDING"`
		ADUSERDN          interface{} `json:"AD_USER_DN"`
		APBON             bool        `json:"APB_ON"`
		BADGE             int         `json:"BADGE"`
		BOOLPARAM1        bool        `json:"BOOLPARAM1"`
		BOOLPARAM2        bool        `json:"BOOLPARAM2"`
		BOOLPARAM3        bool        `json:"BOOLPARAM3"`
		BOOLPARAM4        bool        `json:"BOOLPARAM4"`
		CODEKEY           string      `json:"CODEKEY"`
		CODEKEYTIME       time.Time   `json:"CODEKEYTIME"`
		CODEKEYDISPFORMAT string      `json:"CODEKEY_DISP_FORMAT"`
		CREATEDTIME       time.Time   `json:"CREATEDTIME"`
		DESCRIPTION       string      `json:"DESCRIPTION"`
		EXPTIME           interface{} `json:"EXPTIME"`
		EXTID             interface{} `json:"EXTID"`
		FIREDTIME         interface{} `json:"FIREDTIME"`
		ID                int         `json:"ID"`
		LASTPASSAP        int         `json:"LASTPASS_AP"`
		LOCATIONACT       time.Time   `json:"LOCATIONACT"`
		LOCATIONZONE      int         `json:"LOCATIONZONE"`
		NAME              string      `json:"NAME"`
		PARENTID          int         `json:"PARENT_ID"`
		POS               string      `json:"POS"`
		SIDEPARAM0        interface{} `json:"SIDEPARAM0"`
		SIDEPARAM1        interface{} `json:"SIDEPARAM1"`
		SIDEPARAM2        interface{} `json:"SIDEPARAM2"`
		SIDEPARAM3        interface{} `json:"SIDEPARAM3"`
		SIDEPARAM4        interface{} `json:"SIDEPARAM4"`
		SIDEPARAM5        interface{} `json:"SIDEPARAM5"`
		SMSENDDATE        interface{} `json:"SMS_ENDDATE"`
		SMSLASTAP         int         `json:"SMS_LAST_AP"`
		SMSLASTDIR        int         `json:"SMS_LAST_DIR"`
		SMSLASTSENT       bool        `json:"SMS_LAST_SENT"`
		SMSLASTTIME       interface{} `json:"SMS_LAST_TIME"`
		SMSPASSENABLED    bool        `json:"SMS_PASS_ENABLED"`
		SMSPASSTEXT       interface{} `json:"SMS_PASS_TEXT"`
		SMSPAYENABLED     bool        `json:"SMS_PAY_ENABLED"`
		SMSPAYTEXT        interface{} `json:"SMS_PAY_TEXT"`
		SMSSTARTDATE      interface{} `json:"SMS_STARTDATE"`
		SMSTARGETNUMBER   interface{} `json:"SMS_TARGETNUMBER"`
		STATUS            string      `json:"STATUS"`
		TABID             string      `json:"TABID"`
		TYPE              string      `json:"TYPE"`
	} `json:"person"`
	Workday           string      `json:"workday"`
	TotalAll          string      `json:"total_all"`
	Workdaysec        int         `json:"workdaysec"`
	Weekendsec        interface{} `json:"weekendsec"`
	Weekend           string      `json:"weekend"`
	Truancysec        interface{} `json:"truancysec"`
	Truancy           string      `json:"truancy"`
	Illnesssec        interface{} `json:"illnesssec"`
	Illness           string      `json:"illness"`
	Businesstripsec   interface{} `json:"businesstripsec"`
	Businesstrip      string      `json:"businesstrip"`
	Vacationsec       interface{} `json:"vacationsec"`
	Vacation          string      `json:"vacation"`
	VacationUnpaidsec interface{} `json:"vacation_unpaidsec"`
	VacationUnpaid    string      `json:"vacation_unpaid"`
	OfflineWorksec    interface{} `json:"offline_worksec"`
	OfflineWork       string      `json:"offline_work"`
	Futuresec         interface{} `json:"futuresec"`
	Future            string      `json:"future"`
	Exchangesec       interface{} `json:"exchangesec"`
	Exchange          string      `json:"exchange"`
	AvgWorkingTime    string      `json:"avg_working_time"`
}

func (info *Info) GetAverage() []int {
	value := info.AvgWorkingTime
	avgS := strings.Split(strings.TrimSpace(value), "h ")
	avgS[1] = avgS[1][:2]
	avg := make([]int, 2)
	avg[0], _ = strconv.Atoi(avgS[0])
	avg[1], _ = strconv.Atoi(avgS[1])
	return avg
}

func (info *Info) GetHomeTime() (since time.Duration, exit time.Time, sinceIdeal time.Duration, exitIdeal time.Time, ifNow time.Time) {
	now := time.Now()
	today := info.Days[now.Day()]
	exit = today.Enter.Add(8 * time.Hour)
	since = exit.Sub(now)

	workDays := 0
	n := time.Time{}
	for _, d := range info.Days {
		if d == nil {
			continue
		}
		if d.Enter != n {
			workDays++
		}
	}
	if workDays == 0 {
		workDays = 1
	}

	avg := info.GetAverage()
	norm := (8*60 + 5) * workDays
	avgT := avg[0]*60 + avg[1]
	current := avgT * (workDays - 1)
	offset := norm - current
	ideal := []int{offset / 60, offset % 60}
	exitIdeal = time.Date(now.Year(), now.Month(), now.Day(), today.Enter.Hour()+ideal[0], today.Enter.Minute()+ideal[1], 0, 0, time.Local)
	sinceIdeal = exitIdeal.Sub(now)
	currAvg := (avgT*workDays - int(since.Minutes())) / workDays
	ifNow = time.Date(now.Year(), now.Month(), now.Day(), currAvg/60, currAvg%60, 0, 0, time.Local)
	return since, exit, sinceIdeal, exitIdeal, ifNow
}
