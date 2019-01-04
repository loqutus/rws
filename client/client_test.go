package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestStorage(t *testing.T) {
	fmt.Println("TestStorage: test storage upload")
	HostName = "http://localhost:8888"
	s, err := storageUpload("test")
	if err != nil {
		fmt.Println(s)
		fmt.Println(err)
		t.Errorf("TestStorage: storage upload error")
	}
	dat, err := ioutil.ReadFile("test")
	if err != nil {
		fmt.Println(err)
		t.Errorf("TestStorage: upload file read error")
	}
	if bytes.Compare(dat, []byte("test\n")) != 0 {
		fmt.Println(s)
		t.Errorf("TestStorage: upload file content error")
	}
	fmt.Println("TestStorage: test storage download")
	s, err2 := storageDownload("test")
	if err2 != nil {
		fmt.Println(err2)
		t.Errorf("storage download error")
	}
	dat, err = ioutil.ReadFile("test")
	if err != nil {
		fmt.Println(err)
		t.Errorf("TestStorage: downloaded file read error")
	}
	if bytes.Compare(dat, []byte("test\n")) != 0 {
		fmt.Println(dat)
		t.Errorf("TestStorage: downloaded file content error")
	}
	fmt.Println("TestStorage: test storage list")
	s, err3 := storageList()
	if err3 != nil {
		fmt.Println(err3)
		t.Errorf("storage list error")
	}
	var c []File
	err4 := json.Unmarshal([]byte(s), &c)
	if err4 != nil {
		fmt.Println(err4)
		t.Errorf("storage list json.Unmarshal error")
	}
	var z = []File{File{"test", "10.0.0.1:8888", 5, 1}}
	if len(z) != len(c) || z[0].Name != c[0].Name || z[0].Host != c[0].Host ||
		z[0].Replicas != c[0].Replicas || z[0].Size != c[0].Size {
		fmt.Println("Got: " + s)
		fileBytes, err := json.Marshal(z)
		if err != nil {
			fmt.Println("json.Marshal error")
		}
		fmt.Println("Should be: " + string(fileBytes))
		t.Errorf("storage list not right")
	}
	fmt.Println("TestStorage: test storage remove")
	_, err5 := storageRemove("test")
	if err5 != nil {
		fmt.Println(err5)
		t.Errorf("storage remove error")
	}
	if _, err := os.Stat("../server/data/test"); os.IsNotExist(err) == false {
		t.Errorf("file test exists, should be removed")
	}
}

func TestContainer(t *testing.T) {
	fmt.Println("test container run")
	cmd := []string{"/bin/sleep", "60"}
	_ = container("container_run", "arm32v6/alpine", "test", cmd)
	fmt.Println("test container list")
	_ = container("container_list", "", "", cmd)
	fmt.Println("test container list local")
	_ = container("container_list_local", "", "", cmd)
	fmt.Println("test container stop")
	_ = container("container_stop", "", "test", cmd)
	fmt.Println("container_remove")
	_ = container("container_remove", "", "test", cmd)
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
