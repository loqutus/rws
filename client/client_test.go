package main

import (
	"fmt"
	"testing"
)

func TestStorage(t *testing.T) {
	fmt.Println("test storage upload")
	err := storageUpload("test")
	if err != nil {
		fmt.Println(err)
		t.Errorf("storage upload error")
	}
	fmt.Println("test storage download")
	err2 := storageDownload("test")
	if err2 != nil {
		fmt.Println(err2)
		t.Errorf("storage download error")
	}
	fmt.Println("test storage list")
	err3 := storageList()
	if err3 != nil {
		fmt.Println(err3)
		t.Errorf("storage list error")
	}
	fmt.Println("test storage remove")
	err4 := storageRemove("test")
	if err4 != nil {
		fmt.Println(err4)
		t.Errorf("storage remove error")
	}
}

func TestContainer(t *testing.T) {
	fmt.Println("test container run")
	_ = container("container_run", "redis", "redis-test")
	//_ = exec.Command("docker ps ", "a-z", "A-Z")
	fmt.Println("test container list")
	_ = container("container_list", "redis", "")
	fmt.Println("test container stop")
	_ = container("container_stop", "redis", "redis-test")
	fmt.Println("container_remove")
	_ = container("container_remove", "redis", "redis-test")
}

/*func TestHost(t *testing.T) {
	_ = hosts("host_add", "localhost")
	_ = hosts("host_list", "")
	_ = hosts("host_remove", "localhost")
}*/
