package raft

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"time"

	pb "github.com/misachi/raft/protos/requestvote"
)

const (
	Follower  = "Follower"
	Candidate = "Candidate"
	Leader    = "Leader"
)

var (
	CurrentNode *Node
	nodeDetail  = fmt.Sprintf("%s/.config/node-detail.json", os.Getenv("HOME"))
)

func getArrayElement(slice []string, element string) (int, error) {
	for idx, name := range slice {
		if name == element {
			return idx, nil
		}
	}
	return -1, fmt.Errorf("element not found")
}

func removeArrayElement(slice []string, idx int) []string {
	slice = append(slice[:idx], slice[idx:]...)
	return slice
}

type Node struct {
	CurrentTerm  int64    `json:"current_term,omitempty"` /* Store the term we're in atm */
	VotedFor     string   `json:"voted_for,omitempty"`    /* CandidateId that received vote */
	State        string   `json:"state,omitempty"`        /* Current state of the node */
	Name         string   `json:"name,omitempty"`         /* Node name */
	CommitIndex  int      `json:"commit_index,omitempty"` /* Index of highest log entry known to be committed */
	LastApplied  int      `json:"last_applied,omitempty"` /* Index of highest log entry applied to state machine */
	Nodes        []string `json:"nodes,omitempty"`        /* Nodes in cluster */
	LogEntryPath string   `json:"log_entry,omitempty"`    /* Command Entries */
	TimeOut      int      `json:"time_out,omitempty"`     /*  */
}

func NewNode(serverName string, clusterNodes []string) *Node {
	srvIndex, err := getArrayElement(clusterNodes, serverName)
	if err != nil {
		log.Fatal(err)
	}
	clusterNodes = removeArrayElement(clusterNodes, srvIndex)
	return &Node{
		State: Follower,
		Name:  serverName,
		Nodes: clusterNodes,
	}
}

func ReadNodeFile(buf []byte) *Node {
	store := NewDiskStore(nodeDetail)
	fileObj, err := store.CreateFile(0644, os.O_RDONLY)
	if err != nil {
		log.Fatal(err)
	}
	defer fileObj.Close()
	total := 0
	for {
		nRead, err := store.ReadFile(buf, fileObj)
		if err != nil {
			break
		}
		total += nRead
	}
	node := Node{}
	if err := json.Unmarshal(buf[:total], &node); err != nil {
		log.Fatal(err)
	}
	return &node
}

func (n *Node) GetNodeFromFile(buf []byte) *Node {
	return ReadNodeFile(buf)
}

func (n *Node) PersistToDisk(perm fs.FileMode, flag int) error {
	buf, err := json.Marshal(n)
	if err != nil {
		return fmt.Errorf("unable to serialize to Json: %v", err)
	}
	store := NewDiskStore(nodeDetail)
	fileObj, err := store.CreateFile(perm, flag)
	if err != nil {
		log.Fatal(err)
	}
	defer fileObj.Close()
	if err := store.WriteFile(buf, fileObj); err != nil {
		return fmt.Errorf("failed %v", err)
	}
	return nil
}

func (n *Node) String() string {
	return fmt.Sprintf("Name: %s State: %s Term: %d", n.Name, n.State, n.CurrentTerm)
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
	voteResponseChan := make(chan *pb.RequestVoteResponse)
	n.CurrentTerm++ /* Increment term by 1 */
	totalVote++     /* Vote for myself */
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, srv_node := range n.Nodes {
		go func(ctx context.Context, srv string, currentTerm int64, name string) {
			voteResponseChan <- requestVoteMsg.Send(srv, currentTerm, n.Name, 0, 0)
		}(ctx, srv_node, n.CurrentTerm, n.Name)
	}

	for range n.Nodes {
		select {
		case <-ctx.Done():
			fmt.Println(ctx.Err())
			return
		case vote_response := <-voteResponseChan:
			new_term := vote_response.GetTerm()
			if new_term >= n.CurrentTerm && !vote_response.GetVoteGranted() {
				n.CurrentTerm = new_term
				n.State = Follower
				goto done
			}
			totalVote++
			if totalVote >= n.avgNodeCount() {
				n.State = Leader
				goto done
			}
		}
	}
done:
	if err := n.PersistToDisk(0644, os.O_CREATE|os.O_WRONLY); err != nil {
		log.Fatal(err)
	}
}

func (n *Node) VoteForClient(client_name string, term int64, lastLogIdx int64, lastLogTerm int64) (int64, bool) {
	if term > n.CurrentTerm {
		n.State = Follower
		n.CurrentTerm = term
		n.VotedFor = client_name
		return term, true
	}
	return n.CurrentTerm, false
}
