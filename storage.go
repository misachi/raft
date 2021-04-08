package raft

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

var (
	FileDir  = ".data"
	FilePath = ""
)

type Store interface {
	CreateFile(perm fs.FileMode, flag int)
	ReadFile(buf []byte, file *os.File)
	WriteFile()
}

type DiskStore struct {
	StoreDir string
	FileName string
}

func NewDiskStore(fileName string) *DiskStore {
	dir, file := filepath.Split(fileName)
	if dir == "" {
		dir = FileDir
	}
	if file == "" {
		log.Fatal("Filename cannot be empty")
	}
	return &DiskStore{StoreDir: dir, FileName: file}
}

func (d *DiskStore) CreateFile(perm fs.FileMode, flag int) (*os.File, error) {
	if d.StoreDir == "" && FileDir == "" {
		return nil, errors.New("directory cannot be empty")
	}

	if d.FileName == "" {
		return nil, errors.New("filename cannot be empty")
	}

	filePath := filepath.Join(d.StoreDir, d.FileName)
	err := os.MkdirAll(d.StoreDir, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("unable to create directory: %v", err)
	}
	file, err := os.OpenFile(filePath, flag, perm)
	if err != nil {
		return nil, fmt.Errorf("unable to create file: %v", err)
	}
	return file, nil
}

func (d *DiskStore) ReadFile(buf []byte, file *os.File) (int, error) {
	if file == nil {
		newFile, err := d.CreateFile(0644, os.O_WRONLY|os.O_CREATE)
		if file == nil {
			return -1, fmt.Errorf("could not create file: %v", err)
		}
		file = newFile
	}
	nRead, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return -1, fmt.Errorf("error reading file: %v", err)
	}

	if nRead <= 0 {
		return nRead, errors.New("file is empty")
	}
	return nRead, nil
}

func (d *DiskStore) WriteFile(buf []byte, file *os.File) error {
	_, err := file.Write(buf)
	if err != nil {
		return fmt.Errorf("write error: %v", err)
	}
	return nil
}
