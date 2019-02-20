package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dchest/uniuri"
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
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const addr = "0.0.0.0:8888"
const DataDir = "data"
const EtcdHost = "http://etcd:2379"

const LocalHostName = "localhost:8888"

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

func getFileNameFromPath(p string) string {
	PathSplit := strings.Split(p, "/")
	fileName := PathSplit[len(PathSplit)-1]
	return fileName
}

func fail(str string, err error, w http.ResponseWriter) {
	log.Println(1, str)
	log.Println(1, err.Error())
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(str))
	w.Write([]byte(err.Error()))
}

func etcdCreateKey(name, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	kAPI := etcdClient.NewKeysAPI(EtcdClient)
	_, err := kAPI.Create(ctx, name, value)
	return err
}

func etcdSetKey(name, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	kAPI := etcdClient.NewKeysAPI(EtcdClient)
	_, err := kAPI.Set(ctx, name, value, nil)
	return err
}

func etcdDeleteKey(name string) error {
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
	log.Println(1, "StorageUploadHandler")
	fileName := getFileNameFromPath(r.URL.Path)
	log.Println(1, "StorageUploadHandler: "+fileName)
	dir, err := EtcdListDir("/rws/storage")
	if err != nil {
		fail("StorageUploadHandler: EtcdListDir error", err, w)
	}
	for _, file := range dir {
		keyName := getFileNameFromPath(file.Key)
		if keyName == fileName {
			fail("StorageUploadHandler: file already exists", errors.New("file already exists"), w)
			return
		}
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fail("StorageUploadHandler: request reading error", err, w)
		return
	}
	FileSize := len(body)
	FilePathName := DataDir + "/" + fileName
	di, err2 := disk.Usage("/")
	if err2 != nil {
		fail("StorageUploadHandler: disk usage get error", err2, w)
		return
	}
	if di.Free > uint64(FileSize) {
		err3 := ioutil.WriteFile(FilePathName, []byte(body), 0644)
		if err3 != nil {
			fail("StorageUploadHandler: file write error", err3, w)
			return
		}
		f := File{fileName, LocalHostName, uint64(FileSize), 1}
		fileBytes, err7 := json.Marshal(f)
		if err7 != nil {
			fail("StorageUploadHandler: json.Marshal error", err7, w)
			return
		}
		err8 := etcdCreateKey("/rws/storage/"+fileName, string(fileBytes))
		if err8 != nil {
			fail("StorageUploadHandler: etcdCreateKey error", err8, w)
			return
		}
		log.Println(1, "StorageUploadHandler: file "+FilePathName+" uploaded")
		w.WriteHeader(http.StatusOK)
		return
	} else {
		hostsListString, err5 := ListHosts()
		if err5 != nil {
			fail("StorageUploadHandler: ListHosts error", err5, w)
			return
		}
		var hostsList []Host
		err4 := json.Unmarshal([]byte(hostsListString), &hostsList)
		if err4 != nil {
			fail("StorageUploadHandler: JsonUnmarshal error", err4, w)
			return
		}
		for _, host := range hostsList {
			url := "http://" + host.Name + "/host_info"
			body2, err5 := http.Get(url)
			if err5 != nil {
				fail("StorageUploadHandler: get error", err5, w)
				continue
			}
			bodyBytes, err := ioutil.ReadAll(body2.Body)
			if err != nil {
				fail("StorageUploadHandler: request reading error", err, w)
				return
			}
			var thatHost Host
			err6 := json.Unmarshal([]byte(bodyBytes), &thatHost)
			if err6 != nil {
				fail("StorageUploadHandler: json.Unmarshal error", err6, w)
				return
			}
			if uint64(FileSize) < thatHost.Disk {
				log.Println(1, "StorageUploadHandler: uploading to "+host.Name)
				url := fmt.Sprintf("%s/storage_upload/%s", host.Name, FilePathName)
				dat, err6 := http.Post(url, "application/octet-stream", r.Body)
				if err6 != nil {
					log.Println("StorageUploadHandle: post error: " + url)
					log.Println(dat)
					continue
				}
				log.Println(1, "StorageUploadHandler: "+fileName+" uploaded")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
				return
			} else {
				log.Println(1, "StorageUploadHandler: not enough free space on "+host.Name)
				continue
			}
		}
		log.Println(1, "StorageUploadHandler: unable to upload file "+fileName)
		log.Println(1, http.StatusInternalServerError)
		w.Write([]byte("StorageUploadHandler: unable to upload file " + fileName))
		return
	}
}

func StorageDownloadHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(1, "StorageDownloadHandler")
	fileName := getFileNameFromPath(r.URL.Path)
	log.Println(1, "StorageDownloadHandler: "+fileName)
	dir, err := EtcdListDir("/rws/storage")
	if err != nil {
		fail("StorageDownloadHandler: EtcdListDir error", err, w)
		return
	}
	found := false
	for _, file := range dir {
		keyName := getFileNameFromPath(file.Key)
		if keyName == fileName {
			found = true
			break
		}
	}
	if found == false {
		fail("StorageDownloadHandler: file not found", errors.New("file not found"), w)
		return
	}
	fileString, err9 := EtcdGetKey("/rws/storage/" + fileName)
	if err9 != nil {
		fail("StorageDownloadHandler: EtcdGetKey error", err9, w)
		return
	}
	var file File
	err10 := json.Unmarshal([]byte(fileString), &file)
	if err10 != nil {
		fail("StorageDownloadHandler: json.Unmarshal error", err10, w)
		return
	}
	if file.Host == LocalHostName {
		dat, err1 := ioutil.ReadFile("data/" + fileName)
		if err1 != nil {
			fail("StorageDownloadHandler: file read error", err1, w)
			return
		}
		_, err2 := w.Write(dat)
		if err2 != nil {
			fail("StorageDownloadHandler: request write error", err2, w)
			return
		}
		w.WriteHeader(http.StatusOK)
		log.Println(1, "StorageDownloadHandler: file "+fileName+" downloaded")
		return
	} else {
		url := "http://" + file.Host + "/storage_download/" + file.Name
		body, err3 := http.Get(url)
		if err3 != nil {
			fail("StorageDownloadHandler: file get error", err3, w)
			return
		}
		bodyBytes, err4 := ioutil.ReadAll(body.Body)
		if err4 != nil {
			fail("StorageDownloadHandler: body read error", err4, w)
			return
		}
		_, err6 := w.Write(bodyBytes)
		if err6 != nil {
			fail("StorageDownloadHandler: request write error", err6, w)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
}

func StorageRemoveHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(1, "StorageRemoveHandler")
	fileName := getFileNameFromPath(r.URL.Path)
	log.Println(1, "StorageRemoveHandler: "+fileName)
	dir, err := EtcdListDir("/rws/storage")
	if err != nil {
		fail("StorageDownloadHandler: EtcdListDir error", err, w)
		return
	}
	found := false
	for _, f := range dir {
		keyName := getFileNameFromPath(f.Key)
		if keyName == fileName {
			found = true
			break
		}
	}
	if found == false {
		fail("StorageDownloadHandler: file not found", errors.New("file not found"), w)
		return
	}
	fileString, err := EtcdGetKey("/rws/storage/" + fileName)
	if err != nil {
		fail("StorageRemoveHandler: EtcdGetKey error", err, w)
		return
	}
	var file File
	err3 := json.Unmarshal([]byte(fileString), &file)
	if err3 != nil {
		fail("StorageRemoveHandler: json.Unmarshal error", err3, w)
		return
	}
	if file.Host == LocalHostName {
		err := os.Remove("data/" + fileName)
		if err != nil {
			fail("StorageRemoveHandler: file remove error", err, w)
			return
		}
		log.Println(1, "StorageRemoveHandler: file "+fileName+" removed locally")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	} else {
		url := "http://" + file.Host + "/storage_remove/" + fileName
		resp, err3 := http.Get(url)
		if err3 != nil {
			fail("StorageRemoveHandler: file remove get error", err3, w)
			return
		}
		if resp.StatusCode != http.StatusOK {
			fail("StorageRemoveHandler: file remove get error", err3, w)
			return
		}
		log.Println(1, "StorageRemoveHandler: file "+fileName+" removed from host "+file.Host)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}
	err2 := etcdDeleteKey("/rws/storage/" + fileName)
	if err2 != nil {
		fail("StorageRemoveHandler: etcdDeleteKey error", err2, w)
	}
	log.Println(1, "StorageDeleteHandler: "+fileName+" deleted")
	return
}

