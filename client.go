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
	dat, err1 := ioutil.ReadFile(name)
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
	url := fmt.Sprintf("%s/storage_download/%s", hostname, name)
	fmt.Println(url)
	dat, err1 := http.Get(url)
	if err1 != nil {
		panic("download error")
	}
	if dat.StatusCode != 200 {
		fmt.Println(dat.StatusCode)
		panic("download error")
	}
	bodyBytes, err2 := ioutil.ReadAll(dat.Body)
	if err2 != nil {
		panic("download error")
	}
	err3 := ioutil.WriteFile(name, []byte(bodyBytes), 0644)
	if err3 != nil {
		panic("file write error")
	}
	return nil
}

func storage(action, name string) {
	if name == "" {
		fmt.Println("upload or download")
		return
	}
	switch action {
	case "help":
		fmt.Println("upload or download and filename")
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
	}
}

func mysqlRun(name string) error {
	return nil
}

func mysqlStop(name string) error {
	return nil
}

func mysql(action, name string) {
	if name == "" {
		fmt.Println("run or stop")
		return
	}
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
	if name == "" {
		fmt.Println("run or stop")
		return
	}
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
		if len(os.Args) < 4 {
			if len(os.Args) > 2 {
				storage(os.Args[2], "")
			} else {
				fmt.Println("upload or download and filename")
			}
		} else {
			storage(os.Args[2], os.Args[3])
		}
	case "mysql":
		if len(os.Args) < 4 {
			if len(os.Args) > 2 {
				fmt.Println("run or stop")
				return
			}
			mysql(os.Args[2], "")
		} else {
			mysql(os.Args[2], os.Args[3])
		}
	case "redis":
		if len(os.Args) < 4 {
			if len(os.Args) > 2 {
				fmt.Println("run or stop")
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
