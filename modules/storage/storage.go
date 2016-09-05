package storage

import (
	"fmt"
	"log"
	"time"

	// couchdb "github.com/rhinoman/couchdb-go"

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

func (stor *Storage) ClearNotes() {
	creds := *stor
	_, err := r.DB(creds["database"]).Table("notes").Delete(r.DeleteOpts{}).Run(conn)
	if err != nil {
		log.Println(err)
	}
}

func (stor *Storage) GetNotes() []Note {
	creds := *stor
	notes := []Note{}
	res, err := r.DB(creds["database"]).Table("notes").Run(conn)
	defer res.Close()
	if err != nil {
		log.Println(err)
	}
	res.All(&notes)
	return notes
}

func (stor *Storage) SaveNote(text string) {
	creds := *stor
	note := Note{
		Text:      text,
		Timestamp: time.Now(),
	}
	_, err := r.DB(creds["database"]).Table("notes").Insert(note).Run(conn)
	if err != nil {
		log.Println(err)
	}
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

type Note struct {
	Text      string    `json:"Text"`
	Timestamp time.Time `json:"Timestamp"`
}

type Event struct {
	Event     string    `json:"Event"`
	Note      string    `json:"Note"`
	Timestamp time.Time `json:"Timestamp"`
}

func (e Event) String() string {
	if e.Note == "" {
		return fmt.Sprintf("E: %s", e.Event)
	}
	return fmt.Sprintf("E: %s [%s]", e.Event, e.Note)
}
