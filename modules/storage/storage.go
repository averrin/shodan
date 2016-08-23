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
	timeout := time.Duration(500 * time.Millisecond)
	conn, err := couchdb.NewConnection(creds["host"], port, timeout)
	if err != nil {
		log.Println(err)
	}
	auth := couchdb.BasicAuth{Username: creds["user"], Password: creds["password"]}
	notes = conn.SelectDB("notes", &auth)
	if notes != nil {
		log.Println("Storage connected")
	}
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
	Text      string
	Timestamp time.Time
}
