package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestStorage(t *testing.T) {
	fmt.Println("TestStorage: test storage upload")
	HostName = "http://localhost:8888"
	err := storageUpload("test")
	if err != nil {
		fmt.Println(err)
		t.Errorf("TestStorage: storage upload error")
	}
	dat, err := ioutil.ReadFile("test")
	if err != nil {
		fmt.Println(err)
		t.Errorf("TestStorage: upload file read error")
	}
	if bytes.Compare(dat, []byte("test\n")) != 0 {
		fmt.Println(dat)
		t.Errorf("TestStorage: upload file content error")
	}
	fmt.Println("TestStorage: test storage download")
	err2 := storageDownload("test")
	if err2 != nil {
		fmt.Println(err2)
		t.Errorf("storage download error")
	}
	fmt.Println("TestStorage: test storage list")
	err3 := storageList()
	if err3 != nil {
		fmt.Println(err3)
		t.Errorf("storage list error")
	}
	fmt.Println("TestStorage: test storage list all")
	err4 := storageListAll()
	if err4 != nil {
		fmt.Println(err4)
		t.Errorf("storage list all error")
	}
	fmt.Println("TestStorage: test storage remove")
	err5 := storageRemove("test")
	if err5 != nil {
		fmt.Println(err5)
		t.Errorf("storage remove error")
	}
}

func TestContainer(t *testing.T) {
	fmt.Println("test container run")
	var empty_cmd []string
	_ = container("container_run", "redis", "redis-test", empty_cmd)
	fmt.Println("test container list")
	_ = container("container_list", "redis", "", empty_cmd)
	fmt.Println("test container list all")
	_ = container("container_list_all", "redis", "", empty_cmd)
	fmt.Println("test container stop")
	_ = container("container_stop", "redis", "redis-test", empty_cmd)
	fmt.Println("container_remove")
	_ = container("container_remove", "redis", "redis-test", empty_cmd)
}

func TestHost(t *testing.T) {
	fmt.Println("test host add")
	_ = hosts("host_add", "localhost", "9999")
	fmt.Println("test host list")
	_ = hosts("host_list", "", "")
	fmt.Println("test host remove")
	_ = hosts("host_remove", "localhost", "9999")
}

func TestHostInfo(t *testing.T) {
	fmt.Println("test host info")
	_ = hosts("host_info", "", "")
}

func TestPod(t *testing.T) {
	fmt.Println("test Pod add")
	_ = pods("pod_add", Pod{})
	fmt.Println("test Pod list")
	_ = pods("pod_list", Pod{})
	fmt.Println("test Pod remove")
	_ = pods("pod_remove", Pod{})
}
