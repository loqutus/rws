package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/loqutus/rws/pkg/client/containers"
	"github.com/loqutus/rws/pkg/client/hosts"
	"github.com/loqutus/rws/pkg/client/pods"
	"github.com/loqutus/rws/pkg/client/storage"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
)

func TestHosts(t *testing.T) {
	fmt.Println("TestHosts: add hosts")
	hosts.HostsAction("host_add", "localhost", "8888")
	for i := 1; i <= 5; i++ {
		hosts.HostsAction("host_add", "pi"+strconv.Itoa(i), "8888")
	}
}

func TestStorage(t *testing.T) {
	fmt.Println("TestStorage: test storage upload")
	s, err := storage.Upload("../../test/test")
	if err != nil {
		fmt.Println(s)
		fmt.Println(err)
		t.Errorf("TestStorage: storage upload error")
	}
	/*dat, err := ioutil.ReadFile(conf.StorageTestDir + "/data/test")
	if err != nil {
		fmt.Println(err)
		t.Errorf("TestStorage: upload file read error")
	}
	if bytes.Compare(dat, []byte("test\n")) != 0 {
		fmt.Println("TestStorage: upload file content error:")
		fmt.Println("Should be: test\\n")
		fmt.Println("Got: ", dat)
		t.Errorf("TestStorage: upload file content error")
	}
	*/
	fmt.Println("TestStorage: test storage download")
	s, err2 := storage.Download("test")
	if err2 != nil {
		fmt.Println(err2)
		t.Errorf("storage download error")
	}
	dat, err := ioutil.ReadFile("test")
	if err != nil {
		fmt.Println(err)
		t.Errorf("TestStorage: downloaded file read error")
	}
	if bytes.Compare(dat, []byte("test\n")) != 0 {
		fmt.Println(dat)
		t.Errorf("TestStorage: downloaded file content error")
	}
	fmt.Println("TestStorage: test storage list")
	s, err3 := storage.List()
	if err3 != nil {
		fmt.Println(err3)
		t.Errorf("storage list error")
	}
	var c []storage.File
	err4 := json.Unmarshal([]byte(s), &c)
	if err4 != nil {
		fmt.Println(err4)
		t.Errorf("storage list json.Unmarshal error")
	}
	var z = []storage.File{storage.File{"test", "localhost:8888", 5, 1}}
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
	_, err5 := storage.Remove("test")
	if err5 != nil {
		fmt.Println(err5)
		t.Errorf("storage remove error")
	}
	if _, err := os.Stat("../server/data/test"); os.IsNotExist(err) == false {
		t.Errorf("file test exists, should be removed")
	}
}

func ListLocalContainers() ([]types.Container, error) {
	c := client.WithVersion("1.38")
	cli, err := client.NewClientWithOpts(c)
	if err != nil {
		fmt.Println("ListLocalContainers: client create error")
		fmt.Println(err)
		return nil, err
	}
	localContainers, err2 := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err2 != nil {
		fmt.Println("ListLocalContainers: containerList error")
		fmt.Println(err2)
		return nil, err2
	}
	return localContainers, nil
}

func TestContainer(t *testing.T) {
	fmt.Println("test container run")
	cmd := []string{"/bin/sleep", "60"}
	containerID := containers.ContainerAction("container_run", "arm32v6/alpine", "test", cmd)
	localContainers1, err := ListLocalContainers()
	if err != nil {
		panic(err)
	}
	found := false
	for _, localContainer := range localContainers1 {
		if localContainer.ID == containerID {
			found = true
			break
		}
	}
	if found != true {
		panic("container is not running")
	}
	fmt.Println("test container list")
	allContainersString := containers.ContainerAction("container_list", "", "", cmd)
	var allContainers []containers.Container
	err2 := json.Unmarshal([]byte(allContainersString), &allContainers)
	if err2 != nil {
		fmt.Println(err2)
		panic("json.Unmarshal error")
	}
	found = false
	for _, allContainer := range allContainers {
		if allContainer.ID == containerID {
			found = true
			break
		}
	}
	if found != true {
		panic("container is not found in all containers list")
	}
	fmt.Println("test container list local")
	localContainersString := containers.ContainerAction("container_list_local", "", "", cmd)
	var localContainers []containers.Container
	err3 := json.Unmarshal([]byte(localContainersString), &localContainers)
	if err3 != nil {
		panic("json.Unmarshal error")
	}
	found = false
	for _, localContainer := range localContainers {
		if localContainer.ID == containerID {
			found = true
			break
		}
	}
	if found != true {
		panic("container not found in all containers list")
	}
	fmt.Println("test container stop")
	_ = containers.ContainerAction("container_stop", "", "test", cmd)
	localContainersString = containers.ContainerAction("container_list_local", "", "", cmd)
	localContainers = []containers.Container{}
	err4 := json.Unmarshal([]byte(localContainersString), &localContainers)
	if err4 != nil {
		panic("json.Unmarshal error")
	}
	found = false
	for _, localContainer := range localContainers {
		if localContainer.ID == containerID {
			found = true
			break
		}
	}
	if found == true {
		panic("container found in local containers list")
	}
	fmt.Println("container_remove")
	_ = containers.ContainerAction("container_remove", "", "test", cmd)
	for i := 1; i < 10; i++ {
		_ = containers.ContainerAction("container_run", "arm32v6/alpine", "test"+strconv.Itoa(i), []string{"/bin/sleep", "60"})
	}
}

func TestHostInfo(t *testing.T) {
	fmt.Println("test host info")
	_ = hosts.HostsAction("host_info", "", "")
}

func TestPod(t *testing.T) {
	fmt.Println("test Pod add")
	cmd := []string{"/bin/sleep", "60"}
	var cont []containers.Container
	pod := pods.Pod{"pod-test", "alpine", 1, 1, 1, 1, cmd, cont}
	_ = pods.PodsAction("pod_add", pod)
	fmt.Println("test Pod list")
	_ = pods.PodsAction("pod_list", pods.Pod{})
	fmt.Println("test Pod remove")
	_ = pods.PodsAction("pod_remove", pod)
}

func TestHost(t *testing.T) {
	fmt.Println("test host list")
	_ = hosts.HostsAction("host_list", "", "")
	fmt.Println("test host remove")
	_ = hosts.HostsAction("host_remove", "localhost", "8888")
}
