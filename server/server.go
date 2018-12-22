package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	etcdClient "go.etcd.io/etcd/client"
	"golang.org/x/net/context"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const addr = "0.0.0.0:8888"
const DataDir = "data"
const EtcdHost = "http://10.0.0.1:2379"

var LocalHostName string
var LocalIPPrefix = "10.0.0."
var LocalPort = "8888"

var EtcdClient etcdClient.Client

type Host struct {
	Name   string
	Disk   uint64
	Memory uint64
	Cores  uint64
}

type File struct {
	Name     string
	Host     string
	Size     uint64
	Replicas uint64
}

type Container struct {
	Image  string
	Name   string
	Disk   uint64
	Memory uint64
	Cores  uint64
	Host   string
	ID     string
	Cmd    []string
}

type Pod struct {
	Name       string
	Image      string
	Count      uint64
	Cores      uint64
	Memory     uint64
	Disk       uint64
	Cmd        []string
	Containers []Container
}

type Info struct {
	Storage    []File
	Hosts      []Host
	Pods       []Pod
	Containers []Container
}

func Fail(str string, err error, w http.ResponseWriter) {
	fmt.Println(str)
	fmt.Println(err.Error())
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(str))
	w.Write([]byte(err.Error()))
}

func EtcdCreateKey(name, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	kAPI := etcdClient.NewKeysAPI(EtcdClient)
	_, err := kAPI.Create(ctx, name, value)
	return err
}

func EtcdSetKey(name, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	kAPI := etcdClient.NewKeysAPI(EtcdClient)
	_, err := kAPI.Set(ctx, name, value, nil)
	return err
}

func EtcdDeleteKey(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	kAPI := etcdClient.NewKeysAPI(EtcdClient)
	_, err := kAPI.Delete(ctx, name, nil)
	return err
}

func EtcdGetKey(name string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	kAPI := etcdClient.NewKeysAPI(EtcdClient)
	resp, err := kAPI.Get(ctx, name, nil)
	if err != nil {
		return "", err
	}
	return resp.Node.Value, nil
}

func EtcdListDir(name string) (etcdClient.Nodes, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	kAPI := etcdClient.NewKeysAPI(EtcdClient)
	resp, err := kAPI.Get(ctx, name, nil)
	if err != nil {
		return nil, err
	}
	return resp.Node.Nodes, nil
}

func StorageUploadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("storage upload")
	PathSplit := strings.Split(r.URL.Path, "/")
	fileName := PathSplit[len(PathSplit)-1]
	dir, err := EtcdListDir("/rws/storage")
	if err != nil {
		Fail("EtcdListDir error", err, w)
	}
	for _, file := range dir {
		if file.Key == fileName {
			Fail("file already exists", errors.New("File already exists"), w)
		}
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Fail("request reading error", err, w)
		return
	}
	FileSize := len(body)
	FilePathName := DataDir + "/" + fileName
	files, err3 := StorageList()
	if err3 != nil {
		Fail("storage list error", err3, w)
		return
	}
	var x []File
	err4 := json.Unmarshal([]byte(files), &x)
	if err4 != nil {
		Fail("json.Unmarshal error", err4, w)
	}
	for _, file := range x {
		if file.Name == FilePathName {
			Fail("file already exists", errors.New("file already exists"), w)
			return
		}
	}
	di, err2 := disk.Usage("/")
	if err2 != nil {
		Fail("disk usage get error", err2, w)
		return
	}
	if di.Free > uint64(FileSize) {
		err3 := ioutil.WriteFile(FilePathName, []byte(body), 0644)
		if err3 != nil {
			Fail("file write error", err3, w)
			return
		}
		f := File{fileName, LocalHostName, uint64(FileSize), 1}
		fileBytes, err7 := json.Marshal(f)
		if err7 != nil {
			Fail("json.Marshal error", err7, w)
		}
		err8 := EtcdCreateKey("/rws/storage/"+fileName, string(fileBytes))
		if err8 != nil {
			Fail("EtcdCreateKey error", err8, w)
		}
		fmt.Println("file " + FilePathName + " uploaded")
		w.WriteHeader(http.StatusOK)
		return
	} else {
		hostsListString, err5 := ListHosts()
		if err5 != nil {
			Fail("ListHosts error", err5, w)
			return
		}
		var hostsList []Host
		err4 := json.Unmarshal([]byte(hostsListString), &hostsList)
		if err4 != nil {
			fmt.Println("JsonUnmarshal error")
			return
		}
		for _, host := range hostsList {
			url := "http://" + host.Name + "/host_info"
			body, err5 := http.Get(url)
			if err5 != nil {
				fmt.Println("get error: " + url)
				fmt.Println(body)
				continue
			}
			var ThatHost Host
			json.NewDecoder(body.Body).Decode(&ThatHost)
			if uint64(FileSize) < ThatHost.Disk {
				fmt.Println("uploading to " + host.Name)
				url := fmt.Sprintf("%s/storage_upload/%s", host.Name, FilePathName)
				dat, err6 := http.Post(url, "application/octet-stream", r.Body)
				if err6 != nil {
					fmt.Println("post error: " + url)
					fmt.Println(dat)
					continue
				}
			}
		}
	}
	fmt.Println("file upload error")

}

func StorageDownloadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("storage download")
	PathSplit := strings.Split(r.URL.Path, "/")
	fileName := PathSplit[len(PathSplit)-1]
	fmt.Println("filename: " + fileName)
	dir, err := EtcdListDir("/rws/storage")
	if err != nil {
		Fail("EtcdListDir error", err, w)
	}
	found := false
	for _, f := range dir {
		if f.Key == fileName {
			found = true
		}
	}
	if found == false {
		Fail("file not found", errors.New(""), w)
	}
	fileString, err9 := EtcdGetKey("/rws/storage/" + fileName)
	if err9 != nil {
		Fail("EtcdGetKey error", err9, w)
	}
	var file File
	err10 := json.Unmarshal([]byte(fileString), &file)
	if err10 != nil {
		Fail("json.Unmarshal error", err10, w)
	}
	if file.Host == LocalHostName {
		dat, err1 := ioutil.ReadFile(fmt.Sprintf("data/%s", fileName))
		if err1 != nil {
			Fail("file read error", err1, w)
			return
		}
		_, err2 := w.Write(dat)
		if err2 != nil {
			Fail("request write error", err2, w)
			return
		}
		fmt.Println("file " + fileName + " downloaded")
		w.WriteHeader(http.StatusOK)
		return
	} else {
		url := "http://" + file.Host + "/storage_download/" + file.Name
		body, err3 := http.Get(url)
		if err3 != nil {
			Fail("file get error", err3, w)
		}
		bodyBytes, err4 := ioutil.ReadAll(body.Body)
		if err4 != nil {
			fmt.Println(err4)
			Fail("body read error", err4, w)
		}
		w.Write(bodyBytes)
		w.WriteHeader(http.StatusOK)
		return
	}
}

func StorageRemoveHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("storage remove")
	PathSplit := strings.Split(r.URL.Path, "/")
	fileName := PathSplit[len(PathSplit)-1]
	fileString, err := EtcdGetKey("/rws/storage/" + fileName)
	if err != nil {
		Fail("EtcdGetKey error", err, w)
	}
	var file File
	err3 := json.Unmarshal([]byte(fileString), &file)
	if err3 != nil {
		Fail("json.Unmarshal error", err3, w)
	}
	if file.Host == LocalHostName {
		err := os.Remove("data/" + fileName)
		if err != nil {
			Fail("file remove error", err, w)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	} else {
		url := "http://" + file.Host + "/storage_remove/" + fileName
		_, err3 := http.Get(url)
		if err3 != nil {
			Fail("file remove get error", err3, w)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
	err2 := EtcdDeleteKey("/rws/storage/" + fileName)
	if err2 != nil {
		Fail("EtcdDeleteKey error", err2, w)
	}
	return
}

func StorageList() (string, error) {
	filesNodes, err := EtcdListDir("/rws/storage")
	if err != nil {
		return "", errors.New("EtcdListDir error")
	}
	var l []File
	for _, Key := range filesNodes {
		var x File
		err := json.Unmarshal([]byte(Key.Value), &x)
		if err != nil {
			fmt.Println("json unmarshal error")
			return "", err
		}
		l = append(l, x)
	}
	b, err2 := json.Marshal(l)
	if err2 != nil {
		return "", err2
	}
	return string(b), nil
}

func StorageListHandler(w http.ResponseWriter, _ *http.Request) {
	fmt.Println("storage list")
	s, err := StorageList()
	if err != nil {
		Fail("StorageList error", err, w)
		return
	}
	_, err2 := w.Write([]byte(s))
	if err2 != nil {
		Fail("request write error", err2, w)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func StorageFileSizeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("storage Size")
	PathSplit := strings.Split(r.URL.Path, "/")
	fileName := PathSplit[len(PathSplit)-1]
	fmt.Println("filename: " + fileName)
	found := false
	dir, err := EtcdListDir("/rws/storage")
	if err != nil {
		Fail("EtcdListDir error", err, w)
	}
	for _, Key := range dir {
		var file File
		err := json.Unmarshal([]byte(Key.Value), &file)
		if err != nil {
			fmt.Println("json unmarshal error")
			continue
		}
		if file.Name == fileName {
			found = true
		}
	}
	if found == false {
		Fail("file not found", err, w)
		return
	}
	var f File
	key, err := EtcdGetKey("/rws/storage/" + fileName)
	if err != nil {
		Fail("EtcdGetKey error", err, w)
	}
	err2 := json.Unmarshal([]byte(key), &f)
	if err2 != nil {
		Fail("json.Unmarshal error", err2, w)
	}
	w.Write([]byte(strconv.Itoa(int(f.Size))))
	w.WriteHeader(http.StatusOK)
	return
}

func ListLocalContainers() (string, error) {
	fmt.Println("list containers")
	c := client.WithVersion("1.38")
	cli, err := client.NewClientWithOpts(c)
	if err != nil {
		fmt.Println("client create error")
		fmt.Println(err)
		return "client create error", err
	}
	localContainers, err2 := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err2 != nil {
		fmt.Println("containerList error")
		fmt.Println(err2)
		return "containerList error", err2
	}
	if len(localContainers) == 0 {
		fmt.Println("there is no running containers on this Host")
		return "{}", nil
	}
	var allContainers []Container
	allContainersString, err3 := ListAllContainers()
	if err3 != nil {
		fmt.Println("ListAllContainers error")
		return "", err3
	}
	err4 := json.Unmarshal([]byte(allContainersString), &allContainers)
	if err4 != nil {
		fmt.Println("json unmarshal error")
		return "", err4
	}
	var returnContainers []Container
	for _, localContainer := range localContainers {
		for _, allContainer := range allContainers {
			if localContainer.ID == allContainer.ID {
				returnContainers = append(returnContainers, allContainer)
			}
		}
	}
	b, err3 := json.Marshal(returnContainers)
	if err3 != nil {
		fmt.Println("json marshal error")
		fmt.Println(err3)
		return "json marshal error", err3
	}
	return string(b), nil
}

func ListAllContainers() (string, error) {
	containersNodes, err := EtcdListDir("/rws/containers")
	if err != nil {
		return "", errors.New("EtcdListDir error")
	}
	var l []Container
	for _, Key := range containersNodes {
		var x Container
		err := json.Unmarshal([]byte(Key.Value), &x)
		if err != nil {
			fmt.Println("json.Unmarshal error")
			return "", err
		}
		l = append(l, x)
	}
	b, err2 := json.Marshal(l)
	if err2 != nil {
		return "", err2
	}
	return string(b), nil
}

func ContainerListHandler(w http.ResponseWriter, _ *http.Request) {
	s, err := ListAllContainers()
	if err != nil {
		Fail(s, err, w)
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(s))
}

func ContainerListLocalHandler(w http.ResponseWriter, _ *http.Request) {
	s, err := ListLocalContainers()
	if err != nil {
		Fail(s, err, w)
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(s))
}

func RunContainer(imageName, containerName string, cmd []string) (string, error) {
	fmt.Println("RunContainer")
	ctx := context.Background()
	c := client.WithVersion("1.38")
	cli, err := client.NewClientWithOpts(c)
	if err != nil {
		fmt.Println("RunContainer: client create error")
		fmt.Println(err)
		return "", err
	}
	out, err2 := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err2 != nil {
		fmt.Println("RunContainer: image pull error")
		fmt.Println(out)
		fmt.Println(err2)
		return "", err2
	}
	resp, err3 := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		Cmd:   cmd,
	}, nil, nil, containerName)
	if err3 != nil {
		fmt.Println("RunContainer: container create error")
		fmt.Println(resp)
		fmt.Println(err3)
		return "", err3
	}
	err4 := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err4 != nil {
		fmt.Println("RunContainer: container start error")
		fmt.Println(err4)
		return "", err4
	}
	cont := Container{Name: containerName, Image: imageName, Host: LocalHostName, ID:resp.ID}
	containerBytes, err5 := json.Marshal(cont)
	if err5 != nil {
		return "", err5
	}
	err6 := EtcdCreateKey("/rws/containers/"+containerName, string(containerBytes))
	if err6 != nil {
		fmt.Println("ContainerStopHandler:: EtcdCreateKey error ")
		fmt.Println(err6)
		to := 5 * time.Second
		err7 := cli.ContainerStop(ctx, resp.ID, &to)
		if err7 == nil {
			_ = cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{})
		}
		return "", nil

	}
	fmt.Println("RunContainer: container " + resp.ID + " running")
	return resp.ID, nil
}

