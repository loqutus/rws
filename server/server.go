package main

import (
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
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
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	pathSplit := strings.Split(r.URL.Path, "/")
	filename := fmt.Sprintf("%s/%s", dataDir, pathSplit[len(pathSplit)-1])
	err2 := ioutil.WriteFile(filename, []byte(body), 0644)
	if err2 != nil {
		fmt.Println("file write error")
		fmt.Println(err2)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Println("file %s uploaded", filename)
	w.WriteHeader(http.StatusOK)
}

func storageDownloadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("storage download")
	pathSplit := strings.Split(r.URL.Path, "/")
	filename := pathSplit[len(pathSplit)-1]
	dat, err1 := ioutil.ReadFile(fmt.Sprintf("data/%s", filename))
	if err1 != nil {
		fmt.Println("file read error: " + filename)
		fmt.Println(err1)

		return
	}
	_, err2 := w.Write(dat)
	if err2 != nil {
		fmt.Println("request write error")
		fmt.Println(err2)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Println("file %s downloaded", filename)
	w.WriteHeader(http.StatusOK)
}

func storageListHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("storage list")
	files, err := ioutil.ReadDir(dataDir)
	if err != nil {
		fmt.Println("dir list error")
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var l []string
	for _, f := range files {
		l = append(l, f.Name())
	}
	s := strings.Join(l, "\n")
	_, err2 := w.Write([]byte(s))
	if err2 != nil {
		fmt.Println("request write error")
		fmt.Println(err2)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func listContainers(typeName string) string {
	c := client.WithVersion("1.38")
	cli, err := client.NewClientWithOpts(c)
	if err != nil {
		fmt.Println("client create error")
		fmt.Println(err)
		return ""
	}
	var l []string
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		fmt.Println("containerList error")
		fmt.Println(err)
		return ""
	}
	for _, c := range containers {
		l = append(l, c.ID)
	}
	s := strings.Join(l, "\n")
	return s
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	pathSplit := strings.Split(r.URL.Path, "/")
	typeName := pathSplit[len(pathSplit)-1]
	s := listContainers(typeName)
	w.Write([]byte(s))
}

func runContainer(imageName string) (string, error) {
	ctx := context.Background()
	c := client.WithVersion("1.38")
	cli, err := client.NewClientWithOpts(c)
	if err != nil {
		fmt.Println("client create error")
		fmt.Println(err)
		return "", err
	}
	out, err2 := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err2 != nil {
		fmt.Println("image pull error")
		fmt.Println(out)
		return "", err2
	}
	resp, err3 := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
	}, nil, nil, "")
	if err3 != nil {
		fmt.Println("container create error")
		fmt.Println(resp)
		return "", err3
	}
	err4 := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err4 != nil {
		fmt.Println("container start error")
		return "", err4
	}
	return resp.ID, nil
}

func stopContainer(containerId string) error {
	ctx := context.Background()
	c := client.WithVersion("1.38")
	cli, err1 := client.NewClientWithOpts(c)
	if err1 != nil {
		fmt.Println("client create error")
		fmt.Println(err1)
		return err1
	}
	err2 := cli.ContainerStop(ctx, containerId, nil)
	if err2 != nil {
		fmt.Println("container stop error")
		fmt.Println(err2)
		return err2
	}
	return nil
}

func runHandler(w http.ResponseWriter, r *http.Request) {
	pathSplit := strings.Split(r.URL.Path, "/")
	imageName := pathSplit[len(pathSplit)-1]
	id, err := runContainer(imageName)
	if err == nil {
		fmt.Fprintf(w, id)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
	}
}

func stopHandler(w http.ResponseWriter, r *http.Request) {
	pathSplit := strings.Split(r.URL.Path, "/")
	containerId := pathSplit[len(pathSplit)-1]
	err := stopContainer(containerId)
	if err == nil {
		fmt.Fprintf(w, "OK")
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
	}
}

func main() {
	fmt.Println("starting server")
	http.HandleFunc("/upload/", storageUploadHandler)
	http.HandleFunc("/download/", storageDownloadHandler)
	http.HandleFunc("/list", storageListHandler)
	http.HandleFunc("/run/", runHandler)
	http.HandleFunc("/stop/", stopHandler)
	http.HandleFunc("/list/", listHandler)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
