package main

import (
	"fmt"
	"testing"
)

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
	container("redis", "run", "")
}
