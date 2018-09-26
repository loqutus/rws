package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
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

func req(actionType, method, name, id string) ([]byte, error) {
	url := ""
	if name != "" {
		// http://localhost:8888/run/redis
		url = fmt.Sprintf("%s/%s/%s/%s", hostname, actionType, name, id)
	} else {
		// http://localhost:8888/stop/redis/ID
		url = fmt.Sprintf("%s/%s/%s", hostname, actionType, name)
	}
	fmt.Print(url)
	var err1 error
	var resp *http.Response
	if method == "post" {
		body := bytes.NewBuffer([]byte(""))
		resp, err1 = http.Post(url, "application/octet-stream", body)
		defer resp.Body.Close()
	} else {
		resp, err1 = http.Get(url)
	}
	if err1 != nil {
		fmt.Println(err1)
		panic("post error")
	}
	if resp.StatusCode != 200 {
		fmt.Println(resp.StatusCode)
		panic("post error")
	}
	b, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		fmt.Println(err2)
		panic("response read error")
	}
	return b, nil
}

func printHelp() {
	fmt.Println("storage, mysql or redis")
}

func main() {
	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "help":
			printHelp()
		case "storage":
			if len(os.Args) < 4 {
				if len(os.Args) > 2 {
					storage(os.Args[2], "")
				} else {
					fmt.Println("upload, download or list and filename")
				}
			} else {
				storage(os.Args[2], os.Args[3])
			}
		case "mysql":
			if len(os.Args) < 4 {
				if len(os.Args) > 2 {
					fmt.Println("run, list or stop")
					return
				}
				req("mysql", "get", os.Args[2], "")
			} else {
				req("mysql", "post", os.Args[2], os.Args[3])
			}
		case "redis":
			if len(os.Args) < 4 {
				if len(os.Args) > 2 {
					fmt.Println("run, list or stop")
					return
				}
				req("redis", "get", os.Args[2], "")
			} else {
				req("redis", "post", os.Args[2], os.Args[3])
			}
		default:
			printHelp()

		}
	} else {
		printHelp()
		os.Exit(1)
	}
}
