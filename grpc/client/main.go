package main

import (
	"log"
	"os"

	"github.com/misachi/raft"
)

func main() {
	/* Client test code */
	buf := make([]byte, raft.GetBufferSize())
	node := new(raft.Node)
	node = node.ReadNodeFromFile(buf, os.O_RDONLY)

	if node == nil {
		log.Fatal("The server is not running. Ensure the server is started.")
	}

	node.SendRequestVote()
}
