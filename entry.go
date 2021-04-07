package raft

import (
	"encoding/json"
	"log"
	"os"
)

const minId = 1

type Entry struct {
	Id      int    /* Entry index. Increases monotonically */
	Term    int64  /*  Term when entry was received */
	Command []byte /* Command value from client */
}

func NewEntry(id int, term int64, command []byte) *Entry {
	if id <= 0 {
		id = minId
	}
	return &Entry{
		Id:      id,
		Term:    term,
		Command: command,
	}
}

func ReadEntryLog(entries []Entry) error {
	fileSize, err := GetFileSize(EntryLogFile)
	store := DiskStore{}
	if err != nil {
		log.Fatalln(err.Error())
	}
	buf := make([]byte, fileSize)
	file, err := store.CreateFile(0644, os.O_RDONLY)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := store.ReadFile(buf, file); err != nil {
		return err
	}
	if err := json.Unmarshal(buf, &entries); err != nil {
		return err
	}
	return nil
}

func WriteEntryLog(entry Entry) error {
	store := DiskStore{}
	entries := []Entry{}
	if err := ReadEntryLog(entries); err != nil {
		return err
	}
	entries = append([]Entry{entry}, entries...)  /* Write at top of file */
	buf, err := json.Marshal(entries)
	if err != nil {
		return err
	}
	file, err := store.CreateFile(0644, os.O_RDWR)
	if err != nil {
		return err
	}
	defer file.Close()
	file.Truncate(0)
	if err := store.WriteFile(buf, file); err != nil {
		return err
	}
	return nil
}
