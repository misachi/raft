package main

import (
	"log"
	"os"
	"unsafe"

	"github.com/misachi/raft"
)

func main() {
	/* Client test code */
	buf := make([]byte, int(unsafe.Sizeof(raft.Node{})))
	node := raft.ReadNodeFile(buf, os.O_RDONLY)
	if node == nil {
		log.Fatal("The server is not running. Ensure the server is started.")
	}

	node.SendRequestVote()
}
