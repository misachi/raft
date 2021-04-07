package raft

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
		Id: id,
		Term: term,
		Command: command,
	}
}