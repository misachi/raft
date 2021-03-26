package main

import (
	"context"
	"fmt"
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

func (r *RequestVoteMsg) Send(node *Node, term int, cId string, lastLogIdx int, lastLogTerm int) {
	for _, node := range node.Nodes {
		func ()  {
			conn, err := grpc.Dial(node, grpc.WithInsecure(), grpc.WithBlock())
			if err != nil {
				log.Fatal("Could not connect to remote server: %v", err)
			}
			defer conn.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			client := pb.NewRequestVoteClient(conn)
			fmt.Println(client.GetVote(ctx, &pb.RequestVoteDetail{Term: term, LastLogIndex: lastLogIdx, LastLogTerm: lastLogTerm, CandidateId: cId}))
		}()
	}
}

func (r *RequestVoteMsg) Receive() {}