func StopContainer(containerName string) error {
	fmt.Println("StopContainer")
	ctx := context.Background()
	c := client.WithVersion("1.38")
	cli, err1 := client.NewClientWithOpts(c)
	if err1 != nil {
		fmt.Println("StopContainer: NewClientWithOpts error")
		fmt.Println(err1)
		return err1
	}
	ContainerId, err := GetContainerId(containerName)
	if err != nil {
		fmt.Println("StopContainer: ContainerId error")
		fmt.Println(err)
		return err
	}
	fmt.Println(containerName)
	fmt.Println(ContainerId)
	err2 := cli.ContainerStop(ctx, ContainerId, nil)
	if err2 != nil {
		fmt.Println("StopContainer: ContainerStop error")
		fmt.Println(err2)
		return err2
	}
	return nil
}

func GetContainerId(containerName string) (string, error) {
	fmt.Println("GetContainerId")
	dir, err := EtcdListDir("/rws/containers/")
	if err != nil {
		return "", err
	}
	found := false
	for _, c := range dir {
		keySplit := strings.Split(c.Key, "/")
		keyName := keySplit[len(keySplit)-1]
		if keyName == containerName {
			found = true
		}
	}
	if found == false {
		return "", errors.New("container doesn't exist")
	}
	containerString, err2 := EtcdGetKey("/rws/containers/" + containerName)
	if err2 != nil {
		return "", err2
	}
	var cont Container
	err3 := json.Unmarshal([]byte(containerString), &cont)
	if err3 != nil {
		return "", err3
	}
	return cont.ID, nil
}

func RemoveContainer(containerName string) error {
	fmt.Println("remove container")
	ContainerID, err := GetContainerId(containerName)
	if err != nil {
		fmt.Println("get container id error")
		fmt.Println(err)
		return err
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
	err2 := cli.ContainerRemove(ctx, ContainerID, opts)
	if err2 != nil {
		fmt.Println("container remove error:")
		fmt.Println(err2)
		return err2
	}
	return nil
}

func ContainerRunHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("ContainerRunHandler")
	bodyBytes, err2 := ioutil.ReadAll(r.Body)
	if err2 != nil {
		Fail("ContainerRunHandler: response read error", err2, w)
	}
	var c Container
	err := json.Unmarshal(bodyBytes, &c)
	if err != nil {
		Fail("ContainerRunHandler: json.Unmarshal error", err, w)
		return
	}
	var ThatHost Host
	hostInfo, err := HostInfo()
	if err != nil {
		Fail("PodAddHandler: HostInfo error", err, w)
		return
	}
	err3 := json.Unmarshal([]byte(hostInfo), &ThatHost)
	if err3 != nil {
		Fail("PodAddHandler: json.Unmarshal error", err3, w)
	}
	if ThatHost.Disk >= c.Disk &&
		ThatHost.Cores >= c.Cores &&
		ThatHost.Memory >= c.Memory {
		id, err := RunContainer(c.Image, c.Name, c.Cmd)
		if err != nil {
			Fail("PodAddHandler: RunContainer error", err, w)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(id))
	} else {
		Fail("PodAddHandler: this host can't run this container", errors.New("can't run container on this host"), w)
		return
	}
}

func ContainerStopHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("ContainerStopHandler")
	bodyBytes, err2 := ioutil.ReadAll(r.Body)
	if err2 != nil {
		Fail("ContainerStopHandler: response read error", err2, w)
		return
	}
	var c Container
	err := json.Unmarshal(bodyBytes, &c)
	if err != nil {
		Fail("ContainerStopHandler: json.Unmarshal error", err, w)
		return
	}
	dir, err4 := EtcdListDir("/rws/containers")
	if err4 != nil {
		Fail("ContainerStopHandler: EtcdListDir error", err4, w)
		return
	}
	var cont Container
	found := false
	for _, k := range dir {
		keySplit := strings.Split(k.Key, "/")
		keyName := keySplit[len(keySplit)-1]
		if keyName == c.Name {
			found = true
			contString, err5 := EtcdGetKey(k.Key)
			if err5 != nil {
				Fail("ContainerStopHandler: EtcdGetKey error", err5, w)
				return
			}
			err6 := json.Unmarshal([]byte(contString), &cont)
			if err6 != nil {
				Fail("ContainerStopHandler: json.Unmarshal error", err6, w)
				return
			}
			break
		}
	}
	if found == false {
		Fail("ContainerStopHandler: container not found", errors.New(""), w)
		return
	}
	if cont.Host == LocalHostName {
		err2 := StopContainer(cont.Name)
		if err2 == nil {
			fmt.Fprintf(w, "OK")
		} else {
			Fail("ContainerStopHandler: stopContainer failure", err2, w)
			return
		}
	} else {
		url := "http://" + cont.Host + "/container_stop/" + cont.Name
		b, err2 := json.Marshal(cont)
		if err2 != nil {
			Fail("ContainerStopHandler: json Marshal error", err2, w)
			return
		}
		buf := bytes.NewBuffer(b)
		body, err3 := http.Post(url, "application/json", buf)
		if err3 == nil {
			if body.StatusCode != 200 {
				Fail("ContainerStopHandler: http.Post status code error: "+string(body.StatusCode), err3, w)
				return
			}
		} else {
			Fail("ContainerStopHandler: http.Post error", err3, w)
			return
		}
	}
	fmt.Println("ContainerStopHandler: container " + c.ID)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	return
}

func ContainerRemoveHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("ContainerRemoveHandler")
	bodyBytes, err2 := ioutil.ReadAll(r.Body)
	if err2 != nil {
		Fail("ContainerRemoveHandler: response read error", err2, w)
	}
	var c Container
	err := json.Unmarshal(bodyBytes, &c)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	dir, err4 := EtcdListDir("/rws/containers")
	if err4 != nil {
		Fail("ContainerRemoveHandler: EtcdListDir error", err4, w)
		return
	}
	var cont Container
	found := false
	for _, k := range dir {
		keySplit := strings.Split(k.Key, "/")
		keyName := keySplit[len(keySplit)-1]
		if keyName == c.Name {
			found = true
			contString, err5 := EtcdGetKey(k.Key)
			if err5 != nil {
				Fail("ContainerStopHandler: EtcdGetKey error", err5, w)
			}
			err6 := json.Unmarshal([]byte(contString), &cont)
			if err6 != nil {
				Fail("ContainerStopHandler: json.Unmarshal error", err6, w)
			}
		}
	}
	if found == false {
		Fail("ContainerRemoveHandler: container not found", errors.New(""), w)
		return
	}
	if c.Host == LocalHostName {
		err2 := RemoveContainer(c.Name)
		if err2 == nil {
			fmt.Fprintf(w, "OK")
		} else {
			Fail("ContainerStopHandler: stopContainer failure", err2, w)
			return
		}
	} else {
		url := "http://" + cont.Host + "/container_remove/" + cont.Name
		b, err2 := json.Marshal(c)
		if err2 != nil {
			fmt.Println(err2)
			panic("json Marshal error")
		}
		buf := bytes.NewBuffer(b)
		body, err3 := http.Post(url, "application/json", buf)
		if err3 == nil {
			if body.StatusCode != 200 {
				Fail("ContainerRemovepHandler: http.Post status code error: "+string(body.StatusCode), err3, w)
				return
			}
		} else {
			Fail("ContainerStopHandler: http.Post error", err3, w)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	return
}

func AddHost(hostName string) error {
	fmt.Println("Host add")
	dir, err := EtcdListDir("/rws/hosts")
	if err != nil {
		return err
	}
	found := false
	for _, node := range dir {
		if hostName == node.Key {
			found = true
			break
		}
	}
	if found == true {
		return errors.New("host already exists")
	}
	HostInfo, err3 := GetHostInfo(hostName)
	if err3 != nil {
		fmt.Println("AddHost: host info get error")
		return err3
	}
	b, err4 := json.Marshal(HostInfo)
	if err4 != nil {
		fmt.Println("AddHost: host info json marshal error")
		return err4
	}
	HostInfoString := string(b)
	if found == false {
		err2 := EtcdCreateKey("/rws/hosts/"+hostName, HostInfoString)
		if err2 != nil {
			return err2
		}
		fmt.Println("AddHost: host " + hostName + " added")
	} else {
		fmt.Println("AddHost: host already exists")
		return errors.New("host already exists")
	}
	return nil
}

func HostAddHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("HostAddHandler")
	var h Host
	err := json.NewDecoder(r.Body).Decode(&h)
	if err != nil {
		fmt.Println("HostAddHandler: " + err.Error())
		http.Error(w, err.Error(), 500)
		return
	}
	err2 := AddHost(h.Name)
	if err2 == nil {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "HostAddHandler: OK")
	} else {
		Fail("HostAddHandler: host create error", err2, w)
	}
}

