package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/misachi/raft"
	pb "github.com/misachi/raft/protos/requestvote"
	"google.golang.org/grpc"
)

type RequestVoteServer struct {
	pb.UnimplementedRequestVoteServer
}

func (r *RequestVoteServer) GetVote(ctx context.Context, detail *pb.RequestVoteDetail) (*pb.RequestVoteResponse, error) {
	term, voted := raft.CurrentNode.VoteForClient(detail.CandidateId, detail.Term, detail.LastLogIndex, detail.LastLogTerm)
	return &pb.RequestVoteResponse{Term: term, VoteGranted: voted}, nil
}

func main() {
	serverName := fmt.Sprintf("localhost:%d", 4000)
	nodes := []string{serverName}
	lis, err := net.Listen("tcp", serverName)
	if err != nil {
		log.Fatalf("Error starting up server: %v", err)
	}
	if raft.CurrentNode == nil {
		raft.CurrentNode = raft.NewNode(serverName, nodes)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterRequestVoteServer(grpcServer, &RequestVoteServer{})
	grpcServer.Serve(lis)
}
