package main

import (
	"fmt"
	"testing"
)

func TestStorage(t *testing.T) {
	err := storageUpload("test")
	if err != nil {
		fmt.Println(err)
		t.Errorf("storage upload error")
	}
	err2 := storageDownload("test")
	if err2 != nil {
		fmt.Println(err2)
		t.Errorf("storage download error")
	}
	err3 := storageList()
	if err3 != nil {
		fmt.Println(err3)
		t.Errorf("storage list error")
	}
	err4 := storageRemove("test")
	if err4 != nil {
		fmt.Println(err4)
		t.Errorf("storage remove error")
	}
}

func TestContainer(t *testing.T) {
	_ = container("container_run", "redis", "redis-test")
	//_ = exec.Command("docker ps ", "a-z", "A-Z")
	_ = container("container_list", "redis", "")
	_ = container("container_stop", "redis", "redis-test")
	_ = container("container_remove", "redis", "redis-test")
}

/*func TestHost(t *testing.T) {
	_ = hosts("host_add", "localhost")
	_ = hosts("host_list", "")
	_ = hosts("host_remove", "localhost")
}*/
