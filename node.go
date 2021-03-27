package main

import (
	"fmt"
	"sync"
)

const (
	Follower = iota
	Candidate
	Leader
)

type Node struct {
	mut         *sync.Mutex
	CurrentTerm int    /* Store the term we're in atm */
	VotedFor    string /* CandidateId that received vote */
	State       int    /* Current state of the node */
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

func (n *Node) avgNodeCount() int {
	return len(n.Nodes) / 2
}

func (n *Node) setName(name string) {
	if name == "" {
		panic("Node name cannot be empty")
	}
	n.Name = name
}

func (n *Node) SendRequestVote() {
	requestVoteMsg := RequestVoteMsg{}
	totalVote := 0

	for _, srv_node := range n.Nodes {
		vote_response := requestVoteMsg.Send(srv_node, n.CurrentTerm, n.Name, 0, 0)

		fmt.Println(vote_response)
		new_term := vote_response.GetTerm()
		if new_term >= n.CurrentTerm && !vote_response.GetVoteGranted() {
			n.CurrentTerm = new_term
			n.State = Follower
			break
		}
		totalVote++
		if totalVote >= n.avgNodeCount() {
			n.State = Leader
			return
		}
	}
}

func (n *Node) voteForClient(client_name string, term int, lastLogIdx int, lastLogTerm int) (int, bool) {
	if term > n.CurrentTerm {
		n.State = Follower
		n.CurrentTerm = term
		n.VotedFor = client_name
		return term, true
	}
	return n.CurrentTerm, false
}
