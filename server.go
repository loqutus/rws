package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
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

func run(t string) error {
	fmt.Println(t + " run")
	pathSplit := strings.Split(r.URL.Path, "/")
	name := pathSplit[len(pathSplit)-1]
	fmt.Println(name)
	c := client.WithVersion("1.38")
	cli, err := client.NewClientWithOpts(c)
	if err != nil {
		fmt.Println("mysql run error")
		fmt.Println(err)
	}
}

func mysqlRunHandler(w http.ResponseWriter, r *http.Request) {
	run("mysql")
	containers, err2 := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err2 != nil {
		fmt.Println("mysql run error")
		fmt.Println(err2)
	}

	for _, container := range containers {
		fmt.Printf("%s %s\n", container.ID[:10], container.Image)
	}
}

func mysqlStopHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("mysql stop")
	pathSplit := strings.Split(r.URL.Path, "/")
	name := pathSplit[len(pathSplit)-1]
	fmt.Println(name)
	_, err := client.NewEnvClient()
	if err != nil {
		fmt.Println("mysql create error")
		fmt.Println(err)
	}

}

func main() {
	fmt.Println("starting server")
	http.HandleFunc("/storage_upload/", storageUploadHandler)
	http.HandleFunc("/storage_download/", storageDownloadHandler)
	http.HandleFunc("/mysql_run/", mysqlRunHandler)
	http.HandleFunc("/mysql_stop/", mysqlStopHandler)
	http.HandleFunc("/redis_run/", mysqlRunHandler)
	http.HandleFunc("/redis_stop/", mysqlRunHandler)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
