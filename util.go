package raft

import (
	"fmt"
	"os"
	"unsafe"
)

func GetBufferSize() int {
	var bufSize = int(unsafe.Sizeof(Node{}))
	fInfo, err := os.Stat(NodeDetail)
	if err != nil {
		fmt.Printf("File Stat error: %v", err)
	} else {
		bufSize = int(fInfo.Size())
	}
	return bufSize
}
