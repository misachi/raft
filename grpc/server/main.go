package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/misachi/raft"
	pb "github.com/misachi/raft/protos/requestvote"
	"google.golang.org/grpc"
)

const (
	port       = 4009
	minTimeout = 250
	maxTimeout = 300
)

var (
	requestVoteCh = make(chan bool, 1)
	appendEntryCh = make(chan *pb.AppendEntryRequestDetail, 1)
)

type RequestVoteServer struct {
	pb.UnimplementedRequestVoteServer
}

func (r *RequestVoteServer) GetVote(ctx context.Context, detail *pb.RequestVoteDetail) (*pb.RequestVoteResponse, error) {
	node := raft.NodeFromDisk()
	log.Printf("Received requestVote from %s", detail.GetCandidateId())
	term, voted := node.VoteForClient(detail.GetCandidateId(), detail.GetTerm(), detail.GetLastLogIndex(), detail.GetLastLogTerm())
	if err := node.PersistToDisk(0644, os.O_CREATE|os.O_WRONLY); err != nil {
		log.Fatal(err)
	}
	requestVoteCh <- true
	return &pb.RequestVoteResponse{Term: term, VoteGranted: voted}, nil
}

type AppendEntryServer struct {
	pb.UnimplementedAppendEntryServer
}

func (r *AppendEntryServer) AddEntry(ctx context.Context, req *pb.AppendEntryRequestDetail) (*pb.AppendEntryResponse, error) {
	buf := make([]byte, raft.GetBufferSize())
	node := new(raft.Node)
	node = node.ReadNodeFromFile(buf, os.O_RDONLY)
	log.Printf("Received appendEntry from %s", req.GetLeaderId())
	appendLog := raft.NewAppendLog(node)
	response, err := appendLog.AppendEntryLog(req)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	appendEntryCh <- req
	return response, nil
}

func main() {
	/* The addresses are for tests only */
	serverName := fmt.Sprintf("172.24.0.2:%d", 4002)
	nodes := []string{"172.24.0.3:4000"}

	log.Println("Starting server...")
	log.Printf("Listening on port %d\n", port)
	lis, err := net.Listen("tcp", serverName)
	if err != nil {
		log.Fatalf("Error starting up server: %v", err)
	}

	var bufSize = raft.GetBufferSize()

	buf := make([]byte, bufSize)
	node := new(raft.Node)
	log.Println("Loading config file...")
	node = node.ReadNodeFromFile(buf, os.O_CREATE|os.O_RDONLY)
	if node == nil {
		node = raft.NewNode(serverName, nodes)
	}

	if err := node.PersistToDisk(0644, os.O_WRONLY); err != nil {
		log.Fatal(err)
	}

	store := raft.NewDiskStore(raft.EntryLogFile)
	file, err := store.CreateFile(0644, os.O_CREATE)
	if err != nil {
		log.Fatalf("Unable to open/create entry-log file: %v", err)
	}
	file.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func(c context.Context, n *raft.Node) {
		backoff := raft.NewExponentialInterval(minTimeout, maxTimeout, 0)
		for {
			select {
			case <-c.Done():
				log.Printf("Server context cancelled: %v", c.Err())
				return
			case <-requestVoteCh:
				log.Printf("%s Received RequestVote", n.Name)
			case req := <-appendEntryCh:
				log.Printf("%s Received appendEntry request", n.Name)
				if req.Term >= n.CurrentTerm {
					n.CurrentTerm = req.Term
					n.State = raft.Follower
				} else {
					if n.State != raft.Leader {
						n.State = raft.Candidate
					}
				}
			case <-time.After(backoff.Next() * time.Millisecond):
				/* If we time out waiting for AppendEntry, we change to Candidate and send RequestVote to everyone */
				if n.State != raft.Leader {
					n.State = raft.Candidate
					n.SendRequestVote()
				}
			}
		}
	}(ctx, node)

	grpcServer := grpc.NewServer()
	pb.RegisterRequestVoteServer(grpcServer, &RequestVoteServer{})
	grpcServer.Serve(lis)
}
