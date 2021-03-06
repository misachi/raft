package raft

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
	"unsafe"

	pb "github.com/misachi/raft/protos/requestvote"
)

const (
	minId   = 1
	timeout = 180
)

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
	store := NewDiskStore(EntryLogFile)
	if err != nil {
		log.Fatalln(err.Error())
	}
	buf := make([]byte, fileSize)
	file, err := store.CreateFile(0644, os.O_RDONLY)
	if err != nil {
		return err
	}
	defer file.Close()
	nRead, err := store.ReadFile(buf, file)
	if err != nil && nRead > 0 {
		return err
	}
	if err := json.Unmarshal(buf, &entries); err != nil && nRead > 0 {
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
	entries = append(entries, entry)
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

type AppendLog struct {
	SrvNode   *Node
	LastEntry *Entry
}

func lastEntry(node *Node) *Entry {
	if eSize := len(node.LogEntry); eSize >= 1 {
		return &node.LogEntry[eSize-1]
	}
	return &Entry{}
}

func NewAppendLog(node *Node) *AppendLog {
	return &AppendLog{
		SrvNode:   node,
		LastEntry: lastEntry(node),
	}
}

func (a *AppendLog) buildRequest() *pb.AppendEntryRequestDetail {
	a.LastEntry.Id++
	entry := pb.AppendEntryRequestDetail_Entry{
		Id:      int32(a.LastEntry.Id),
		Term:    a.LastEntry.Term,
		Command: a.LastEntry.Command,
	}
	return &pb.AppendEntryRequestDetail{
		Term:         a.SrvNode.CurrentTerm,
		PrevLogIndex: int64(a.LastEntry.Id),
		PrevLogTerm:  a.LastEntry.Term,
		LeaderId:     a.SrvNode.Name,
		Entry:        []*pb.AppendEntryRequestDetail_Entry{&entry},
	}
}

func (a *AppendLog) SendHeartBeat() {
	request := &pb.AppendEntryRequestDetail{
		Term:         a.SrvNode.CurrentTerm,
		PrevLogIndex: int64(a.LastEntry.Id),
		PrevLogTerm:  a.LastEntry.Term,
		LeaderId:     a.SrvNode.Name,
		Entry:        []*pb.AppendEntryRequestDetail_Entry{},
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Millisecond)
	defer cancel()

	appReqMsg := AppendRequestMsg{}
	for _, srv_name := range a.SrvNode.Nodes {
		go func(c context.Context, srv string, req *pb.AppendEntryRequestDetail) {
			select {
			case <-c.Done():
				log.Printf("%s has been interrupted: %s", srv, c.Err().Error())
				return
			default:
			}
			appReqMsg.Send(srv, req)
		}(ctx, srv_name, request)
	}
}

func (a *AppendLog) SendAppendEntryLog() {
	appendEntryResp := make(chan *pb.AppendEntryResponse)
	totalVote, avgSrvCount := 0, avgNodeCount(a.SrvNode.Nodes)
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Millisecond)
	defer cancel()

	appReqMsg := AppendRequestMsg{}
	request := a.buildRequest()
	for _, srv_name := range a.SrvNode.Nodes {
		go func(c context.Context, srv string, req *pb.AppendEntryRequestDetail) {
			select {
			case <-c.Done():
				log.Printf("%s has been interrupted: %s", srv, c.Err().Error())
				return
			case appendEntryResp <- appReqMsg.Send(srv, req):
			}
		}(ctx, srv_name, request)
	}

	for appendResponse := range appendEntryResp {
		if appendResponse != nil {
			if appendResponse.GetSuccess() {
				totalVote++
				if totalVote >= avgSrvCount {
					a.AddEntryToNode(request)
					break
				}
			} else {
				term := appendResponse.GetTerm()
				if term >= a.SrvNode.CurrentTerm {
					a.SrvNode.CurrentTerm = term
					a.SrvNode.State = Follower
					a.SrvNode.TruncNodeFile(int64(unsafe.Sizeof(a.SrvNode)))
					if err := a.SrvNode.PersistToDisk(0644, os.O_CREATE|os.O_WRONLY); err != nil {
						log.Printf("%s failed to persist server state to disk on AppendEntry request: %v", a.SrvNode.Name, err)
					}
					break
				}
			}
		}
	}
}

func (a *AppendLog) AddEntryToNode(req *pb.AppendEntryRequestDetail) {
	reqEntries := make([]Entry, len(req.GetEntry()))
	for idx, entry := range req.GetEntry() {
		reqEntries[idx] = Entry{
			Id:      int(entry.GetId()),
			Term:    entry.GetTerm(),
			Command: entry.GetCommand(),
		}
	}
	a.SrvNode.LogEntry = append(a.SrvNode.LogEntry, reqEntries...)
}

func (a *AppendLog) AppendEntryLog(req *pb.AppendEntryRequestDetail) (*pb.AppendEntryResponse, error) {
	if len(req.Entry) < 1 {
		return nil, nil
	}

	if a.SrvNode.CurrentTerm >= req.GetTerm() {
		return &pb.AppendEntryResponse{
			Term:    a.SrvNode.CurrentTerm,
			Success: false,
		}, nil
	}
	entry := lastEntry(a.SrvNode)
	if a.SrvNode.CurrentTerm < req.GetTerm() {
		if req.GetPrevLogIndex() < int64(entry.Id) || req.GetPrevLogTerm() < entry.Term {
			return &pb.AppendEntryResponse{
				Term:    req.GetTerm(),
				Success: false,
			}, nil
		}
	}

	a.SrvNode.CurrentTerm = req.GetTerm()
	a.AddEntryToNode(req)

	a.SrvNode.TruncNodeFile(int64(unsafe.Sizeof(a.SrvNode)))
	if err := a.SrvNode.PersistToDisk(0644, os.O_CREATE|os.O_WRONLY); err != nil {
		return nil, fmt.Errorf("%s failed to persist server state to disk on AppendEntry request: %v", a.SrvNode.Name, err)
	}
	return &pb.AppendEntryResponse{
		Term:    req.GetTerm(),
		Success: true,
	}, nil
}
