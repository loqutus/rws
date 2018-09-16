package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func storageUploadHandler(w http.ResponseWriter, r *http.Request) {

}

func storageDownloadHandler(w http.ResponseWriter, r *http.Request) {
	pathSplit := strings.Split(r.URL.Path, "/")
	filename := pathSplit[len(pathSplit)-1]
	dat, err := ioutil.ReadFile(fmt.Sprintf("data/%s", filename))
	if err != nil {
		fmt.Println("file reading error: %s", filename)
	}
	w.Write(dat)
}

func main() {
	http.HandleFunc("/storage_upload/*", storageUploadHandler)
	http.HandleFunc("/storage_download/*", storageDownloadHandler)
	if err := http.ListenAndServe(":8888", nil); err != nil {
		panic(err)
	}
}
