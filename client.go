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

func storageUpload(name string) error {
	dat, err1 := ioutil.ReadFile(fmt.Sprintf("%s/%s", dataDir, name))
	if err1 != nil {
		panic("file read error")
	}
	url := fmt.Sprintf("%s/storage_upload/%s", hostname, name)
	body := bytes.NewBuffer(dat)
	_, err2 := http.Post(url, "application/octet-stream", body)
	if err2 != nil {
		panic("upload error")
	}
	return nil
}

func storageDownload(name string) error {
	dat, err1 := http.Get(fmt.Sprintf("%s/storage_download/%s", hostname, name))
	if err1 != nil {
		panic("download error")
	}
	bodyBytes, err2 := ioutil.ReadAll(dat.Body)
	if err2 != nil {
		panic("download error")
	}
	err3 := ioutil.WriteFile(fmt.Sprintf("%s/%s", dataDir, name), []byte(bodyBytes), 0644)
	if err3 != nil {
		panic("file write error")
	}
	return nil
}

func storage(action, name string) {
	switch action {
	case "help":
		fmt.Println("upload or download and filename")
	case "upload":
		err := storageUpload(name)
		if err != nil {
			panic("storage upload failure")
		}
	case "download":
		err := storageDownload(name)
		if err != nil {
			panic("storage download failure")
		}
	}
}

func mysqlRun(name string) error {
	return nil
}

func mysqlStop(name string) error {
	return nil
}

func mysql(action, name string) {
	switch action {
	case "help":
		fmt.Println("run or stop")
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
	default:
		fmt.Println("run or stop")

	}
}

func redisRun(name string) error {
	return nil
}

func redisStop(name string) error {
	return nil
}

func redis(action, name string) {
	switch action {
	case "help":
		fmt.Println("run or stop")
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
	default:
		fmt.Println("run or stop")

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
		storage(os.Args[2], os.Args[3])
	case "mysql":
		mysql(os.Args[2], os.Args[3])
	case "redis":
		redis(os.Args[2], os.Args[3])
	default:
		printHelp()
	}
}