func StorageList() (string, error) {
	filesNodes, err := EtcdListDir("/rws/storage")
	if err != nil {
		return "", errors.New("StorageList: EtcdListDir error")
	}
	var l []File
	for _, Key := range filesNodes {
		var x File
		err := json.Unmarshal([]byte(Key.Value), &x)
		if err != nil {
			log.Println(1, "StorageList: json unmarshal error")
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
	log.Println(1, "StorageListHandler")
	s, err := StorageList()
	if err != nil {
		fail("StorageListHandler: StorageList error", err, w)
		return
	}
	_, err2 := w.Write([]byte(s))
	if err2 != nil {
		fail("StorageListHandler: request write error", err2, w)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func StorageFileSizeHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(1, "StorageFileSizeHandler: storage file size")
	fileName := getFileNameFromPath(r.URL.Path)
	log.Println(1, "StorageFileSizeHandler: "+fileName)
	found := false
	dir, err := EtcdListDir("/rws/storage")
	if err != nil {
		fail("StorageFileSizeHandler: EtcdListDir error", err, w)
		return
	}
	for _, Key := range dir {
		keyName := getFileNameFromPath(Key.Key)
		if keyName == fileName {
			found = true
		}
	}
	if found == false {
		fail("StorageFileSizeHandler: file not found", err, w)
		return
	}
	var f File
	key, err := EtcdGetKey("/rws/storage/" + fileName)
	if err != nil {
		fail("StorageFileSizeHandler: EtcdGetKey error", err, w)
	}
	err2 := json.Unmarshal([]byte(key), &f)
	if err2 != nil {
		fail("StorageFileSizeHandler: json.Unmarshal error", err2, w)
	}
	w.Write([]byte(strconv.Itoa(int(f.Size))))
	w.WriteHeader(http.StatusOK)
	return
}

func ListLocalContainers() (string, error) {
	log.Println(1, "ListLocalContainers")
	c := client.WithVersion("1.38")
	cli, err := client.NewClientWithOpts(c)
	if err != nil {
		log.Println(1, "ListLocalContainers: client create error")
		log.Println(1, err)
		return "client create error", err
	}
	localContainers, err2 := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err2 != nil {
		log.Println(1, "ListLocalContainers: containerList error")
		log.Println(1, err2)
		return "containerList error", err2
	}
	if len(localContainers) == 0 {
		log.Println(1, "ListLocalContainers: there is no running containers on this Host")
		return "{}", nil
	}
	var allContainers []Container
	allContainersString, err3 := ListAllContainers()
	if err3 != nil {
		log.Println(1, "ListLocalContainers: ListAllContainers error")
		return "", err3
	}
	err4 := json.Unmarshal([]byte(allContainersString), &allContainers)
	if err4 != nil {
		log.Println(1, "ListLocalContainers: json unmarshal error")
		return "", err4
	}
	var returnContainers []Container
	for _, localContainer := range localContainers {
		for _, allContainer := range allContainers {
			if localContainer.ID == allContainer.ID {
				returnContainers = append(returnContainers, allContainer)
				break
			}
		}
	}
	b, err3 := json.Marshal(returnContainers)
	if err3 != nil {
		log.Println(1, "json marshal error")
		log.Println(1, err3)
		return "json marshal error", err3
	}
	return string(b), nil
}

func ListAllContainers() (string, error) {
	log.Println(1, "ListAllContainers")
	containersNodes, err := EtcdListDir("/rws/containers")
	if err != nil {
		log.Println(1, "ListAllContainers: EtcdListDir error")
		return "", err
	}
	var l []Container
	for _, Key := range containersNodes {
		var x Container
		err := json.Unmarshal([]byte(Key.Value), &x)
		if err != nil {
			log.Println(1, "ListAllContainers: json.Unmarshal error")
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
		fail(s, err, w)
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(s))
}

func ContainerListLocalHandler(w http.ResponseWriter, _ *http.Request) {
	s, err := ListLocalContainers()
	if err != nil {
		fail(s, err, w)
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(s))
}

func RunContainer(imageName, containerName string, cmd []string) (string, error) {
	log.Println(1, "RunContainer")
	ctx := context.Background()
	c := client.WithVersion("1.38")
	cli, err := client.NewClientWithOpts(c)
	if err != nil {
		log.Println(1, "RunContainer: client create error")
		log.Println(1, err)
		return "", err
	}
	out, err2 := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err2 != nil {
		log.Println(1, "RunContainer: image pull error")
		log.Println(1, out)
		log.Println(1, err2)
		return "", err2
	}
	resp, err3 := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		Cmd:   cmd,
	}, nil, nil, containerName)
	if err3 != nil {
		log.Println(1, "RunContainer: container create error")
		log.Println(1, resp)
		log.Println(1, err3)
		return "", err3
	}
	err4 := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err4 != nil {
		log.Println(1, "RunContainer: container start error")
		log.Println(1, err4)
		return "", err4
	}
	cont := Container{Name: containerName, Image: imageName, Host: LocalHostName, ID: resp.ID}
	containerBytes, err5 := json.Marshal(cont)
	if err5 != nil {
		return "", err5
	}
	err6 := etcdCreateKey("/rws/containers/"+containerName, string(containerBytes))
	if err6 != nil {
		log.Println(1, "RunContainer: etcdCreateKey error")
		log.Println(1, err6)
		return "", err6
	}
	log.Println(1, "RunContainer: container "+resp.ID+" running")
	return resp.ID, nil
}

func StopContainer(containerName string) error {
	log.Println(1, "StopContainer")
	ctx := context.Background()
	c := client.WithVersion("1.38")
	cli, err1 := client.NewClientWithOpts(c)
	if err1 != nil {
		log.Println(1, "StopContainer: NewClientWithOpts error")
		log.Println(1, err1)
		return err1
	}
	ContainerId, err := GetContainerId(containerName)
	if err != nil {
		log.Println(1, "StopContainer: ContainerId error")
		log.Println(1, err)
		return err
	}
	err2 := cli.ContainerStop(ctx, ContainerId, nil)
	if err2 != nil {
		log.Println(1, "StopContainer: ContainerStop error")
		log.Println(1, err2)
		return err2
	}
	return nil
}

func GetContainerId(containerName string) (string, error) {
	log.Println(1, "GetContainerId")
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
			break
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
	log.Println(1, "RemoveContainer")
	ContainerID, err := GetContainerId(containerName)
	if err != nil {
		log.Println(1, "RemoveContainer: GetContainerId error")
		log.Println(1, err)
		return err
	}
	ctx := context.Background()
	c := client.WithVersion("1.38")
	cli, err1 := client.NewClientWithOpts(c)
	if err1 != nil {
		log.Println(1, "client create error:")
		log.Println(1, err1)
		return err1
	}
	opts := types.ContainerRemoveOptions{RemoveVolumes: false, RemoveLinks: false, Force: false}
	err2 := cli.ContainerRemove(ctx, ContainerID, opts)
	if err2 != nil {
		log.Println(1, "RemoveContainer: container remove error:")
		log.Println(1, err2)
		return err2
	}
	log.Println(1, "RemoveContainer: Container "+ContainerID+" removed")
	return nil
}

func ContainerRunHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(1, "ContainerRunHandler")
	bodyBytes, err2 := ioutil.ReadAll(r.Body)
	if err2 != nil {
		fail("ContainerRunHandler: response read error", err2, w)
	}
	var c Container
	err := json.Unmarshal(bodyBytes, &c)
	if err != nil {
		fail("ContainerRunHandler: json.Unmarshal error", err, w)
		return
	}
	var ThatHost Host
	hostInfo, err := HostInfo()
	if err != nil {
		fail("ContainerRunHandler: HostInfo error", err, w)
		return
	}
	err3 := json.Unmarshal([]byte(hostInfo), &ThatHost)
	if err3 != nil {
		fail("ContainerRunHandler: json.Unmarshal error", err3, w)
	}
	if ThatHost.Disk >= c.Disk &&
		ThatHost.Cores >= c.Cores &&
		ThatHost.Memory >= c.Memory {
		id, err := RunContainer(c.Image, c.Name, c.Cmd)
		if err != nil {
			fail("ContainerRunHandler: RunContainer error", err, w)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(id))
	} else {
		fail("ContainerRunHandler: this host can't run this container", errors.New("can't run container on this host"), w)
		return
	}
}

func ContainerStopHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(1, "ContainerStopHandler")
	bodyBytes, err2 := ioutil.ReadAll(r.Body)
	if err2 != nil {
		fail("ContainerStopHandler: response read error", err2, w)
		return
	}
	var c Container
	err := json.Unmarshal(bodyBytes, &c)
	if err != nil {
		fail("ContainerStopHandler: json.Unmarshal error", err, w)
		return
	}
	dir, err4 := EtcdListDir("/rws/containers")
	if err4 != nil {
		fail("ContainerStopHandler: EtcdListDir error", err4, w)
		return
	}
	var cont Container
	found := false
	for _, k := range dir {
		keyName := getFileNameFromPath(k.Key)
		if keyName == c.Name {
			found = true
			contString, err5 := EtcdGetKey(k.Key)
			if err5 != nil {
				fail("ContainerStopHandler: EtcdGetKey error", err5, w)
				return
			}
			err6 := json.Unmarshal([]byte(contString), &cont)
			if err6 != nil {
				fail("ContainerStopHandler: json.Unmarshal error", err6, w)
				return
			}
			break
		}
	}
	if found == false {
		fail("ContainerStopHandler: container not found", errors.New(""), w)
		return
	}
	if cont.Host == LocalHostName {
		err2 := StopContainer(cont.Name)
		if err2 != nil {
			fail("ContainerStopHandler: stopContainer failure", err2, w)
			return
		}
	} else {
		url := "http://" + cont.Host + "/container_stop/" + cont.Name
		b, err2 := json.Marshal(cont)
		if err2 != nil {
			fail("ContainerStopHandler: json Marshal error", err2, w)
			return
		}
		buf := bytes.NewBuffer(b)
		body, err3 := http.Post(url, "application/json", buf)
		if err3 == nil {
			if body.StatusCode != 200 {
				fail("ContainerStopHandler: http.Post status code error: "+string(body.StatusCode), err3, w)
				return
			}
		} else {
			fail("ContainerStopHandler: http.Post error", err3, w)
			return
		}
	}
	log.Println(1, "ContainerStopHandler: container "+c.ID+" stopped")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	return
}

func ContainerRemoveHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(1, "ContainerRemoveHandler")
	bodyBytes, err2 := ioutil.ReadAll(r.Body)
	if err2 != nil {
		fail("ContainerRemoveHandler: response read error", err2, w)
	}
	var c Container
	err := json.Unmarshal(bodyBytes, &c)
	if err != nil {
		fail("ContainerRemovehandler: json.Unmarshal error", err, w)
		return
	}
	dir, err4 := EtcdListDir("/rws/containers")
	if err4 != nil {
		fail("ContainerRemoveHandler: EtcdListDir error", err4, w)
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
				fail("ContainerRemoveHandler: EtcdGetKey error", err5, w)
			}
			err6 := json.Unmarshal([]byte(contString), &cont)
			if err6 != nil {
				fail("ContainerRemove¡Handler: json.Unmarshal error", err6, w)
			}
		}
	}
	if found == false {
		fail("ContainerRemoveHandler: container not found", errors.New(""), w)
		return
	}
	if c.Host == LocalHostName {
		err2 := RemoveContainer(c.Name)
		if err2 == nil {
			fmt.Fprintf(w, "OK")
		} else {
			fail("ContainerStopHandler: stopContainer failure", err2, w)
			return
		}
	} else {
		url := "http://" + cont.Host + "/container_remove/" + cont.Name
		b, err2 := json.Marshal(c)
		if err2 != nil {
			log.Println(1, err2)
			panic("json Marshal error")
		}
		buf := bytes.NewBuffer(b)
		body, err3 := http.Post(url, "application/json", buf)
		if err3 == nil {
			if body.StatusCode != 200 {
				fail("ContainerRemovepHandler: http.Post status code error: "+string(body.StatusCode), err3, w)
				return
			}
		} else {
			fail("ContainerStopHandler: http.Post error", err3, w)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	return
}

func AddHost(hostName string) error {
	log.Println(1, "Host add")
	dir, err := EtcdListDir("/rws/hosts")
	if err != nil {
		return err
	}
	found := false
	for _, node := range dir {
		keySplit := strings.Split(node.Key, "/")
		keyName := keySplit[len(keySplit)-1]
		if hostName == keyName {
			found = true
			break
		}
	}
	if found == true {
		return errors.New("host already exists")
	}
	HostInfo, err3 := GetHostInfo(hostName)
	if err3 != nil {
		log.Println(1, "AddHost: host info get error")
		return err3
	}
	b, err4 := json.Marshal(HostInfo)
	if err4 != nil {
		log.Println(1, "AddHost: host info json marshal error")
		return err4
	}
	HostInfoString := string(b)
	if found == false {
		err2 := etcdCreateKey("/rws/hosts/"+hostName, HostInfoString)
		if err2 != nil {
			return err2
		}
		log.Println(1, "AddHost: host "+hostName+" added")
	} else {
		log.Println(1, "AddHost: host already exists")
		return errors.New("host already exists")
	}
	return nil
}

func HostAddHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(1, "HostAddHandler")
	var h Host
	err := json.NewDecoder(r.Body).Decode(&h)
	if err != nil {
		log.Println(1, "HostAddHandler: "+err.Error())
		http.Error(w, err.Error(), 500)
		return
	}
	err2 := AddHost(h.Name)
	if err2 == nil {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "HostAddHandler: OK")
	} else {
		fail("HostAddHandler: host create error", err2, w)
	}
}

