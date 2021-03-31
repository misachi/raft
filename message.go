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
	Receive()
}

type RequestVoteMsg struct{}

func (r *RequestVoteMsg) getConnection(srv_name string) (context.Context, *grpc.ClientConn) {
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
	// ctx, conn := r.getConnection((server_name))
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

func (r *RequestVoteMsg) Receive() {}
