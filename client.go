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

func list(name string) error {
	url := fmt.Sprintf("%s/%s_list", name, hostname)
	body := bytes.NewBuffer([]byte(""))
	resp, err1 := http.Post(url, "application/octet-stream", body)
	if err1 != nil {
		fmt.Println(err1)
		panic("run error")
	}
	defer resp.Body.Close()
	b, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		fmt.Println(err2)
		panic("response read error")
	}
	fmt.Println(b)
	return nil
}

func run(name, t string) error {
	url := fmt.Sprintf("%s/%s_run/%s", hostname, t, name)
	body := bytes.NewBuffer([]byte(""))
	dat, err1 := http.Post(url, "application/octet-stream", body)
	if err1 != nil {
		fmt.Println(err1)
		panic("run error")
	}
	if dat.StatusCode != 200 {
		fmt.Println(dat.StatusCode)
		panic("post error")
	}
	return nil
}

func stop(name, t string) error {
	url := fmt.Sprintf("%s/%s_stop/%s", t, name)
	body := bytes.NewBuffer([]byte(""))
	dat, err1 := http.Post(url, "application/octet-stream", body)
	if err1 != nil {
		fmt.Println(err1)
		panic("stop error")
	}
	if dat.StatusCode != 200 {
		fmt.Println(dat.StatusCode)
		panic("post error")
	}
	return nil
}

func storageUpload(name string) error {
	dat, err1 := ioutil.ReadFile(name)
	if err1 != nil {
		panic("file read error")
	}
	url := fmt.Sprintf("%s/storage_upload/%s", hostname, name)
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
		fmt.Println(err2)
		panic("download error")
	}
	err3 := ioutil.WriteFile(name, []byte(bodyBytes), 0644)
	if err3 != nil {
		fmt.Println(err3)
		panic("file write error")
	}
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
	}
}

func mysqlRun(name string) error {
	run(name, "mysql")
	return nil
}

func mysqlStop(name string) error {
	stop(name, "mysql")
	return nil
}
func mysqlList() error {
	list("mysql")
	return nil
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

func redisRun(name string) error {
	run(name, "redis")
	return nil
}

func redisStop(name string) error {
	stop(name, "redis")
	return nil
}

func redisList() error {
	list("redis")
	return nil
}

func redisHelp() {
	fmt.Println("run, list or stop")
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
