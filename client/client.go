package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
)

const actions = "storage_upload, storage_download, storage_remove, storage_list, storage_list_all, container_run, container_stop, container_list, container_list_all, container_remove, host_add, host_remove, host_list, host_info, pod_add, pod_stop, pod_list, pod_remove"

var HostName string

type Container struct {
	Image string
	Name  string
}

func storageUpload(name string) error {
	dat, err1 := ioutil.ReadFile(name)
	if err1 != nil {
		fmt.Println(err1)
		panic("file read error")
	}
	url := fmt.Sprintf("%s/storage_upload/%s", HostName, name)
	body := bytes.NewBuffer(dat)
	dat2, err2 := http.Post(url, "application/octet-stream", body)
	if err2 != nil {
		fmt.Println(err2)
		panic("upload error")
		return err2
	}
	if dat2.StatusCode != 200 {
		fmt.Println(dat2.StatusCode)
		panic("status code error")
		return http.ErrServerClosed
	}
	return nil
}

func storageDownload(name string) error {
	url := fmt.Sprintf("%s/storage_download/%s", HostName, name)
	dat, err1 := http.Get(url)
	if err1 != nil {
		fmt.Println(err1)
		panic("get error")
	}
	if dat.StatusCode != 200 {
		fmt.Println(dat.StatusCode)
		panic("status code error")
	}
	bodyBytes, err2 := ioutil.ReadAll(dat.Body)
	if err2 != nil {
		fmt.Println(err2)
		panic("body read error")
	}
	err3 := ioutil.WriteFile(name, []byte(bodyBytes), 0644)
	if err3 != nil {
		fmt.Println(err3)
		panic("file write error")
	}
	return nil
}

func storageRemove(name string) error {
	url := fmt.Sprintf("%s/storage_remove/%s", HostName, name)
	dat, err1 := http.Get(url)
	if err1 != nil {
		fmt.Println(err1)
		panic("get error")
	}
	if dat.StatusCode != 200 {
		fmt.Println(dat.StatusCode)
		panic("status code error")
	}
	bodyBytes, err2 := ioutil.ReadAll(dat.Body)
	if err2 != nil {
		fmt.Println(err2)
		panic("body read error")
	}
	print(string(bodyBytes))
	return nil
}

func storageList() error {
	url := fmt.Sprintf("%s/storage_list", HostName)
	dat, err1 := http.Get(url)
	if err1 != nil {
		panic("get error")
	}
	if dat.StatusCode != 200 {
		fmt.Println(dat.StatusCode)
		panic("status code error")
	}
	bodyBytes, err2 := ioutil.ReadAll(dat.Body)
	if err2 != nil {
		fmt.Println(err2)
		panic("body read error")
	}
	fmt.Println(string(bodyBytes))
	return nil
}

func storageHelp() {
	fmt.Println("upload, download, list or list_all")
}

func storageListAll() error {
	url := fmt.Sprintf("%s/storage_list_all", HostName)
	dat, err1 := http.Get(url)
	if err1 != nil {
		panic("get error")
	}
	if dat.StatusCode != 200 {
		fmt.Println(dat.StatusCode)
		panic("status code error")
	}
	bodyBytes, err2 := ioutil.ReadAll(dat.Body)
	if err2 != nil {
		fmt.Println(err2)
		panic("body read error")
	}
	return nil
}

func storage(action, name string) {
	if name == "" {
		storageHelp()
		return
	}
	switch action {
	case "help":
		storageHelp()
	case "storage_upload":
		err := storageUpload(name)
		if err != nil {
			fmt.Println(err)
			panic("storage upload failed")
		}
	case "storage_download":
		err := storageDownload(name)
		if err != nil {
			fmt.Println(err)
			panic("storage download failed")
		}
	case "storage_list":
		err := storageList()
		if err != nil {
			fmt.Println(err)
			panic("storage list failure")
		}
	case "storage_list_all":
		err := storageListAll()
		if err != nil {
			fmt.Println(err)
			panic("storage list all failed")
		}
	case "storage_remove":
		err := storageRemove("name")
		if err != nil {
			fmt.Println(err)
			panic("storage remove failed")
		}
	}
}

// get mysql run
// get mysql stop id
func req(action string, bodyBuffer *bytes.Buffer) ([]byte, error) {
	// http://localhost:8888/container_add
	url := fmt.Sprintf("%s/%s", HostName, action)
	resp, err1 := http.Post(url, "application/json", bodyBuffer)
	defer resp.Body.Close()
	if err1 != nil {
		fmt.Println(err1)
		panic("request error")
	}
	if resp.StatusCode != 200 {
		fmt.Println(resp.StatusCode)
		fmt.Println(resp)
		panic("request status code error")
	}
	b, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		fmt.Println(err2)
		panic("response read error")
	}
	return b, nil
}

// mysql run
// mysql stop id
// mysql list
func container(action, image, name string) string {
	var err error
	var resp []byte
	c := Container{image, name}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(c)
	switch action {
	case "container_list", "container_run", "container_stop", "container_remove", "container_list_all":
		resp, err = req(action, b)
	default:
		panic("unknown action")
	}
	if err != nil {
		fmt.Println(err)
		panic("get error")
	}
	return string(resp)
}

type Host struct {
	Name string
	Port string
}

// hosts add localhost
// hosts delete localhost
// hosts list
func hosts(action, hostName, hostPort string) string {
	var resp []byte
	var err error
	h := Host{hostName, hostPort}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(h)
	switch action {
	case "host_add", "host_remove", "host_list", "host_info":
		resp, err = req(action, b)
		if err != nil {
			fmt.Println(err)
			panic("get error")
		}
		return string(resp)
	default:
		panic("unknown action")
	}
}

type pod struct {
	name    string
	image   string
	cpus    int
	disk    uint64
	memory  uint64
	count   int
	enabled bool
	ids     []string
}

func pods(action, podConfig string) string {
	b := new(bytes.Buffer)
	b.Write([]byte(podConfig))
	switch action {
	case "pod_add", "pod_remove", "pod_list", "pod_info":
		resp, err := req(action, b)
		if err != nil {
			fmt.Println(err)
			panic("get error")
		}
		return string(resp)
	default:
		panic("unknown action")
	}
}

func main() {
	// client --type storage --action upload --name file
	// client --type storage --action list
	var action, name, image, port, podConfig string
	flag.StringVar(&action, "action", "", actions)
	flag.StringVar(&image, "image", "", "redis or mysql")
	flag.StringVar(&name, "name", "", "container/file/host name")
	flag.StringVar(&port, "port", "", "host port")
	flag.StringVar(&podConfig, "pod_config", "", "pod config")
	flag.StringVar(&HostName, "hostname", "http://localhost:8888", "hostname to connect to")
	flag.Parse()
	switch action {
	case "storage_upload", "storage_download", "storage_remove", "storage_list", "storage_list_all":
		if name != "" && action != "storage_list" && action != "storage_list_all" {
			storage(action, name)
		} else if name == "" && action == "storage_list" {
			storage(action, "")
		} else {
			panic("file name required")
		}
	case "container_run", "container_stop", "container_list, container_list_all container_remove":
		_ = container(action, image, name)
	case "host_add", "host_remove", "host_list", "host_info":
		_ = hosts(action, name, port)
	case "pod_add", "pod_stop", "pod_remove", "pod_list":
		_ = pods(action, podConfig)
	default:
		panic(actions)
	}
}
