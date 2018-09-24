package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

const hostname = "http://localhost:8888"
const dataDir = "data"

func get(actionType, name string) ([]bytes, error) {
	url := fmt.Sprintf("%s/%s/%s", hostname, actionType, name)
	body := bytes.NewBuffer([]byte(""))
	resp, err1 := http.Post(url, "application/octet-stream", body)
	defer resp.Body.Close()
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
	return string(b), nil
}

func list(name string) error {
	e := req(name, "list")
	return e
}

func run(name, t string) error {
	e := req(name, "run")
	return e
}

func stop(name, t string) error {
	e := req(name, t)
	return e
}

func storageUpload(name string) error {
	dat, err1 := ioutil.ReadFile(name)
	if err1 != nil {
		panic("file read error")
	}
	url := fmt.Sprintf("%s/upload/%s", hostname, name)
	body := bytes.NewBuffer(dat)
	dat2, err2 := http.Post(url, "application/octet-stream", body)
	if err2 != nil {
		panic("upload error")
	}
	if dat2.StatusCode != 200 {
		fmt.Println(dat2.StatusCode)
		panic("download error")
	}
	return nil
}

func storageDownload(name string) error {
	url := fmt.Sprintf("%s/download/%s", hostname, name)
	fmt.Println(url)
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
	err3 := ioutil.WriteFile(name, []byte(bodyBytes), 0644)
	if err3 != nil {
		fmt.Println(err3)
		panic("file write error")
	}
	return nil
}

func storageList() error {
	url := fmt.Sprintf("%s/list", hostname)
	fmt.Println(url)
	dat, err1 := http.Get(url)
	if err1 != nil {
		panic("get error")
	}
	if dat.StatusCode != 200 {
		fmt.Println(dat.StatusCode)
		panic("status code error")
	}
	bodyBytes, err2 := ioutil.ReadAll(dat.Body)
	if err2 != nil{
		fmt.Println(err2)
		panic("body read error")
	}
	fmt.Println(bodyBytes)
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

func mysqlHelp() {
	fmt.Println("run, list or stop")
}

func mysql(action, name string) {
	if name == "" {
		mysqlHelp()
		return
	}
	switch action {
	case "help":
		mysqlHelp()
	case "run":
		err := mysqlRun(name)
		if err != nil {
			panic("mysql run error")
		}
	case "stop":
		err := mysqlStop(name)
		if err != nil {
			panic("mysql stop error")
		}
	case "list":
		err := mysqlList()
		if err != nil {
			panic("mysql list error")
		}
	default:
		mysqlHelp()

	}
}

func redis(action, name string) {
	if name == "" {
		redisHelp()
		return
	}
	switch action {
	case "help":
		redisHelp()
	case "run":
		err := redisRun(name)
		if err != nil {
			panic("redis run error")
		}
	case "stop":
		err := redisStop(name)
		if err != nil {
			panic("redis stop error")
		}
	case "list":
		err := redisList()
		if err != nil {
			panic("redis list error")
		}
	default:
		redisHelp()
		return
	}
}

func printHelp() {
	fmt.Println("storage, mysql or redis")
}

func main() {
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
			mysql(os.Args[2], "")
		} else {
			mysql(os.Args[2], os.Args[3])
		}
	case "redis":
		if len(os.Args) < 4 {
			if len(os.Args) > 2 {
				fmt.Println("run, list or stop")
				return
			}
			redis(os.Args[2], "")
		} else {
			redis(os.Args[2], os.Args[3])
		}
	default:
		printHelp()
	}
}
