package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

const addr = "localhost:8888"
const dataDir = "data"

var hosts map[string]bool

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
	fileName := fmt.Sprintf("%s/%s", dataDir, pathSplit[len(pathSplit)-1])
	err2 := ioutil.WriteFile(fileName, []byte(body), 0644)
	if err2 != nil {
		fmt.Println("file write error")
		fmt.Println(err2)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Println("file " + fileName + " uploaded")
	w.WriteHeader(http.StatusOK)
}

func storageDownloadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("storage download")
	pathSplit := strings.Split(r.URL.Path, "/")
	fileName := pathSplit[len(pathSplit)-1]
	dat, err1 := ioutil.ReadFile(fmt.Sprintf("data/%s", fileName))
	if err1 != nil {
		fmt.Println("file read error: " + fileName)
		fmt.Println(err1)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err2 := w.Write(dat)
	if err2 != nil {
		fmt.Println("request write error")
		fmt.Println(err2)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Println("file " + fileName + " downloaded")
	w.WriteHeader(http.StatusOK)
}

func storageRemoveHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("storage remove")
	pathSplit := strings.Split(r.URL.Path, "/")
	fileName := pathSplit[len(pathSplit)-1]
	err := os.Remove(fmt.Sprintf("data/%s", fileName))
	if err != nil {
		fmt.Println("file remove error: " + fileName)
		fmt.Println(err)
		return
	}
	fmt.Println("file " + fileName + " removed")
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
	fmt.Println("list containers")
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
		for _, n := range c.Names {
			l = append(l, n)
		}
	}
	s := strings.Join(l, "\n")
	return s
}

func containerListhandler(w http.ResponseWriter, r *http.Request) {
	pathSplit := strings.Split(r.URL.Path, "/")
	typeName := pathSplit[len(pathSplit)-1]
	s := listContainers(typeName)
	w.Write([]byte(s))
}

func runContainer(imageName, containerName string) (string, error) {
	fmt.Println("run container")
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
		fmt.Println(err2)
		return "", err2
	}
	resp, err3 := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
	}, nil, nil, containerName)
	if err3 != nil {
		fmt.Println("container create error")
		fmt.Println(resp)
		fmt.Println(err3)
		return "", err3
	}
	err4 := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err4 != nil {
		fmt.Println("container start error")
		return "", err4
	}
	return resp.ID, nil
}

func stopContainer(containerName string) error {
	fmt.Println("stop container")
	ctx := context.Background()
	c := client.WithVersion("1.38")
	cli, err1 := client.NewClientWithOpts(c)
	if err1 != nil {
		fmt.Println("client create error")
		fmt.Println(err1)
		return err1
	}
	containerId, _ := getContainerId(containerName)
	err2 := cli.ContainerStop(ctx, containerId, nil)
	if err2 != nil {
		fmt.Println("container stop error")
		fmt.Println(err2)
		return err2
	}
	return nil
}

func getContainerId(containerName string) (string, error) {
	fmt.Println("get containerId")
	c := client.WithVersion("1.38")
	cli, err := client.NewClientWithOpts(c)
	if err != nil {
		fmt.Println("client create error")
		fmt.Println(err)
		return "", err
	}
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		fmt.Println("containerList error")
		fmt.Println(err)
		return "", err
	}
	for _, c := range containers {
		for _, name := range c.Names {
			if name == containerName {
				return c.ID, nil
			}
		}
	}
	return "", errors.New("container not found")
}

func removeContainer(containerName string) error {
	fmt.Println("remove container")
	containerID, err := getContainerId(containerName)
	if err != nil {
		fmt.Println("get container id errors")
		panic(err)
	}
	ctx := context.Background()
	c := client.WithVersion("1.38")
	cli, err1 := client.NewClientWithOpts(c)
	if err1 != nil {
		fmt.Println("client create error:")
		fmt.Println(err1)
		return err1
	}
	opts := types.ContainerRemoveOptions{RemoveVolumes: false, RemoveLinks: false, Force: false}
	err2 := cli.ContainerRemove(ctx, containerID, opts)
	if err2 != nil {
		fmt.Println("container remove error:")
		fmt.Println(err2)
		return err2
	}
	return nil
}

type Container struct {
	Image string
	Name  string
}

func containerRunHandler(w http.ResponseWriter, r *http.Request) {
	var c Container
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	id, err := runContainer(c.Image, c.Name)
	if err == nil {
		fmt.Fprintf(w, id)
	} else {
		http.Error(w, err.Error(), 500)
	}
}

func containerStopHandler(w http.ResponseWriter, r *http.Request) {
	var c Container
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	err2 := stopContainer(c.Name)
	if err2 == nil {
		fmt.Fprintf(w, "OK")
	} else {
		http.Error(w, err.Error(), 500)
	}
}

func containerRemoveHandler(w http.ResponseWriter, r *http.Request) {
	var c Container
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	pathSplit := strings.Split(r.URL.Path, "/")
	containerId := pathSplit[len(pathSplit)-1]
	err2 := removeContainer(containerId)
	if err2 == nil {
		fmt.Fprintf(w, "OK")
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
	}
}

func addHost(hostName string) error {
	fmt.Println("host add")
	if _, ok := hosts[hostName]; ok {
		return errors.New("host already exists")
	} else {
		hosts[hostName] = true
	}
	return nil
}

func hostAddHandler(w http.ResponseWriter, r *http.Request) {
	pathSplit := strings.Split(r.URL.Path, "/")
	hostName := pathSplit[len(pathSplit)-1]
	err := addHost(hostName)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
	}
}

func removeHost(hostName string) error {
	fmt.Println("remove host")
	if _, ok := hosts[hostName]; ok {
		delete(hosts, hostName)
	} else {
		return errors.New("host not found")
	}
	return nil
}

func hostRemoveHandler(w http.ResponseWriter, r *http.Request) {
	pathSplit := strings.Split(r.URL.Path, "/")
	hostName := pathSplit[len(pathSplit)-1]
	err := removeHost(hostName)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
	}
}

func listHosts() (string, error) {
	fmt.Println("list hosts")
	var l []string
	for k, _ := range hosts {
		l = append(l, k)
	}
	if len(l) > 0 {
		s := strings.Join(l, "\n")
		fmt.Println(s)
		return s, nil
	} else {
		return "", errors.New("hosts list is empty")
	}
}

func hostListHandler(w http.ResponseWriter, r *http.Request) {
	s, err := listHosts()
	if err != nil {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, s)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, s)
	}
}

func main() {
	fmt.Println("starting server")
	hosts = make(map[string]bool)
	http.HandleFunc("/storage_upload/", storageUploadHandler)
	http.HandleFunc("/storage_download/", storageDownloadHandler)
	http.HandleFunc("/storage_remove/", storageRemoveHandler)
	http.HandleFunc("/storage_list", storageListHandler)
	http.HandleFunc("/container_run", containerRunHandler)
	http.HandleFunc("/container_stop", containerStopHandler)
	http.HandleFunc("/container_list", containerListhandler)
	http.HandleFunc("/container_remove", containerRemoveHandler)
	http.HandleFunc("/host_add", hostAddHandler)
	http.HandleFunc("/host_remove", hostRemoveHandler)
	http.HandleFunc("/host_list", hostListHandler)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
