package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"

	pb "github.com/raft/protos/vote"
)

type Message interface {
	Send()
	Receive()
}

type RequestVoteMsg struct{}

func (r *RequestVoteMsg) getConnection(srv_name string) (context.Context, *grpc.ClientConn) {
	conn, err := grpc.Dial(srv_name, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatal("Could not connect to remote server: %v", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return ctx, conn
}

func (r *RequestVoteMsg) Send(server_name string, term int, cId string, lastLogIdx int, lastLogTerm int) *pb.RequestVoteResponse {
	ctx, conn := r.getConnection((server_name))

	client := pb.NewRequestVoteClient(conn)
	return client.GetVote(ctx,
		&pb.RequestVoteDetail{
			Term:         term,
			LastLogIndex: lastLogIdx,
			LastLogTerm:  lastLogTerm,
			CandidateId:  cId})
}

func (r *RequestVoteMsg) Receive() {}
