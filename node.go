package raft

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"sync"
	"time"
	"unsafe"

	pb "github.com/misachi/raft/protos/requestvote"
)

const (
	Follower  = "Follower"
	Candidate = "Candidate"
	Leader    = "Leader"
)

var (
	CurrentNode  *Node
	NodeDetail   = fmt.Sprintf("%s/.config/node-detail.json", os.Getenv("HOME"))
	EntryLogFile = fmt.Sprintf("%s/.config/entry_logs.json", os.Getenv("HOME"))
)

type Node struct {
	mu          sync.Mutex
	CurrentTerm int64    `json:"current_term,omitempty"` /* Store the term we're in atm */
	VotedFor    string   `json:"voted_for,omitempty"`    /* CandidateId that received vote */
	State       string   `json:"state,omitempty"`        /* Current state of the node */
	Name        string   `json:"name,omitempty"`         /* Node name */
	CommitIndex int      `json:"commit_index,omitempty"` /* Index of highest log entry known to be committed */
	LastApplied int      `json:"last_applied,omitempty"` /* Index of highest log entry applied to state machine */
	Nodes       []string `json:"nodes,omitempty"`        /* Nodes in cluster */
	LogEntry    []Entry  `json:"log_entry,omitempty"`    /* Command Entries */
}

func NewNode(serverName string, clusterNodes []string) *Node {
	srvIndex, err := getArrayElement(clusterNodes, serverName)
	if err != nil {
		log.Fatal(err)
	}
	clusterNodes = removeArrayElement(clusterNodes, srvIndex)
	node := Node{
		State: Follower,
		Name:  serverName,
		Nodes: clusterNodes,
	}
	entry := []Entry{}
	if err := ReadEntryLog(entry); err != nil {
		log.Fatalf("Could not read entry log: %v", err)
	}
	node.LogEntry = entry
	return &node
}

func (n *Node) ReadNodeFromFile(buf []byte, flag int) *Node {
	n.mu.Lock()
	defer n.mu.Unlock()
	store := NewDiskStore(NodeDetail)
	fileObj, err := store.CreateFile(0644, flag)
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

	if total > 0 {
		var node Node
		if err := json.Unmarshal(buf[:total], &node); err != nil {
			log.Fatal(err)
		}
		return &node
	}
	return nil
}

func (n *Node) PersistToDisk(perm fs.FileMode, flag int) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	buf, err := json.Marshal(n)
	if err != nil {
		return fmt.Errorf("unable to serialize to Json: %v", err)
	}
	store := NewDiskStore(NodeDetail)
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

func avgNodeCount(nList []string) int {
	return (len(nList) + 1) / 2
}

func (n *Node) TruncNodeFile(size int64) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if err := os.Truncate(NodeDetail, size); err != nil {
		log.Fatalf("unable to resize file: %v", err)
	}
}

func (n *Node) RefreshFromDisk() *Node {
	return NodeFromDisk()
}

func getRequestVoteResponse(ctx context.Context, n *Node, voteResponseChan chan *pb.RequestVoteResponse, nodeName chan string) {
	requestVoteMsg := RequestVoteMsg{}
	for _, srv_node := range n.Nodes {
		select {
		case <-ctx.Done():
			log.Println(ctx.Err())
			return
		default:
			go func(ctx context.Context, srv string, currentTerm int64, name string) {
				select {
				case <-ctx.Done():
					log.Println(ctx.Err())
					return
				case voteResponseChan <- requestVoteMsg.Send(srv, currentTerm, n.Name, 0, 0):
					nodeName <- srv
				}
			}(ctx, srv_node, n.CurrentTerm, n.Name)
		}
	}
}

func (n *Node) SendRequestVote() {
	if n.State != Candidate {
		log.Printf("%s Only Candidate nodes can send RequestVote\n", n.Name)
		return
	}
	totalVote, nodeCount := 0, avgNodeCount(n.Nodes)
	voteResponseChan := make(chan *pb.RequestVoteResponse)
	nodeName := make(chan string)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	n.CurrentTerm++ /* Increment term by 1 */
	totalVote++     /* Vote for myself */

	go getRequestVoteResponse(ctx, n, voteResponseChan, nodeName)

	for vote_response := range voteResponseChan {
		granted := vote_response.GetVoteGranted()
		new_term := vote_response.GetTerm()
		log.Printf("%s voted %v in term %d\n", <-nodeName, granted, new_term)
		if new_term >= n.CurrentTerm && !granted {
			n.CurrentTerm = new_term
			n.State = Follower
			cancel() // Cleanup if a leader exists already
			break
		}
		if granted {
			totalVote++
		}
		if totalVote >= nodeCount {
			log.Printf("%s won Election", n.Name)
			n.State = Leader
			cancel() // Cleanup if we have majority votes
			break
		}
	}

	n.TruncNodeFile(int64(unsafe.Sizeof(n)))
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