func RemoveHost(hostName string) error {
	log.Println(1, "RemoveHost")
	dir, err := EtcdListDir("/rws/hosts")
	if err != nil {
		return err
	}
	found := false
	for _, node := range dir {
		keySplit := strings.Split(node.Key, "/")
		keyName := keySplit[len(keySplit)-1]
		if hostName == keyName {
			found = true
			break
		}
	}
	if found == false {
		return errors.New("RemoveHost: host not found")
	} else {
		err2 := etcdDeleteKey("/rws/hosts/" + hostName)
		if err2 != nil {
			return err2
		}
	}
	return nil
}

func HostRemoveHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(1, "HostRemoveHandler")
	bodyBytes, err3 := ioutil.ReadAll(r.Body)
	if err3 != nil {
		fail("HostRemoveHandler: response read error", err3, w)
	}
	var h Host
	err := json.Unmarshal(bodyBytes, &h)
	if err != nil {
		fail("HostRemoveHandler: json.Unmarshal error", err, w)
		return
	}
	err2 := RemoveHost(h.Name)
	if err2 == nil {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	} else {
		fail("HostRemoveHandler: RemoveHost error", err2, w)
		return
	}
}

func ListHosts() (string, error) {
	log.Println(1, "ListHosts")
	hosts, err := EtcdListDir("/rws/hosts")
	if err != nil {
		log.Println(1, "EtcdListDir error")
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
	log.Println(1, "HostListHandler")
	s, err := ListHosts()
	if err == nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(s))
	} else {
		fail("ListHosts error", err, w)
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
	name := LocalHostName
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
		fail("HostInfo error", err, w)
	}
}

