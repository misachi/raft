package main

import (
	"fmt"
	"log"

	"github.com/misachi/raft"
)

func main() {
	// serverName := fmt.Sprintf("localhost:%d", 4000)
	// nodes := []string{serverName}
	if raft.CurrentNode == nil {
		log.Fatal("The server is not running. Ensure the server is started.")
	}
	for {
		raft.CurrentNode.SendRequestVote()
	}
}
