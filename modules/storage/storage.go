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

type Note struct {
	ID        string    `json:"_id"`
	Rev       string    `json:"_rev"`
	Text      string    `json:"Text"`
	Timestamp time.Time `json:"Timestamp"`
}

func getNotesViews() map[string]View {
	return map[string]View{
		"list": {
			Map: `function(doc){
				emit(doc.Timestamp, doc);
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
		ID    string    `json:"id"`
		Key   time.Time `json:"key"`
		Value Note      `json:"value"`
	} `json:"rows"`
}
