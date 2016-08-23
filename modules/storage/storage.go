package storage

import (
	"fmt"
	"log"
	"strconv"
	"time"

	couchdb "github.com/rhinoman/couchdb-go"
	uuid "github.com/satori/go.uuid"
)

type Storage map[string]string

var notes *couchdb.Database
var events *couchdb.Database

func Connect(creds map[string]string) *Storage {
	stor := Storage{}
	for k, v := range creds {
		stor[k] = v
	}
	stor.NewDB()
	return &stor
}

func (stor *Storage) NewDB() {
	creds := *stor
	port, _ := strconv.Atoi(creds["port"])
	timeout := time.Duration(3 * time.Second)
	conn, err := couchdb.NewConnection(creds["host"], port, timeout)
	if err != nil {
		log.Println(err)
	}
	auth := couchdb.BasicAuth{Username: creds["user"], Password: creds["password"]}
	events = conn.SelectDB("events", &auth)
	notes = conn.SelectDB("notes", &auth)
	if notes != nil {
		log.Println("Storage connected")
	}
	ddoc := DesignDocument{
		Language: "javascript",
		Views:    getNotesViews(),
	}
	notes.SaveDesignDoc("notes", ddoc, "")
}

func (stor *Storage) ClearNotes() {
	results := ViewResult{}
	err := notes.GetView("notes", "list", &results, nil)
	if err != nil {
		log.Println(err)
	}
	for _, n := range results.Rows {
		log.Println(notes.Delete(n.Key[0], n.Key[0]))
	}
}

func (stor *Storage) GetNotes() []Note {
	results := ViewResult{}
	err := notes.GetView("notes", "list", &results, nil)
	if err != nil {
		log.Println(err)
	}
	r := []Note{}
	for _, n := range results.Rows {
		r = append(r, n.Value)
	}
	return r
}

func (stor *Storage) SaveNote(text string) {
	note := Note{
		Text:      text,
		Timestamp: time.Now(),
	}
	_, err := notes.Save(note, fmt.Sprintf("%s", uuid.NewV4()), "")
	if err != nil {
		log.Println(err)
	}
}

func (stor *Storage) ReportEvent(event string, note string) {
	e := Event{
		Event:     event,
		Note:      note,
		Timestamp: time.Now(),
	}
	log.Println(e)
	_, err := events.Save(e, fmt.Sprintf("%s", uuid.NewV4()), "")
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

func getNotesViews() map[string]View {
	return map[string]View{
		"list": {
			Map: `function(doc){
				emit([doc._id, doc._rev], doc);
			}`,
		},
	}
}

type DesignDocument struct {
	Language string          `json:"language"`
	Views    map[string]View `json:"views"`
}

type View struct {
	Map    string `json:"map"`
	Reduce string `json:"reduce,omitempty"`
}

type ViewResult struct {
	TotalRows int `json:"total_rows"`
	Offset    int `json:"offset"`
	Rows      []struct {
		ID    string   `json:"id"`
		Key   []string `json:"key"`
		Value Note     `json:"value"`
	} `json:"rows"`
}
