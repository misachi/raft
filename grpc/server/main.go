package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"unsafe"

	"github.com/misachi/raft"
	pb "github.com/misachi/raft/protos/requestvote"
	"google.golang.org/grpc"
)

type RequestVoteServer struct {
	pb.UnimplementedRequestVoteServer
}

func (r *RequestVoteServer) GetVote(ctx context.Context, detail *pb.RequestVoteDetail) (*pb.RequestVoteResponse, error) {
	buf := make([]byte, int(unsafe.Sizeof(raft.Node{})))
	node := raft.ReadNodeFile(buf, os.O_RDONLY)
	term, voted := node.VoteForClient(detail.CandidateId, detail.Term, detail.LastLogIndex, detail.LastLogTerm)
	if err := node.PersistToDisk(0644, os.O_CREATE|os.O_WRONLY); err != nil {
		log.Fatal(err)
	}
	return &pb.RequestVoteResponse{Term: term, VoteGranted: voted}, nil
}

func main() {
	/* The addresses are for tests only */
	serverName := fmt.Sprintf("172.24.0.2:%d", 4002)
	nodes := []string{"172.24.0.3:4000"}
	lis, err := net.Listen("tcp", serverName)
	if err != nil {
		log.Fatalf("Error starting up server: %v", err)
	}
	buf := make([]byte, int(unsafe.Sizeof(raft.Node{})))
	node := raft.ReadNodeFile(buf, os.O_CREATE|os.O_WRONLY)
	if node == nil {
		node = raft.NewNode(serverName, nodes)
	}
	if err := node.PersistToDisk(0644, os.O_CREATE|os.O_WRONLY); err != nil {
		log.Fatal(err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterRequestVoteServer(grpcServer, &RequestVoteServer{})
	grpcServer.Serve(lis)
}
