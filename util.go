package raft

import (
	"errors"
	"fmt"
	"os"
	"unsafe"
)

func getArrayElement(slice []string, element string) (int, error) {
	for idx, name := range slice {
		if name == element {
			return idx, nil
		}
	}
	return -1, fmt.Errorf("element not found")
}

func removeArrayElement(slice []string, idx int) []string {
	slice = append(slice[:idx], slice[idx+1:]...)
	return slice
}

func GetFileSize(fileName string) (int64, error) {
	fInfo, err := os.Stat(fileName)
	if err != nil {
		return -1, errors.New(err.Error())
	}
	return fInfo.Size(), nil
}

func GetBufferSize() int {
	var bufSize int
	if fSize, _ := GetFileSize(NodeDetail); fSize > 0 {
		bufSize = int(fSize)
	} else {
		bufSize = int(unsafe.Sizeof(Node{}))
	}
	return bufSize
}

func NodeFromDisk() *Node {
	buf := make([]byte, GetBufferSize())
	node := new(Node)
	return node.ReadNodeFromFile(buf, os.O_RDONLY)
}
