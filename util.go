package raft

import (
	"log"
	"os"
	"unsafe"
)

func GetBufferSize() int {
	var bufSize = int(unsafe.Sizeof(Node{}))
	fInfo, err := os.Stat(NodeDetail)
	if err != nil {
		log.Printf("File Stat error: %v\n", err)
	} else {
		bufSize = int(fInfo.Size())
	}
	return bufSize
}
