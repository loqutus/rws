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

func TestStorageRemove(t *testing.T) {
	err := storageRemove("test")
	if err != nil {
		fmt.Println(err)
		t.Errorf("error")
	}
}

func TestRunContainer(t *testing.T) {
	id = container("container_run", "redis")
}

func TestListContainer(t *testing.T) {
	_ = container("container_list", "")
}

func TestStopContainer(t *testing.T) {
	_ = container("container_stop", id)
}

func TestRemoveContainer(t *testing.T) {
	_ = container("container_remove", id)
}

func TestAddHost(t *testing.T) {
	_ = hosts("host_add", "localhost")
}

func TestRemoveHost(t *testing.T) {
	_ = hosts("host_remove", "localhost")
}

func TestListHosts(t *testing.T) {
	_ = hosts("host_list", "")
}