func RemoveHost(hostName string) error {
	fmt.Println("remove Host")
	dir, err := EtcdListDir("/rws/hosts")
	if err != nil {
		return err
	}
	found := false
	for _, node := range dir {
		if hostName == node.Key {
			found = true
			break
		}
	}
	if found == false {
		return errors.New("host not found")
	} else {
		err2 := EtcdDeleteKey("/rws/hosts/" + hostName)
		if err2 != nil {
			return err2
		}
	}
	return nil
}

func HostRemoveHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("HostRemoveHandler")
	var h Host
	err := json.NewDecoder(r.Body).Decode(&h)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		fmt.Println(err.Error())
		return
	}
	err2 := RemoveHost(h.Name)
	if err2 == nil {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err2.Error())
	}
}

func ListHosts() (string, error) {
	fmt.Println("ListHosts")
	hosts, err := EtcdListDir("/rws/hosts")
	if err != nil {
		fmt.Println("EtcdListDir error")
		return "", err
	}
	var l []map[string]string
	for _, k := range hosts {
		h := map[string]string{k.Key: k.Value}
		l = append(l, h)
	}
	if len(l) > 0 {
		sm, err2 := json.Marshal(l)
		if err2 != nil {
			return "", err2
		}
		return string(sm), nil
	} else {
		return "{}", nil
	}
}

func HostListHandler(w http.ResponseWriter, _ *http.Request) {
	fmt.Println("HostListHandler")
	s, err := ListHosts()
	if err == nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(s))
	} else {
		Fail("ListHosts error", err, w)
	}
}

func HostInfo() (string, error) {
	ci, err1 := cpu.Info()
	if err1 != nil {
		return "", err1
	}
	mi, err2 := mem.VirtualMemory()
	if err2 != nil {
		return "", err2
	}
	di, err3 := disk.Usage("/")
	if err3 != nil {
		return "", err3
	}
	name, err4 := os.Hostname()
	if err4 != nil {
		name = "localhost"
	}
	var c = Host{name, di.Free, mi.Available, uint64(len(ci))}
	b, err := json.Marshal(c)
	return string(b), err
}

func HostInfoHandler(w http.ResponseWriter, _ *http.Request) {
	s, err := HostInfo()
	if err == nil {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, s)
	} else {
		Fail("HostInfo error", err, w)
	}
}

func GetFileSize(filename, host, port string) (uint64, error) {
	fmt.Println("GetFileSize")
	url := fmt.Sprintf("http://%s:%s/storage_file_size/%s", host, port, filename)
	body, err := http.Get(url)
	if err != nil {
		fmt.Println("get error")
		fmt.Println(body.Body)
		return 0, err
	}
	if body.StatusCode != 200 {
		fmt.Println(body.StatusCode)
		fmt.Println("status code error")
		return 0, err
	}
	b, err2 := ioutil.ReadAll(body.Body)
	if err2 != nil {
		fmt.Println(err2)
		fmt.Println("response read error")
		return 0, err2
	}
	i, err3 := strconv.Atoi(string(b))
	if err != nil {
		fmt.Println("atoi error")
		fmt.Println(err3)
		return 0, err3
	}
	return uint64(i), nil
}

func GetHostInfo(host string) (Host, error) {
	url := "http://" + host + "/host_info"
	body, err := http.Get(url)
	if err != nil {
		fmt.Println("GetHostInfo: get error")
		fmt.Println("GetHostInfo: " + url)
		fmt.Println(body)
		return Host{}, err
	}
	var ThatHost Host
	json.NewDecoder(body.Body).Decode(&ThatHost)
	return ThatHost, nil
}

func GetHostPods(host string) ([]Pod, error) {
	url := fmt.Sprintf("http://" + host + "/pod_list")
	body, err := http.Get(url)
	if err != nil {
		fmt.Println("get error")
		fmt.Println(body)
		return []Pod{}, err
	}
	var ThatHostPods []Pod
	json.NewDecoder(body.Body).Decode(&ThatHostPods)
	for _, pod := range ThatHostPods {
		ThatHostPods = append(ThatHostPods, pod)
	}
	return ThatHostPods, nil
}

