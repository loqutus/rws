package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"strconv"

	//"github.com/golang/lint/testdata"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"golang.org/x/net/context"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

const addr = "0.0.0.0:8888"
const DataDir = "data"

var hosts map[string]string

func StorageUploadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("storage upload")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("request reading error")
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	FileSize := len(body)
	PathSplit := strings.Split(r.URL.Path, "/")
	FileName := fmt.Sprintf("%s/%s", DataDir, PathSplit[len(PathSplit)-1])
	files, err3 := StorageList()
	if err3 != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("storage list error")
		return
	}
	FileSplit := strings.Split(files, "\n")
	for _, file := range FileSplit {
		if file == FileName {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println("file already exists")
			return
		}
	}
	di, err2 := disk.Usage("/")
	if err2 != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("Disk usage error")
		return
	}
	if di.Free > uint64(FileSize) {
		err2 := ioutil.WriteFile(FileName, []byte(body), 0644)
		if err2 != nil {
			fmt.Println("file write error")
			fmt.Println(err2)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Println("file " + FileName + " uploaded")
		w.WriteHeader(http.StatusOK)
		return
	} else {
		for host, port := range hosts {
			url := fmt.Sprintf("http://%s:%s/host_info", host, port)
			body, err := http.Get(url)
			if err != nil {
				fmt.Println("get error")
				fmt.Println(body)
				continue
			}
			var ThatHost HostConfig
			json.NewDecoder(body.Body).Decode(&ThatHost)
			if uint64(FileSize) < ThatHost.DISK {
				fmt.Println("uploading to " + host + ":" + port)
				url := fmt.Sprintf("%s:%s/storage_upload/%s", host, port, FileName)
				dat, err := http.Post(url, "application/octet-stream", r.Body)
				if err != nil {
					fmt.Println("post error")
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
	FileName := PathSplit[len(PathSplit)-1]
	fmt.Println("filename: " + FileName)
	files, err3 := StorageList()
	if err3 != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("storage list error")
		return
	}
	FileSplit := strings.Split(files, "\n")
	for _, file := range FileSplit {
		if file == FileName {
			dat, err1 := ioutil.ReadFile(fmt.Sprintf("data/%s", FileName))
			if err1 != nil {
				fmt.Println("file read error: " + FileName)
				fmt.Println(err1)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("file read error"))
				return
			}
			_, err2 := w.Write(dat)
			if err2 != nil {
				fmt.Println("request write error")
				fmt.Println(err2)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("request write error"))
				return
			}
			fmt.Println("file " + FileName + " downloaded")
			w.WriteHeader(http.StatusOK)
			return
		}
	}
	for host, port := range hosts {
		url := fmt.Sprintf("http://%s:%s/list_files", host, port)
		body, err := http.Get(url)
		if err != nil {
			fmt.Println("get error")
			fmt.Println(body)
			continue
		}
		FileSplit := strings.Split(files, "\n")
		for _, file := range FileSplit {
			if file == FileName {
				url := fmt.Sprintf("%s:%s/storage_download/%s", host, port, FileName)
				dat, err1 := http.Get(url)
				if err1 != nil {
					fmt.Println(err1)
					fmt.Println("get error")
				}
				if dat.StatusCode != 200 {
					fmt.Println(dat.StatusCode)
					fmt.Println("status code error")
				}
				BodyBytes, err2 := ioutil.ReadAll(dat.Body)
				if err2 != nil {
					fmt.Println(err2)
					fmt.Println("body read error")
				}
				w.WriteHeader(http.StatusOK)
				w.Write(BodyBytes)
				return
			}
		}
	}
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("file not found"))
	return
}

func StorageRemoveHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("storage remove")
	PathSplit := strings.Split(r.URL.Path, "/")
	FileName := PathSplit[len(PathSplit)-1]
	files, err := StorageList()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("storage list error")
		return
	}
	FileSplit := strings.Split(files, "\n")
	for _, file := range FileSplit {
		if file == FileName {
			err := os.Remove(fmt.Sprintf("data/%s", FileName))
			if err != nil {
				fmt.Println("file remove error: " + FileName)
				fmt.Println(err)
				return
			}
			fmt.Println("file " + FileName + " removed locally")

		}
	}
	for host, port := range hosts {
		url := fmt.Sprintf("http://%s:%s/storage_remove/%s", host, port, FileName)
		body, err := http.Get(url)
		if err != nil {
			fmt.Println("get error")
			fmt.Println(body)
			continue
		}
		if body.StatusCode != 200 {
			fmt.Println(body.StatusCode)
			fmt.Println("status code error")
			continue
		}
		fmt.Println("file " + FileName + " removed from Host " + host)
	}
	w.WriteHeader(http.StatusOK)
}

func StorageListAll() (string, error) {
	var l []string
	for host, port := range hosts {
		url := fmt.Sprintf("http://%s:%s/storage_list", host, port)
		body, err := http.Get(url)
		if err != nil {
			fmt.Println("get error")
			fmt.Println(err)
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
		FileSplit := strings.Split(string(b), "\n")
		for _, FileRemote := range FileSplit {
			found := false
			for _, FileLocal := range l {
				if FileLocal == FileRemote {
					found = true
					break
				}
			}
			if found == false {
				l = append(l, FileRemote+" "+host+" "+port)
			}
		}
	}
	s := strings.Join(l, "\n")
	return s, nil
}

func StorageListAllHandler(w http.ResponseWriter, _ *http.Request) {
	fmt.Println("storage all list")
	s, err := StorageListAll()
	if err != nil {
		fmt.Println("storage all list error")
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
		return
	}
	_, err2 := w.Write([]byte(s))
	if err2 != nil {
		fmt.Println("request write error")
		fmt.Println(err2)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func StorageList() (string, error) {
	files, err := ioutil.ReadDir(DataDir)
	if err != nil {
		return "", err
	}
	var l []string
	for _, f := range files {
		l = append(l, f.Name())
	}
	s := strings.Join(l, "\n")
	return s, nil
}

func StorageListHandler(w http.ResponseWriter, _ *http.Request) {
	fmt.Println("storage list")
	s, err := StorageList()
	if err != nil {
		fmt.Println("storage list error")
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	_, err2 := w.Write([]byte(s))
	if err2 != nil {
		fmt.Println("request write error")
		fmt.Println(err2)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err2.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func StorageFileSizeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("storage Size")
	fmt.Println(r.URL.Path)
	PathSplit := strings.Split(r.URL.Path, "/")
	FileName := PathSplit[len(PathSplit)-1]
	fmt.Println("filename: " + FileName)
	fi, e := os.Stat(DataDir + "/" + FileName)
	if e != nil {
		fmt.Println("file stat error ")
	}
	size := fi.Size()
	w.Write([]byte(string(size)))
	w.WriteHeader(http.StatusOK)
	return
}

func ListContainers() string {
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
	if len(containers) == 0 {
		fmt.Println("there is no running containers on this Host")
		return ""
	}
	for _, c := range containers {
		n := fmt.Sprintf("%s %s", c.ID, c.Image)
		l = append(l, n)
	}
	s := strings.Join(l, "\n")
	return s
}

func ListAllContainers() string {
	LocalContainers := ListContainers()
	LocalContainersSplit := strings.Split(LocalContainers, "\n")
	for host, port := range hosts {
		url := fmt.Sprintf("http://%s:%s/container_list", host, port)
		body, err := http.Get(url)
		if err != nil {
			fmt.Println("get error")
			fmt.Println(body.Body)
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
			continue
		}
		RemoteContainersSplit := strings.Split(string(b), "\n")
		for _, ContainerRemote := range RemoteContainersSplit {
			found := false
			for _, ContainerLocal := range LocalContainersSplit {
				if ContainerLocal == ContainerRemote {
					found = true
					break
				}
			}
			if found == false {
				LocalContainersSplit = append(LocalContainersSplit, ContainerRemote)
			}
		}
	}
	s := strings.Join(LocalContainersSplit, "\n")
	return s
}

func ContainerListHandler(w http.ResponseWriter, _ *http.Request) {
	s := ListContainers()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(s))
}

func ContainerListAllHandler(w http.ResponseWriter, _ *http.Request) {
	s := ListAllContainers()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(s))
}

func RunContainer(imageName, containerName string) (string, error) {
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
		fmt.Println("Image pull error")
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

func StopContainer(containerName string) error {
	fmt.Println("stop container")
	ctx := context.Background()
	c := client.WithVersion("1.38")
	cli, err1 := client.NewClientWithOpts(c)
	if err1 != nil {
		fmt.Println("client create error")
		fmt.Println(err1)
		return err1
	}
	ContainerId, _ := GetContainerId(containerName)
	fmt.Println(ContainerId)
	err2 := cli.ContainerStop(ctx, ContainerId, nil)
	if err2 != nil {
		fmt.Println("container stop error")
		fmt.Println(err2)
		return err2
	}
	return nil
}

func GetContainerId(containerName string) (string, error) {
	fmt.Println("get containerId")
	c := client.WithVersion("1.38")
	cli, err := client.NewClientWithOpts(c)
	if err != nil {
		fmt.Println("client create error")
		fmt.Println(err)
		return "", err
	}
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		fmt.Println("containerList error")
		fmt.Println(err)
		return "", err
	}
	for _, c := range containers {
		for _, name := range c.Names {
			if name[1:] == containerName {
				return c.ID, nil
			}
		}
	}
	return "", errors.New("container not found")
}

func RemoveContainer(containerName string) error {
	fmt.Println("remove container")
	ContainerID, err := GetContainerId(containerName)
	if err != nil {
		fmt.Println("get container id error")
		fmt.Println(err)
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

type Container struct {
	Image string
	Name  string
}

func ContainerRunHandler(w http.ResponseWriter, r *http.Request) {
	var c Container
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	id, err := RunContainer(c.Image, c.Name)
	if err == nil {
		fmt.Fprintf(w, id)
	} else {
		http.Error(w, err.Error(), 500)
	}
}

func ContainerStopHandler(w http.ResponseWriter, r *http.Request) {
	var c Container
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	err2 := StopContainer(c.Name)
	if err2 == nil {
		fmt.Fprintf(w, "OK")
	} else {
		http.Error(w, err.Error(), 500)
	}
}

func ContainerRemoveHandler(w http.ResponseWriter, r *http.Request) {
	var c Container
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	err2 := RemoveContainer(c.Name)
	if err2 == nil {
		fmt.Fprintf(w, "OK")
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
	}
}

func AddHost(hostName, port string) error {
	fmt.Println("Host add")
	if _, ok := hosts[hostName]; ok {
		return errors.New("Host already exists")
	} else {
		hosts[hostName] = port
	}
	return nil
}

type Host struct {
	Name string
	Port string
}

func HostAddHandler(w http.ResponseWriter, r *http.Request) {
	var h Host
	err := json.NewDecoder(r.Body).Decode(&h)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	err2 := AddHost(h.Name, h.Port)
	if err2 == nil {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err2.Error())
	}
}

func RemoveHost(hostName string) error {
	fmt.Println("remove Host")
	if _, ok := hosts[hostName]; ok {
		delete(hosts, hostName)
	} else {
		return errors.New("Host not found")
	}
	return nil
}

func HostRemoveHandler(w http.ResponseWriter, r *http.Request) {
	var h Host
	err := json.NewDecoder(r.Body).Decode(&h)
	if err != nil {
		http.Error(w, err.Error(), 500)
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
	fmt.Println("list hosts")
	var l []string
	for k, _ := range hosts {
		l = append(l, k)
	}
	if len(l) > 0 {
		s := strings.Join(l, "\n")
		return s, nil
	} else {
		return "", errors.New("hosts list is empty")
	}
}

func HostListHandler(w http.ResponseWriter, _ *http.Request) {
	s, err := ListHosts()
	if err == nil {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, s)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error")
	}
}

type HostConfig struct {
	CPUS   uint64
	MEMORY uint64
	DISK   uint64
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
	c := HostConfig{uint64(len(ci)), mi.Available, di.Free}
	b, err := json.Marshal(c)
	return string(b), err
}

func HostInfoHandler(w http.ResponseWriter, _ *http.Request) {
	s, err := HostInfo()
	if err == nil {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, s)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error")
	}
}

type InfoStorage struct {
	Name string
	Size uint64
	Host string
}

type InfoHosts struct {
	Name   string
	CPUS   uint64
	MEMORY uint64
	DISK   uint64
}

type InfoPods struct {
	Name   string
	Image  string
	Count  uint64
	Cpus   uint64
	Memory uint64
	Disk   uint64
}

type InfoContainers struct {
	Name  string
	Image string
	Host  string
}

type Info struct {
	Storage    []InfoStorage
	Hosts      []InfoHosts
	Pods       []InfoPods
	Containers []InfoContainers
}

func GetFileSize(filename, host, port string) (uint64, error) {
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

func GetHostInfo(host, port string) (InfoHosts, error) {
	url := fmt.Sprintf("http://%s:%s/host_info", host, port)
	body, err := http.Get(url)
	if err != nil {
		fmt.Println("get error")
		fmt.Println(body)
		return InfoHosts{}, err
	}
	var ThatHost HostConfig
	json.NewDecoder(body.Body).Decode(&ThatHost)
	h := InfoHosts{host, ThatHost.CPUS, ThatHost.MEMORY, ThatHost.DISK}
	return h, nil
}

func GetHostPods(host, port string) ([]InfoPods, error) {
	url := fmt.Sprintf("http://%s:%s/pod_list", host, port)
	body, err := http.Get(url)
	if err != nil {
		fmt.Println("get error")
		fmt.Println(body)
		return []InfoPods{}, err
	}
	var ThatPods []pod
	json.NewDecoder(body.Body).Decode(&ThatPods)
	var ThatHostPods []InfoPods
	for _, pod := range ThatPods {
		TempPod := InfoPods{pod.name, pod.image, pod.count, pod.cpus, uint64(pod.memory), uint64(pod.disk)}
		ThatHostPods = append(ThatHostPods, TempPod)
	}
	return ThatHostPods, nil
}

func GetHostContainers(host, port string) ([]InfoContainers, error) {
	url := fmt.Sprintf("http://%s:%s/container_list", host, port)
	body, err := http.Get(url)
	if err != nil {
		fmt.Println("get error")
		fmt.Println(body)
		return []InfoContainers{}, err
	}
	BodyBytes, err2 := ioutil.ReadAll(body.Body)
	if err2 != nil {
		fmt.Println("GetHostContainers error")
		fmt.Println(err2)
		return []InfoContainers{}, err2
	}
	if len(BodyBytes) == 0 {
		fmt.Println("no containers running on Host")
		return []InfoContainers{}, errors.New("no containers running on Host")
	}
	ThatHostContainersSplit := strings.Split(string(BodyBytes), "\n")
	var HostContainers []InfoContainers
	for _, ContainerString := range ThatHostContainersSplit {
		ContainerStringSplit := strings.Split(ContainerString, " ")
		ID := ContainerStringSplit[0]
		IMAGE := ContainerStringSplit[1]
		x := InfoContainers{ID, IMAGE, host}
		HostContainers = append(HostContainers, x)
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
				<th>CPUS</th>
				<th>MEM</th>
				<th>DISK</th>
			</tr>
			{{range .Hosts}}
			<tr>
				<td>{{.Name}}</td>
				<td>{{.CPUS}}</td>
				<td>{{.MEMORY}}</td>
				<td>{{.DISK}}</td>
			</tr>
			{{end}}
		</table>
		<h2>Pods</h2>
		<table>
			<tr>
				<th>Name</th>
				<th>Image</th>
				<th>Count</th>
				<th>CPUS</th>
				<th>Memory</th>
				<th>DISK</th>
			</tr>
			{{range .Pods}}
			<tr>
				<td>{{.Name}}</td>
				<td>{{.Image}}</td>
				<td>{{.Count}}</td>
				<td>{{.Cpus}}</td>
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
	//var IsStorageListEmpty, IsPodsListEmpty, IsContainersListEmpty bool
	if len(hosts) == 0 {
		fmt.Println("Host list is empty")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Host list is empty"))
		return
	}
	FilesSplit, err := StorageListAll()
	if err != nil {
		fmt.Println("StorageListAll error")
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	FilesSplitSplit := strings.Split(FilesSplit, "\n")
	for _, FileAndHost := range FilesSplitSplit {
		PathSplit := strings.Split(FileAndHost, " ")
		FileName := PathSplit[0]
		HostName := PathSplit[1]
		Port := PathSplit[2]
		FileSize, err := GetFileSize(FileName, HostName, Port)
		if err != nil {
			fmt.Println("GetFileSize error ")
			fmt.Println("file: " + FileName)
			fmt.Println(err)
			continue
		}
		FileStorage := InfoStorage{FileName, FileSize, HostName}
		info.Storage = append(info.Storage, FileStorage)
	}
	for host, port := range hosts {
		HostInfo, err := GetHostInfo(host, port)
		if err != nil {
			fmt.Println("GetHostInfo error")
			fmt.Println("Host: " + host)
			fmt.Println(err)
			continue
		}
		InfoHost := InfoHosts{host, HostInfo.CPUS, HostInfo.MEMORY, HostInfo.DISK}
		info.Hosts = append(info.Hosts, InfoHost)
	}
	for host, port := range hosts {
		HostPods, err := GetHostPods(host, port)
		if err != nil {
			fmt.Println("GetHostPods error")
			fmt.Println("Host: " + host)
			fmt.Println(err)
			continue
		}
		for _, Pod := range HostPods {
			info.Pods = append(info.Pods, Pod)
		}
	}
	for host, port := range hosts {
		HostContainers, err := GetHostContainers(host, port)
		if err != nil {
			fmt.Println("GetHostContainers error")
			fmt.Println("Host: " + host)
			fmt.Println(err)
			continue
		}
		for _, Container := range HostContainers {
			info.Containers = append(info.Containers, Container)
		}
	}
	t, err2 := template.New("index").Parse(tpl)
	if err2 != nil {
		fmt.Println("index html rendering error")
		fmt.Println(err2)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err2.Error()))
	}
	err3 := t.Execute(w, info)
	if err3 != nil {
		fmt.Println("template error")
		fmt.Println(err3)
	}
	return
}

type pod struct {
	name    string
	image   string
	cpus    uint64
	disk    uint64
	memory  uint64
	count   uint64
	enabled bool
	ids     []string
}

var pods []pod

func PodRunHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("pod run")
	var p pod
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	pods = append(pods, p)
	var i uint64
	for host, port := range hosts {
		if i < p.count {
			url := fmt.Sprintf("http://%s:%s/host_info", host, port)
			body, err := http.Get(url)
			if err != nil {
				fmt.Println("get error")
				fmt.Println(body)
				continue
			}
			var ThatHost HostConfig
			json.NewDecoder(body.Body).Decode(&ThatHost)
			if ThatHost.DISK >= p.disk && ThatHost.CPUS >= p.cpus && ThatHost.MEMORY >= p.memory {
				url := fmt.Sprintf("http://%s:%s/container_run", host, port)
				c := Container{p.image, p.name + "-" + string(i)}
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
				p.ids = append(p.ids, string(body))
				fmt.Println(body)
				i += 1
			}
		} else {
			break
		}
	}
	fmt.Println("all pod containers running")
	return
}

func PodStopHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("pod stop")
	var p pod
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	for _, id := range p.ids {
		for host, port := range hosts {
			url := fmt.Sprintf("http://%s/%s/container_list", host, port)
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
			RemoteContainersSplit := strings.Split(string(b), "\n")
			for _, RemoteContainer := range RemoteContainersSplit {
				if RemoteContainer == id {
					c := Container{"", id}
					b := new(bytes.Buffer)
					json.NewEncoder(b).Encode(c)
					url := fmt.Sprintf("%s:%s/container_stop", host, port)
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

func PodListHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("pod list")
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(pods)
	w.WriteHeader(http.StatusOK)
	w.Write(b.Bytes())
	return
}

func PodRemoveHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("pod remove")
	var p pod
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	removed := false
	for i, other := range pods {
		if other.name == p.name {
			pods = append(pods[:i], pods[i+1:]...)
			removed = true
		}
	}
	if removed {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(p.name))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(p.name))
	}
	return
}

func scheduler() {
	for {
		fmt.Println("run scheduler")
		if len(pods) == 0 {
			fmt.Println("no pods defined")
			time.Sleep(60 * time.Second)
			continue
		}
		for index, p := range pods {
			fmt.Println("Pod %s should have %s containers", p.name, len(p.ids))
			var FoundIDs []string
			for host, port := range hosts {
				url := fmt.Sprintf("http://%s:%s/containers_list", host, port)
				body, err := http.Get(url)
				if err != nil {
					fmt.Println("get error")
					fmt.Println(body.Body)
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
					continue
				}
				RemoteContainersSplit := strings.Split(string(b), "\n")
				for _, remoteId := range RemoteContainersSplit {
					for _, id := range p.ids {
						if id == remoteId {
							FoundIDs = append(FoundIDs, id)
							break
						}
					}
				}
			}
			fmt.Println("Pod %s have %s running containers", p.name, len(FoundIDs))
			if len(FoundIDs) < len(p.ids) {
				for IDNum, ID := range p.ids {
					found := false
					for _, FoundID := range FoundIDs {
						if ID == FoundID {
							found = true
							break
						}
					}
					if found == false {
						p.ids = append(p.ids[:IDNum], p.ids[IDNum+1:]...)
					}
				}
				RunNum := len(p.ids) - len(FoundIDs)
				i := 0
				for {
					for host, port := range hosts {
						if i >= RunNum {
							break
						}
						url := fmt.Sprintf("%s:%s/container_run", host, port)
						ContainerNameId := len(p.ids)
						name := p.name + "-" + string(ContainerNameId)
						c := Container{p.image, name}
						b := new(bytes.Buffer)
						json.NewEncoder(b).Encode(c)
						resp, err1 := http.Post(url, "application/json", b)
						if err1 != nil {
							fmt.Println(err1)
							panic("request error")
						}
						if resp.StatusCode != 200 {
							fmt.Println(resp.StatusCode)
							fmt.Println(resp)
							panic("request status code error")
						}
						b2, err2 := ioutil.ReadAll(resp.Body)
						if err2 != nil {
							fmt.Println(err2)
							panic("response read error")
						}
						fmt.Println("run new container for pod %s", p.name)
						pods[index].ids = append(pods[index].ids, string(b2))
						i += 1
					}
				}
				if i >= RunNum {
					break
				}
			}
			time.Sleep(60 * time.Second)
		}
	}
}

func main() {
	fmt.Println("starting server")
	hosts = make(map[string]string)
	go scheduler()
	http.HandleFunc("/storage_upload/", StorageUploadHandler)
	http.HandleFunc("/storage_download/", StorageDownloadHandler)
	http.HandleFunc("/storage_remove/", StorageRemoveHandler)
	http.HandleFunc("/storage_list", StorageListHandler)
	http.HandleFunc("/storage_list_all", StorageListAllHandler)
	http.HandleFunc("/storage_file_size/", StorageFileSizeHandler)
	http.HandleFunc("/container_run", ContainerRunHandler)
	http.HandleFunc("/container_stop", ContainerStopHandler)
	http.HandleFunc("/container_list", ContainerListHandler)
	http.HandleFunc("/container_list_all", ContainerListAllHandler)
	http.HandleFunc("/container_remove", ContainerRemoveHandler)
	http.HandleFunc("/pod_run", PodRunHandler)
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
