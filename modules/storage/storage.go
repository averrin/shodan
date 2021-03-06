package storage

import (
	"fmt"
	"log"
	"time"

	"github.com/fatih/color"
	r "gopkg.in/dancannon/gorethink.v2"
)

type Storage map[string]string

var conn *r.Session

func Connect(creds map[string]string) *Storage {
	stor := Storage{}
	for k, v := range creds {
		stor[k] = v
	}
	stor.NewDB()
	return &stor
}

func (stor *Storage) GetSession() *r.Session {
	return conn
}

func (stor *Storage) Exec(q r.Term, result interface{}) interface{} {
	res, err := q.Run(conn)
	defer res.Close()
	if err != nil {
		log.Println(err)
	}
	res.All(result)
	return result
}

func (stor *Storage) NewDB() (err error) {
	creds := *stor
	conn, err = r.Connect(r.ConnectOpts{
		Address:  creds["host"],
		Database: creds["database"],
		Username: creds["user"],
		Password: creds["password"],
	})
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

type Change struct {
	NewVal Event `gorethink:"new_val"`
}

type NoteChange struct {
	NewVal Note `gorethink:"new_val"`
}

func (stor *Storage) GetEventsStream() chan Event {
	creds := *stor
	c := make(chan Event)
	res, err := r.DB(creds["database"]).Table("events").Changes(r.ChangesOpts{
	// IncludeInitial: true,
	}).Run(conn)
	if err != nil {
		log.Println(err)
	}
	go func() {
		ch := Change{}
		for res.Next(&ch) {
			c <- ch.NewVal
		}
		res.Close()
	}()
	return c
}

func (stor *Storage) ReportEvent(event string, note string) {
	creds := *stor
	e := Event{
		Event:     event,
		Note:      note,
		Timestamp: time.Now(),
	}
	log.Println(e)
	_, err := r.DB(creds["database"]).Table("events").Insert(e).Run(conn)
	if err != nil {
		log.Println(err)
	}
}

type Event struct {
	Event     string      `json:"Event"`
	Note      string      `json:"Note"`
	Tag       string      `json:"Tag"`
	Payload   interface{} `json:"Payload"`
	Timestamp time.Time   `json:"Timestamp"`
}

func (e Event) String() string {
	yellow := color.New(color.FgYellow).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	if e.Note == "" {
		return fmt.Sprintf("%s %s", yellow("E:"), e.Event)
	}
	return fmt.Sprintf("%s %s [%s]", yellow("E:"), e.Event, green(e.Note))
}