func GetFileSize(filename, host, port string) (uint64, error) {
	log.Println(1, "GetFileSize")
	url := fmt.Sprintf("http://%s:%s/storage_file_size/%s", host, port, filename)
	body, err := http.Get(url)
	if err != nil {
		log.Println(1, "get error")
		log.Println(1, body.Body)
		return 0, err
	}
	if body.StatusCode != 200 {
		log.Println(1, body.StatusCode)
		log.Println(1, "status code error")
		return 0, err
	}
	b, err2 := ioutil.ReadAll(body.Body)
	if err2 != nil {
		log.Println(1, err2)
		log.Println(1, "response read error")
		return 0, err2
	}
	i, err3 := strconv.Atoi(string(b))
	if err != nil {
		log.Println(1, "atoi error")
		log.Println(1, err3)
		return 0, err3
	}
	return uint64(i), nil
}

func GetHostInfo(host string) (Host, error) {
	url := "http://" + host + "/host_info"
	body, err := http.Get(url)
	if err != nil {
		log.Println(1, "GetHostInfo: get error")
		log.Println(1, "GetHostInfo: "+url)
		log.Println(1, body)
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
		log.Println(1, "get error")
		log.Println(1, body)
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
		log.Println(1, "get error")
		log.Println(1, body)
		return []Container{}, err
	}
	BodyBytes, err2 := ioutil.ReadAll(body.Body)
	if err2 != nil {
		log.Println(1, "GetHostContainers error")
		log.Println(1, err2)
		return []Container{}, err2
	}
	if len(BodyBytes) == 0 {
		log.Println(1, "no containers running on Host")
		return []Container{}, nil
	}
	var HostContainers []Container
	err3 := json.Unmarshal(BodyBytes, &HostContainers)
	if err3 != nil {
		log.Println(1, "json unmarshal error")
		log.Println(1, err3)
		return []Container{}, err3
	}
	return HostContainers, nil
}

func IndexHandler(w http.ResponseWriter, _ *http.Request) {
	log.Println(1, "IndexHandler")
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
		fail("StorageListError", err3, w)
	}
	var f []File
	err := json.Unmarshal([]byte(FilesString), &f)
	if err != nil {
		fail("json unmarshal error", err, w)
		return
	}
	for _, file := range f {
		info.Storage = append(info.Storage, file)
	}

	HostsString, err2 := ListHosts()
	if err2 != nil {
		fail("ListHosts error", err2, w)
		return
	}
	var h []Host
	err4 := json.Unmarshal([]byte(HostsString), &h)
	if err4 != nil {
		fail("json unmarshal error", err4, w)
		return
	}
	for _, host := range h {
		info.Hosts = append(info.Hosts, host)
	}

	PodsString, err5 := ListPods()
	if err5 != nil {
		fail("ListPods error", err2, w)
		return
	}
	var p []Pod
	err6 := json.Unmarshal([]byte(PodsString), &h)
	if err6 != nil {
		fail("json unmarshal error", err6, w)
		return
	}
	for _, pod := range p {
		info.Pods = append(info.Pods, pod)
	}

	ContainersString, err := ListAllContainers()
	if err != nil {
		fail("ListAllContainers error", err, w)
	}
	var c []Container
	err7 := json.Unmarshal([]byte(ContainersString), &c)
	if err7 != nil {
		fail("json unmarshal error", err7, w)
	}
	for _, cont := range c {
		info.Containers = append(info.Containers, cont)
	}

	t, err2 := template.New("index").Parse(tpl)
	if err2 != nil {
		log.Println(1, "index html rendering error")
		log.Println(1, err2)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err2.Error()))
	}
	err = t.Execute(w, info)
	if err != nil {
		log.Println(1, "template error")
		log.Println(1, err)
	}
	return
}

func PodAddHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(1, "PodAddHandler")
	bodyBytes, err2 := ioutil.ReadAll(r.Body)
	if err2 != nil {
		fail("PodAddHandler: response read error", err2, w)
	}
	var p Pod
	err := json.Unmarshal(bodyBytes, &p)
	if err != nil {
		fail("PodAddHandler: json.Unmarshal error", err, w)
		return
	}
	dir, err := EtcdListDir("/rws/pods")
	if err != nil {
		fail("PodAddHandler: EtcdListDir error", err, w)
		return
	}
	found := false
	for _, k := range dir {
		keySplit := strings.Split(k.Key, "/")
		keyName := keySplit[len(keySplit)-1]
		if keyName == p.Name {
			found = true
			break
		}
	}
	if found == true {
		fail("PodAddHandler: pod already exists", errors.New("pod already exists"), w)
		return
	}
	hostsDir, err := EtcdListDir("/rws/hosts/")
	if err != nil {
		fail("PodAddHandler: EtcdListDir error", err, w)
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
			fail("PodAddHandler: http.get error", err, w)
			continue
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fail("PodAddHandler: ioutil.ReadAll error", err, w)
			continue
		}
		var ThatHost Host
		err3 := json.Unmarshal(body, &ThatHost)
		if err3 != nil {
			fail("PodAddHandler: json.Unmarshal error", err3, w)
			continue
		}
		if ThatHost.Disk >= p.Disk &&
			ThatHost.Cores >= p.Cores &&
			ThatHost.Memory >= p.Memory {
			keySplit := strings.Split(h.Key, "/")
			keyName := keySplit[len(keySplit)-1]
			url := "http://" + keyName + "/container_run"
			s := uniuri.New()
			pName := p.Name + "_" + s
			c := Container{p.Image, pName, p.Disk, p.Memory, p.Cores, h.Key, "", p.Cmd}
			b, err2 := json.Marshal(c)
			if err2 != nil {
				log.Println("PodAddHandler: json.Marshal error")
				log.Println(err2)
				continue
			}
			buf := bytes.NewBuffer(b)
			resp, err1 := http.Post(url, "application/json", buf)
			if err1 != nil {
				log.Println("PodAddHandler: http.Post error")
				log.Println(err1)
				continue
			}
			if resp.StatusCode != 200 {
				log.Println("PodAddHandler: request status code error")
				log.Println(resp.StatusCode)
				log.Println(resp)
				continue
			}
			body, err2 := ioutil.ReadAll(resp.Body)
			if err2 != nil {
				log.Println("PodAddHandler: response read error")
				log.Println(err2)
				continue
			}
			c.ID = string(body)
			p.Containers = append(p.Containers, c)
			i += 1
		}
	}
	s, err := json.Marshal(p)
	if err != nil {
		fail("PodAddHandler: json.Marshal error", err, w)
	}
	err7 := etcdCreateKey("/rws/pods/"+p.Name, string(s))
	if err7 != nil {
		fail("PodAddHandler: etcdSetKey error", err7, w)
	}
	log.Println("PodAddHandler: all pod containers running")
	w.Write([]byte("OK"))
	return
}

func PodStopHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("PodStopHandler")
	bodyBytes, err2 := ioutil.ReadAll(r.Body)
	if err2 != nil {
		fail("PodAdHandler: response read error", err2, w)
		return
	}
	var p Pod
	err := json.Unmarshal(bodyBytes, &p)
	if err != nil {
		fail("PodAddHandler: json.Unmarshal error", err, w)
		return
	}
	dir, err2 := EtcdListDir("/rws/hosts")
	if err2 != nil {
		fail("PodStopHandler: EtcdListDir error", err2, w)
		return
	}
	for _, c := range p.Containers {
		for _, host := range dir {
			url := "http://" + host.Key + "/container_list"
			body, err := http.Get(url)
			if err != nil {
				log.Println("get error")
				log.Println(body)
				continue
			}
			if body.StatusCode != 200 {
				log.Println("status code error")
				log.Println(body.StatusCode)
				continue
			}
			b, err2 := ioutil.ReadAll(body.Body)
			if err2 != nil {
				log.Println("response read error")
				log.Println(err2)
			}
			var RemoteContainers []Container
			err3 := json.Unmarshal(b, &RemoteContainers)
			if err != nil {
				log.Println("json unmarshal error")
				log.Println(err3)
				continue
			}
			for _, RemoteContainer := range RemoteContainers {
				if RemoteContainer.Name == c.ID {
					b := new(bytes.Buffer)
					json.NewEncoder(b).Encode(c)
					url := "http://" + host.Key + "/container_stop"
					resp, err1 := http.Post(url, "application/json", b)
					if err1 != nil {
						log.Println("request error")
						log.Println(err1)
						continue
					}
					if resp.StatusCode != 200 {
						log.Println("request status code error")
						log.Println(resp.StatusCode)
						log.Println(resp)
						continue
					}
					_, err2 := ioutil.ReadAll(resp.Body)
					if err2 != nil {
						log.Println("response read error")
						log.Println(err2)
						continue
					}
				}
			}
		}
	}
	return
}

