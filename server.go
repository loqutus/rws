package main

import (
	"fmt"
	"github.com/docker/docker/client"
	"io/ioutil"
	"net/http"
	"strings"
)

const addr = "localhost:8888"
const dataDir = "data"

func storageUploadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("storage upload")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("request reading error")
		return
	}
	pathSplit := strings.Split(r.URL.Path, "/")
	filename := pathSplit[len(pathSplit)-1]
	err3 := ioutil.WriteFile(fmt.Sprintf("%s/%s", dataDir, filename), []byte(body), 0644)
	if err3 != nil {
		fmt.Println("file write error")
		return
	}
}

func storageDownloadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("storage download")
	pathSplit := strings.Split(r.URL.Path, "/")
	filename := pathSplit[len(pathSplit)-1]
	fmt.Println("download: " + filename)
	dat, err1 := ioutil.ReadFile(fmt.Sprintf("data/%s", filename))
	if err1 != nil {
		fmt.Println("file read error: " + filename)
		return
	}
	w.Write(dat)
}

func req() *client.Client {
	pathSplit := strings.Split(r.URL.Path, "/")
	name := pathSplit[len(pathSplit)-1]
	fmt.Println(name)
	c := client.WithVersion("1.38")
	cli, err := client.NewClientWithOpts(c)
	if err != nil {
		fmt.Println("client creation error")
		fmt.Println(err)
	}
	return cli
}

func run(t string, w http.ResponseWriter, r *http.Request) error {
	fmt.Println(t + " run")
	c := req()
	return nil
}

func stop(t string, w http.ResponseWriter, r *http.Request) error {
	fmt.Println(t + " stop")
	c := req()
	return nil
}

func mysqlRunHandler(w http.ResponseWriter, r *http.Request) {
	run("mysql", w, r)

}

func mysqlStopHandler(w http.ResponseWriter, r *http.Request) {
	stop("mysql", w, r)
	return
}

func redisRunHandler(w http.ResponseWriter, r *http.Request) {
	run("redis", w, r)
}

func redisStopHandler(w http.ResponseWriter, r *http.Request) {
	stop("redis", w, r)
	return
}

func main() {
	fmt.Println("starting server")
	http.HandleFunc("/storage_upload/", storageUploadHandler)
	http.HandleFunc("/storage_download/", storageDownloadHandler)
	http.HandleFunc("/mysql_run/", mysqlRunHandler)
	http.HandleFunc("/mysql_stop/", mysqlStopHandler)
	http.HandleFunc("/redis_run/", redisRunHandler)
	http.HandleFunc("/redis_stop/", redisStopHandler)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