func GetHostContainers(host string) ([]Container, error) {
	url := "http://" + host + "/container_list_local"
	body, err := http.Get(url)
	if err != nil {
		fmt.Println("get error")
		fmt.Println(body)
		return []Container{}, err
	}
	BodyBytes, err2 := ioutil.ReadAll(body.Body)
	if err2 != nil {
		fmt.Println("GetHostContainers error")
		fmt.Println(err2)
		return []Container{}, err2
	}
	if len(BodyBytes) == 0 {
		fmt.Println("no containers running on Host")
		return []Container{}, nil
	}
	var HostContainers []Container
	err3 := json.Unmarshal(BodyBytes, &HostContainers)
	if err3 != nil {
		fmt.Println("json unmarshal error")
		fmt.Println(err3)
		return []Container{}, err3
	}
	return HostContainers, nil
}

func IndexHandler(w http.ResponseWriter, _ *http.Request) {
	fmt.Println("IndexHandler")
	const tpl = `
<!DOCTYPE html>
<html>
	<head>
		<title>RWS</title>
	</head>
	<body>
		<h2>Storage</h2>
		<table>
			<tr>
				<th>Name</th>
				<th>Size</th>
				<th>Host</th>
			</tr>
			{{range .Storage}}
			<tr>
				<td>{{.Name}}</td>
				<td>{{.Size}}</td>
				<td>{{.Host}}</td>
			</tr>
			{{end}}
		</table>
		<h2>Hosts</h2>
		<table>
			<tr>
				<th>Name</th>
				<th>Cores</th>
				<th>Memory</th>
				<th>Disk</th>
			</tr>
			{{range .Hosts}}
			<tr>
				<td>{{.Name}}</td>
				<td>{{.Cores}}</td>
				<td>{{.Memory}}</td>
				<td>{{.Disk}}</td>
			</tr>
			{{end}}
		</table>
		<h2>Pods</h2>
		<table>
			<tr>
				<th>Name</th>
				<th>Image</th>
				<th>Count</th>
				<th>Cores</th>
				<th>Memory</th>
				<th>DISK</th>
			</tr>
			{{range .Pods}}
			<tr>
				<td>{{.Name}}</td>
				<td>{{.Image}}</td>
				<td>{{.Count}}</td>
				<td>{{.Cores}}</td>
				<td>{{.Memory}}</td>
				<td>{{.Disk}}</td>
			<tr>
			{{end}}
		</table>
		<h2>Containers</h2>
		<table>
			<tr>
				<th>Name</th>
				<th>Image</th>
				<th>Host</th>
				<th>Cores</th>
				<th>Memory</th>
				<th>Disk</th>
			</tr>
			{{range .Containers}}
			<tr>
				<td>{{.Name}}</td>
				<td>{{.Image}}</td>
				<td>{{.Host}}</td>

			<tr>
			{{end}}
		</table>
	</body>
</html>
`
	var info Info

	FilesString, err3 := StorageList()
	if err3 != nil {
		Fail("StorageListError", err3, w)
	}
	var f []File
	err := json.Unmarshal([]byte(FilesString), &f)
	if err != nil {
		Fail("json unmarshal error", err, w)
		return
	}
	for _, file := range f {
		info.Storage = append(info.Storage, file)
	}

	HostsString, err2 := ListHosts()
	if err2 != nil {
		Fail("ListHosts error", err2, w)
		return
	}
	var h []Host
	err4 := json.Unmarshal([]byte(HostsString), &h)
	if err4 != nil {
		Fail("json unmarshal error", err4, w)
		return
	}
	for _, host := range h {
		info.Hosts = append(info.Hosts, host)
	}

	PodsString, err5 := ListPods()
	if err5 != nil {
		Fail("ListPods error", err2, w)
		return
	}
	var p []Pod
	err6 := json.Unmarshal([]byte(PodsString), &h)
	if err6 != nil {
		Fail("json unmarshal error", err6, w)
		return
	}
	for _, pod := range p {
		info.Pods = append(info.Pods, pod)
	}

	ContainersString, err := ListAllContainers()
	if err != nil {
		Fail("ListAllContainers error", err, w)
	}
	var c []Container
	err7 := json.Unmarshal([]byte(ContainersString), &c)
	if err7 != nil {
		Fail("json unmarshal error", err7, w)
	}
	for _, cont := range c {
		info.Containers = append(info.Containers, cont)
	}

	t, err2 := template.New("index").Parse(tpl)
	if err2 != nil {
		fmt.Println("index html rendering error")
		fmt.Println(err2)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err2.Error()))
	}
	err = t.Execute(w, info)
	if err != nil {
		fmt.Println("template error")
		fmt.Println(err)
	}
	return
}

func PodAddHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("PodAddHandler")
	bodyBytes, err2 := ioutil.ReadAll(r.Body)
	if err2 != nil {
		Fail("PodAddHandler: response read error", err2, w)
	}
	var p Pod
	err := json.Unmarshal(bodyBytes, &p)
	if err != nil {
		Fail("PodAddHandler: json.Unmarshal error", err, w)
		return
	}
	dir, err := EtcdListDir("/rws/pods")
	if err != nil {
		Fail("PodAddHandler: EtcdListDir error", err, w)
		return
	}
	//var pod Pod
	found := false
	for _, k := range dir {
		keySplit := strings.Split(k.Key, "/")
		keyName := keySplit[len(keySplit)-1]
		if keyName == p.Name {
			found = true
			//contString, err5 := EtcdGetKey(k.Key)
			//if err5 != nil {
			//	Fail("PodAddHandler: EtcdGetKey error", err5, w)
			//}
			//err6 := json.Unmarshal([]byte(contString), &pod)
			//if err6 != nil {
			//	Fail("PodAddHandler: json.Unmarshal error", err6, w)
		}
	}
	if found == true {
		Fail("PodAddHandler: pod already exists", errors.New("pod already exists"), w)
		return
	}
	ps, err := json.Marshal(p)
	if err != nil {
		Fail("PodAddHandler: json.Marshal error", err, w)
		return
	}
	fmt.Println(string(ps))
	err = EtcdCreateKey("/rws/pods/"+p.Name, string(ps))
	if err != nil {
		Fail("PodAddHandler: EtcdCreateKey error", err, w)
		return
	}
	hostsDir, err := EtcdListDir("/rws/hosts/")
	if err != nil {
		Fail("PodAddHandler: EtcdListDir error", err, w)
	}
	var i uint64
	for _, h := range hostsDir {
		if i >= p.Count {
			break
		}
		keySplit := strings.Split(h.Key, "/")
		keyName := keySplit[len(keySplit)-1]
		url := fmt.Sprintf("http://" + keyName + "/host_info")
		resp, err := http.Get(url)
		if err != nil {
			Fail("PodAddHandler: http.get error", err, w)
			continue
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			Fail("PodAddHandler: ioutil.ReadAll error", err, w)
			return
		}
		var ThatHost Host
		err3 := json.Unmarshal(body, &ThatHost)
		if err3 != nil {
			Fail("PodAddHandler: json.Unmarshal error", err3, w)
			return
		}
		if ThatHost.Disk >= p.Disk &&
			ThatHost.Cores >= p.Cores &&
			ThatHost.Memory >= p.Memory {
			url := "http://" + h.Key + "/container_run"
			c := Container{p.Image, p.Name, p.Disk, p.Memory, p.Cores, h.Key, "", p.Cmd}
			b := new(bytes.Buffer)
			json.NewEncoder(b).Encode(c)
			resp, err1 := http.Post(url, "application/json", b)
			if err1 != nil {
				fmt.Println(err1)
				fmt.Println("request error")
				continue
			}
			if resp.StatusCode != 200 {
				fmt.Println("request status code error")
				fmt.Println(resp.StatusCode)
				fmt.Println(resp)
				continue
			}
			body, err2 := ioutil.ReadAll(resp.Body)
			if err2 != nil {
				fmt.Println("response read error")
				fmt.Println(err2)
				continue
			}
			c.ID = string(body)
			p.Containers = append(p.Containers, c)
			fmt.Println(body)
			i += 1
		}
	}
	s, err := json.Marshal(p)
	if err != nil {
		Fail("json.Marshal error", err, w)
	}
	err7 := EtcdSetKey("/rws/pods/"+p.Name, string(s))
	if err7 != nil {
		Fail("EtcdSetKey error", err7, w)
	}
	fmt.Println("all pod containers running")
	return
}

func PodStopHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("PodStopHandler")
	var p Pod
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	dir, err2 := EtcdListDir("/rws/hosts")
	if err2 != nil {
		Fail("EtcdListDir error", err2, w)
		return
	}
	for _, c := range p.Containers {
		for _, host := range dir {
			url := "http://" + host.Key + "/container_list"
			body, err := http.Get(url)
			if err != nil {
				fmt.Println("get error")
				fmt.Println("body")
				continue
			}
			if body.StatusCode != 200 {
				fmt.Println(body.StatusCode)
				fmt.Println("status code error")
				continue
			}
			b, err2 := ioutil.ReadAll(body.Body)
			if err2 != nil {
				fmt.Println(err2)
				fmt.Println("response read error")
			}
			var RemoteContainers []Container
			err3 := json.Unmarshal(b, &RemoteContainers)
			if err != nil {
				fmt.Println("json unmarshal error")
				fmt.Println(err3)
				continue
			}
			for _, RemoteContainer := range RemoteContainers {
				if RemoteContainer.Name == c.ID {
					b := new(bytes.Buffer)
					json.NewEncoder(b).Encode(c)
					url := "http://" + host.Key + "/container_stop"
					resp, err1 := http.Post(url, "application/json", b)
					if err1 != nil {
						fmt.Println(err1)
						fmt.Println("request error")
						continue
					}
					if resp.StatusCode != 200 {
						fmt.Println(resp.StatusCode)
						fmt.Println(resp)
						fmt.Println("request status code error")
						continue
					}
					_, err2 := ioutil.ReadAll(resp.Body)
					if err2 != nil {
						fmt.Println(err2)
						fmt.Println("response read error")
						continue
					}
				}
			}
		}
	}
	return
}

func ListPods() (string, error) {
	fmt.Println("ListPods")
	pods, err := EtcdListDir("/rws/pods")
	if err != nil {
		fmt.Println("EtcdListDir error")
		return "", err
	}
	var l []map[string]string
	for _, k := range pods {
		p := map[string]string{k.Key: k.Value}
		l = append(l, p)
	}
	if len(l) < 0 {
		return "{}", nil
	}
	sm, err2 := json.Marshal(l)
	if err2 != nil {
		return "", err2
	}
	return string(sm), nil
}

func PodListHandler(w http.ResponseWriter, _ *http.Request) {
	fmt.Println("PodListHandler")
	s, err := ListPods()
	if err != nil {
		Fail("PodsList error", err, w)
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(s))
	return
}

func PodRemoveHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("pod remove")
	var p Pod
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	err2 := EtcdDeleteKey("/rws/pods/" + p.Name)
	if err2 != nil {
		Fail("EtcdDeleteKey error", err2, w)
	}
	return
}

func scheduler() {
	for {
		fmt.Println("scheduler: run scheduler")
		dir, err := EtcdListDir("/rws/pods")
		if err != nil {
			fmt.Println("EtcdListDir error")
			fmt.Println(err)
			time.Sleep(60 * time.Second)
			continue
		}
		var pods []Pod
		for _, pod := range dir {
			var p Pod
			err2 := json.Unmarshal([]byte(pod.Value), &p)
			if err2 != nil {
				fmt.Println("scheduler: json unmarshal error")
				fmt.Println(err2)
				continue
			}
			pods = append(pods, p)
		}
		if len(pods) == 0 {
			fmt.Println("scheduler: no pods found")
			time.Sleep(60 * time.Second)
			continue
		}
		var hosts []Host
		dir, err2 := EtcdListDir("/rws/hosts")
		if err2 != nil {
			fmt.Println("scheduler: EtcdListDir error")
			fmt.Println(err2)
			continue
		}
		for _, host := range dir {
			var h Host
			err2 := json.Unmarshal([]byte(host.Value), &h)
			if err2 != nil {
				fmt.Println("scheduler: json unmarshal error")
				fmt.Println(err2)
				continue
			}
			hosts = append(hosts, h)
		}
		if len(hosts) == 0 {
			fmt.Println("scheduler: no hosts found")
			time.Sleep(60 * time.Second)
			continue
		}
		for _, p := range pods {
			fmt.Println("scheduler: Pod " + p.Name + " should have " + string(p.Count) + " containers")
			var foundContainers uint64
			for _, h := range hosts {
				hostRunningContainers, err4 := GetHostContainers(h.Name)
				if err4 != nil {
					fmt.Println("scheduler: getHostContainers error")
					fmt.Println(err4)
					continue
				}
				for _, podContainer := range p.Containers {
					for _, hostContainer := range hostRunningContainers {
						if podContainer.ID == hostContainer.ID {
							foundContainers += 1
							continue
						}
					}
				}
			}
			if foundContainers == p.Count {
				continue
			}
			var containersToRun uint64
			if foundContainers < p.Count {
				containersToRun = p.Count - foundContainers
			} else {
				containersToRun = foundContainers - p.Count
			}
			var i uint64
			for i = 0; i < containersToRun; i++ {
				for _, host := range hosts {
					id, err := RunContainer(p.Image, p.Name, p.Cmd)
					if err != nil {
						fmt.Println(err)
						continue
					}
					var c = Container{p.Image, p.Name, p.Disk, p.Memory, p.Cores, host.Name, id, p.Cmd}
					p.Containers = append(p.Containers, c)
				}
			}
			podMarshalled, err4 := json.Marshal(p)
			if err4 != nil {
				fmt.Println("scheduler: json.Marshal error")
				fmt.Println(err4)
				continue
			}
			err5 := EtcdSetKey("/rws/pods/"+p.Name, string(podMarshalled))
			if err5 != nil {
				fmt.Println("scheduler: EtcdSetKey error")
				fmt.Println(err5)
				continue
			}
		}
		time.Sleep(60 * time.Second)
	}
}
func main() {
	fmt.Println("starting server")
	hostNameBytes, err1 := ioutil.ReadFile("/etc/hostname")
	if err1 != nil {
		fmt.Println(err1)
		panic("/etc/hostname reading error")
	}
	LocalHostNumber := string(hostNameBytes[len(hostNameBytes)-2])
	LocalHostName = LocalIPPrefix + LocalHostNumber + ":" + LocalPort
	etcdCfg := etcdClient.Config{
		Endpoints: []string{EtcdHost},
		Transport: etcdClient.DefaultTransport,
	}
	var err error
	EtcdClient, err = etcdClient.New(etcdCfg)
	if err != nil {
		fmt.Println(err)
		panic("etcd client initialization error")
	}
	go scheduler()
	http.HandleFunc("/storage_upload/", StorageUploadHandler)
	http.HandleFunc("/storage_download/", StorageDownloadHandler)
	http.HandleFunc("/storage_remove/", StorageRemoveHandler)
	http.HandleFunc("/storage_list", StorageListHandler)
	http.HandleFunc("/storage_file_size/", StorageFileSizeHandler)
	http.HandleFunc("/container_run", ContainerRunHandler)
	http.HandleFunc("/container_stop", ContainerStopHandler)
	http.HandleFunc("/container_list", ContainerListHandler)
	http.HandleFunc("/container_list_local", ContainerListLocalHandler)
	http.HandleFunc("/container_remove", ContainerRemoveHandler)
	http.HandleFunc("/pod_add", PodAddHandler)
	http.HandleFunc("/pod_stop", PodStopHandler)
	http.HandleFunc("/pod_list", PodListHandler)
	http.HandleFunc("/pod_remove", PodRemoveHandler)
	http.HandleFunc("/host_add", HostAddHandler)
	http.HandleFunc("/host_remove", HostRemoveHandler)
	http.HandleFunc("/host_list", HostListHandler)
	http.HandleFunc("/host_info", HostInfoHandler)
	http.HandleFunc("/", IndexHandler)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
