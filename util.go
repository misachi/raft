package raft

import (
	"errors"
	"os"
	"unsafe"
)

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
