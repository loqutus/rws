package storage

import (
	"fmt"
	"github.com/loqutus/rws/pkg/client/conf"
	"io/ioutil"
	"net/http"
	"os"
)

type File struct {
	Name     string
	Host     string
	Size     uint64
	Replicas uint64
}

func Upload(name string) (string, error) {
	file, err1 := os.Open(name)
	if err1 != nil {
		fmt.Println(err1)
		panic("file read error")
	}
	url := conf.HostName + "/storage_upload/" + name
	dat2, err2 := http.Post(url, "application/octet-stream", file)
	if err2 != nil {
		fmt.Println(err2)
		panic("upload error")
	}
	if dat2.StatusCode != 200 {
		fmt.Println(dat2.StatusCode)
		fmt.Println(err2)
		panic("status code error")
	}
	bodyBytes, err2 := ioutil.ReadAll(dat2.Body)
	if err2 != nil {
		fmt.Println(err2)
		panic("body read error")
	}
	return string(bodyBytes), nil
}

func Download(name string) (string, error) {
	url := fmt.Sprintf("%s/storage_download/%s", conf.HostName, name)
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
	return "OK", nil
}

func Remove(name string) (string, error) {
	url := fmt.Sprintf("%s/storage_remove/%s", conf.HostName, name)
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
	return "OK", nil
}

func List() (string, error) {
	url := fmt.Sprintf("%s/storage_list", conf.HostName)
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
	return string(bodyBytes), nil
}

func Help() {
	fmt.Println("upload, download, list or list_all")
}

func ListAll() (string, error) {
	url := conf.HostName + "/storage_list_all"
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
	return string(bodyBytes), nil
}

func Storage(action, name string) {
	if name == "" {
		Help()
		return
	}
	switch action {
	case "help":
		Help()
	case "storage_upload":
		s, err := Upload(name)
		if err != nil {
			fmt.Println(s)
			fmt.Println(err)
			panic("storage upload failed")
		}
		fmt.Println(s)
	case "storage_download":
		s, err := Download(name)
		if err != nil {
			fmt.Println(s)
			fmt.Println(err)
			panic("storage download failed")
		}
		fmt.Println(s)
	case "storage_list":
		s, err := List()
		if err != nil {
			fmt.Println(s)
			fmt.Println(err)
			panic("storage list failure")
		}
		fmt.Println(s)
	case "storage_list_all":
		s, err := ListAll()
		if err != nil {
			fmt.Println(s)
			fmt.Println(err)
			panic("storage list all failed")
		}
		fmt.Println(s)
	case "storage_remove":
		s, err := Remove("name")
		if err != nil {
			fmt.Println(s)
			fmt.Println(err)
			panic("storage remove failed")
		}
		fmt.Println(s)
	}
}
