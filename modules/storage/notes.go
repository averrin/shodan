package storage

import (
	"log"
	"time"

	r "gopkg.in/dancannon/gorethink.v2"
)

type Note struct {
	Text      string    `json:"Text"`
	Timestamp time.Time `json:"Timestamp"`
	ID        string    `gorethink:"id" json:"ID"`
}

func (stor *Storage) GetNotesStream() chan Note {
	creds := *stor
	c := make(chan Note)
	res, err := r.DB(creds["database"]).Table("notes").Changes(r.ChangesOpts{
	// IncludeInitial: true,
	}).Run(conn)
	if err != nil {
		log.Println(err)
	}
	go func() {
		ch := NoteChange{}
		for res.Next(&ch) {
			c <- ch.NewVal
		}
		res.Close()
	}()
	return c
}

func (stor *Storage) ClearNotes() {
	creds := *stor
	_, err := r.DB(creds["database"]).Table("notes").Delete(r.DeleteOpts{}).Run(conn)
	if err != nil {
		log.Println(err)
	}
	stor.ReportEvent("notesCleared", "")
}

func (stor *Storage) DeleteNote(id string) {
	creds := *stor
	_, err := r.DB(creds["database"]).Table("notes").Get(id).Delete(r.DeleteOpts{}).Run(conn)
	if err != nil {
		log.Println(err)
	}
	stor.ReportEvent("noteDeleted", "")
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
	log.Println(note)
	c, err := r.DB(creds["database"]).Table("notes").Insert(note).Run(conn)
	log.Println(c, err)
	if err != nil {
		log.Println(err)
	}
	stor.ReportEvent("note", text)
}
