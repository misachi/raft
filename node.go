package main

import (
	"sync"
)

const (
	Follower = iota
	Candidate
	Leader
)

type Node struct {
	mut       *sync.Mutex
	CurrentTerm int      /* Store the term we're in atm */
	VotedFor    string   /* CandidateId that received vote */
	State       int      /* Current state of the node */
	// Net         *Message /* Communication method */
	Name        string   /* Node name */
	CommitIndex int      /* Index of highest log entry known to be committed */
	LastApplied int      /* Index of highest log entry applied to state machine */
	Nodes       []string /* Nodes in cluster */
	LogEntry    []*Entry /* Command Entries */
	TimeOut     int      /*  */
}

func (n *Node) getServerStateStr() string {
	switch n.State {
	case Follower:
		return "Follower"
	case Candidate:
		return "Candidate"
	case Leader:
		return "Leader"
	}
	return ""
}

func (n *Node) updateTerm() {
	n.CurrentTerm++
}

func (n *Node) getTerm() int {
	return n.CurrentTerm
}
