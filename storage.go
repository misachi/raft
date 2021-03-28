package raft

import (
	"fmt"
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
		return nil, fmt.Errorf("directory cannot be empty")
	}

	if d.FileName == "" {
		return nil, fmt.Errorf("filename cannot be empty")
	}

	filePath := filepath.Join(d.StoreDir, d.FileName)
	err := os.MkdirAll(filePath, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("unable to create directory: %v", err)
	}
	file, err := os.OpenFile(filePath, flag, perm)
	if err != nil {
		return nil, fmt.Errorf("unable to create file: %v", err)
	}
	return file, nil
}

func (d *DiskStore) ReadFile(buf []byte, file *os.File) error {
	if file == nil {
		newFile, err := d.CreateFile(0644, os.O_WRONLY|os.O_CREATE)
		if file == nil {
			log.Fatalf("Could not create file: %v", err)
		}
		file = newFile
	}
	nRead, err := file.Read(buf)
	if err != nil {
		log.Fatalf("Error when reading file: %v", err)
	}

	if nRead <= 0 {
		return fmt.Errorf("file is empty")
	}
	return nil
}

func (d *DiskStore) WriteFile(buf []byte, file *os.File) error {
	_, err := file.Write(buf)
	if err != nil {
		log.Fatalf("Write error: %v", err)
	}
	return nil
}
