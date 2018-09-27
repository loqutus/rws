package main

import (
	"fmt"
	"testing"
)

var id string

func TestStorageUpload(t *testing.T) {
	err := storageUpload("test")
	if err != nil {
		fmt.Println(err)
		t.Errorf("error")
	}
}

func TestStorageDownload(t *testing.T) {
	err := storageDownload("test")
	if err != nil {
		fmt.Println(err)
		t.Errorf("error")
	}
}

func TestStorageList(t *testing.T) {
	err := storageList()
	if err != nil {
		fmt.Println(err)
		t.Errorf("error")
	}
}

func TestRunContainer(t *testing.T) {
	id = container("redis", "run", "")
}

func TestListContainer(t *testing.T) {
	fmt.Println("test list")
	l := container("redis", "list", "")
	fmt.Println(l)
}

func TestStopContainer(t *testing.T) {
	_ = container("redis", "stop", id)
}
