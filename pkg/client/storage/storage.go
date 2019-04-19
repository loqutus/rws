package storage

import (
	"bytes"
	"fmt"
	"github.com/loqutus/rws/pkg/client/conf"
	"io/ioutil"
	"net/http"
)

type File struct {
	Name     string
	Host     string
	Size     uint64
	Replicas uint64
}

func Upload(name string) (string, error) {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		fmt.Println(err)
		panic("file open error")
	}
	url := fmt.Sprintf("%s/%s/%s", conf.HostName, "storage_upload", name)
	resp, err1 := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err1 != nil {
		fmt.Println(err1)
		panic("request error")
	}
	defer resp.Body.Close()
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
	return string(b), nil
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