func ListPods() (string, error) {
	log.Println("ListPods")
	pods, err := EtcdListDir("/rws/pods")
	if err != nil {
		log.Println("EtcdListDir error")
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
	log.Println("PodListHandler")
	s, err := ListPods()
	if err != nil {
		fail("PodsList error", err, w)
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(s))
	return
}

func PodRemoveHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("pod remove")
	var p Pod
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	err2 := etcdDeleteKey("/rws/pods/" + p.Name)
	if err2 != nil {
		fail("etcdDeleteKey error", err2, w)
	}
	return
}

func scheduler() {
	for {
		log.Println("scheduler: check pods and containers")
		dir, err := EtcdListDir("/rws/pods")
		if err != nil {
			log.Println("EtcdListDir error")
			log.Println(err)
			time.Sleep(60 * time.Second)
			continue
		}
		var pods []Pod
		for _, pod := range dir {
			var p Pod
			err2 := json.Unmarshal([]byte(pod.Value), &p)
			if err2 != nil {
				log.Println("scheduler: json.Unmarshal error")
				log.Println(err2)
				time.Sleep(60 * time.Second)
				continue
			}
			pods = append(pods, p)
		}
		dir2, err2 := EtcdListDir("/rws/containers")
		if err2 != nil {
			log.Println("EtcdListDir error")
			log.Println(err2)
			time.Sleep(60 * time.Second)
			continue
		}
		var containers []Container
		for _, cont := range dir2 {
			var c Container
			err3 := json.Unmarshal([]byte(cont.Value), &c)
			if err3 != nil {
				log.Println("scheduler: json.Unmarshal error")
				log.Println(err3)
				time.Sleep(60 * time.Second)
				continue
			}
			containers = append(containers, c)
		}
		if len(pods) == 0 {
			log.Println("scheduler: no pods found")
			time.Sleep(60 * time.Second)
			continue
		}
		var hosts []Host
		dir2, err5 := EtcdListDir("/rws/hosts")
		if err5 != nil {
			log.Println("scheduler: EtcdListDir error")
			log.Println(err5)
			time.Sleep(60 * time.Second)
			continue
		}
		for _, host := range dir2 {
			var h Host
			err2 := json.Unmarshal([]byte(host.Value), &h)
			if err2 != nil {
				log.Println("scheduler: json unmarshal error")
				log.Println(err2)
				time.Sleep(60 * time.Second)
				continue
			}
			hosts = append(hosts, h)
		}
		if len(hosts) == 0 {
			log.Println("scheduler: no hosts found")
			time.Sleep(60 * time.Second)
			continue
		}
		for _, p := range pods {
			log.Println("scheduler: Pod " + p.Name + " should have " + string(p.Count) + " containers")
			var foundContainers uint64
			for _, h := range hosts {
				hostRunningContainers, err4 := GetHostContainers(h.Name)
				if err4 != nil {
					log.Println("scheduler: getHostContainers error")
					log.Println(err4)
					time.Sleep(60 * time.Second)
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
						log.Println("scheduler: RunContainer error")
						log.Println(err)
						time.Sleep(60 * time.Second)
						continue
					}
					var c = Container{p.Image, p.Name, p.Disk, p.Memory, p.Cores, host.Name, id, p.Cmd}
					p.Containers = append(p.Containers, c)
				}
			}
			podMarshalled, err4 := json.Marshal(p)
			if err4 != nil {
				log.Println("scheduler: json.Marshal error")
				log.Println(err4)
				time.Sleep(60 * time.Second)
				continue
			}
			err5 := etcdSetKey("/rws/pods/"+p.Name, string(podMarshalled))
			if err5 != nil {
				log.Println("scheduler: etcdSetKey error")
				log.Println(err5)
				time.Sleep(60 * time.Second)
				continue
			}
		}
		time.Sleep(60 * time.Second)
	}
}
func main() {
	log.Println("starting server")
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile | log.Lmicroseconds)
	//hostNameBytes, err1 := ioutil.ReadFile("/etc/hostname")
	//if err1 != nil {
	//	log.Println(err1)
	//	panic("/etc/hostname reading error")
	//}
	//LocalHostNumber := string(hostNameBytes[len(hostNameBytes)-2])
	//LocalHostName = LocalIPPrefix + LocalHostNumber + ":" + LocalPort
	etcdCfg := etcdClient.Config{
		Endpoints: []string{EtcdHost},
		Transport: etcdClient.DefaultTransport,
	}
	var err error
	EtcdClient, err = etcdClient.New(etcdCfg)
	if err != nil {
		log.Println(err)
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
