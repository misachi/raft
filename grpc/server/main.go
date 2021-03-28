package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/misachi/raft"
	pb "github.com/misachi/raft/protos/requestvote"
	"google.golang.org/grpc"
)

type RequestVoteServer struct {
	pb.UnimplementedRequestVoteServer
}

func (r *RequestVoteServer) GetVote(ctx context.Context, detail *pb.RequestVoteDetail) (*pb.RequestVoteResponse, error) {
	node := raft.ReadNodeFile()
	fmt.Println(node)
	term, voted := node.VoteForClient(detail.CandidateId, detail.Term, detail.LastLogIndex, detail.LastLogTerm)
	return &pb.RequestVoteResponse{Term: term, VoteGranted: voted}, nil
}

func main() {
	serverName := fmt.Sprintf("localhost:%d", 4000)
	nodes := []string{serverName}
	lis, err := net.Listen("tcp", serverName)
	if err != nil {
		log.Fatalf("Error starting up server: %v", err)
	}
	node := raft.NewNode(serverName, nodes)
	fmt.Println(node)
	if err := node.PersistToDisk(0644, os.O_CREATE|os.O_WRONLY); err != nil {
		log.Fatal(err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterRequestVoteServer(grpcServer, &RequestVoteServer{})
	grpcServer.Serve(lis)
}
