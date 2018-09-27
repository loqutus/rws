package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
)

const hostname = "http://localhost:8888"

func storageUpload(name string) error {
	dat, err1 := ioutil.ReadFile(name)
	if err1 != nil {
		fmt.Println(err1)
		panic("file read error")
	}
	url := fmt.Sprintf("%s/upload/%s", hostname, name)
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
	url := fmt.Sprintf("%s/download/%s", hostname, name)
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

func storageList() error {
	url := fmt.Sprintf("%s/list", hostname)
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
	fmt.Println("upload, download or list")
}

func storage(action, name string) {
	if name == "" {
		storageHelp()
		return
	}
	switch action {
	case "help":
		storageHelp()
	case "upload":
		err := storageUpload(name)
		if err != nil {
			fmt.Println(err)
			panic("storage upload failure")
		}
	case "download":
		err := storageDownload(name)
		if err != nil {
			fmt.Println(err)
			panic("storage download failure")
		}
	case "list":
		err := storageList()
		if err != nil {
			fmt.Println(err)
			panic("storage list failure")
		}
	}
}

// get mysql run
// get mysql stop id
func req(httpMethod, action, containerType, id string, body []byte) ([]byte, error) {
	url := ""
	if id != "" {
		// http://localhost:8888/stop/redis/ID
		url = fmt.Sprintf("%s/%s/%s/%s", hostname, action, containerType, id)
	} else {
		// http://localhost:8888/start/redis
		url = fmt.Sprintf("%s/%s/%s", hostname, action, containerType)
	}
	var err1 error
	var resp *http.Response
	if httpMethod == "post" {
		bodybuffer := bytes.NewBuffer(body)
		resp, err1 = http.Post(url, "application/octet-stream", bodybuffer)
		defer resp.Body.Close()
	} else {
		resp, err1 = http.Get(url)
	}
	if err1 != nil {
		fmt.Println(err1)
		panic("request error")
	}
	if resp.StatusCode != 200 {
		fmt.Println(resp.StatusCode)
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
func container(containerType, action, name string) {
	switch containerType {
	case "mysql", "redis":
		switch action {
		case "run", "stop", "list":
			resp, err := req("get", action, containerType, name, []byte(""))
			if err != nil {
				fmt.Println(err)
				panic("get error")
			}
			fmt.Println(resp)
		}
	default:
		panic("unknown type")
	}
	panic("wrong type")
}

func main() {
	// client --type storage --action upload --name file
	// client --type storage --action list
	var typ, action, name string
	flag.StringVar(&typ, "type", "", "storage, mysql or redis")
	flag.StringVar(&action, "action", "", "upload, download, run, stop or list")
	flag.StringVar(&name, "name", "", "container/file name")
	flag.Parse()
	switch action {
	case "upload":
		if name != "" {
			storage("upload", name)
		} else {
			panic("file name required")
		}
	case "download":
		if name != "" {
			storage("download", name)
		} else {
			panic("file name required")
		}
	case "run", "stop", "list":
		container(typ, action, name)
	default:
		panic("upload, download, run, stop or list")
	}
}
