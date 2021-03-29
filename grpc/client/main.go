package main

import (
	"fmt"

	"github.com/misachi/raft"
)

func main() {
	/* Client test code */
	serverName := fmt.Sprintf("localhost:%d", 4000)
	nodes := []string{serverName}
	if raft.CurrentNode == nil {
		raft.CurrentNode = raft.NewNode(serverName, nodes)
		// log.Fatal("The server is not running. Ensure the server is started.")
	}
	raft.CurrentNode.SendRequestVote()
}
