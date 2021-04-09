package raft

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"

	pb "github.com/misachi/raft/protos/requestvote"
)

type Message interface {
	Send()
}

type RequestVoteMsg struct{}

type AppendRequestMsg struct{}

func getConnection(srv_name string) (context.Context, *grpc.ClientConn) {
	conn, err := grpc.Dial(srv_name, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Could not connect to remote server: %v", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return ctx, conn
}

func (r *RequestVoteMsg) Send(server_name string, term int64, cId string, lastLogIdx int64, lastLogTerm int64) *pb.RequestVoteResponse {
	conn, err := grpc.Dial(server_name, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Could not connect to remote server: %v", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := pb.NewRequestVoteClient(conn)
	response, err := client.GetVote(ctx,
		&pb.RequestVoteDetail{
			Term:         term,
			LastLogIndex: lastLogIdx,
			LastLogTerm:  lastLogTerm,
			CandidateId:  cId})
	if err != nil {
		log.Fatalf("Error when sending requestVote: %v", err)
	}
	return response
}

func (r *AppendRequestMsg) Send(server_name string, req *pb.AppendEntryRequestDetail) *pb.AppendEntryResponse {
	conn, err := grpc.Dial(server_name, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Could not connect to remote server: %v", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := pb.NewAppendEntryClient(conn)
	response, err := client.AddEntry(ctx, req)
	if err != nil {
		log.Fatalf("Error when sending entries: %v", err)
	}
	return response
}
