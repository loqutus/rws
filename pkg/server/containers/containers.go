package containers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/loqutus/rws/pkg/server/conf"
	"github.com/loqutus/rws/pkg/server/etcd"
	"github.com/loqutus/rws/pkg/server/hosts"
	"github.com/loqutus/rws/pkg/server/storage"
	"github.com/loqutus/rws/pkg/server/utils"
	"golang.org/x/net/context"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

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

func GetHostContainers(host string, port uint64) ([]Container, error) {
	url := "http://" + host + ":" + strconv.FormatUint(port, 10) + "/container_list_local"
	body, err := http.Get(url)
	if err != nil {
		log.Println(1, "GetHostContainers: http.get error")
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
	cont := Container{Name: containerName, Image: imageName, Host: conf.LocalHostName, ID: resp.ID}
	containerBytes, err5 := json.Marshal(cont)
	if err5 != nil {
		return "", err5
	}
	err6 := etcd.CreateKey("/rws/containers/"+containerName, string(containerBytes))
	if err6 != nil {
		log.Println(1, "RunContainer: etcd.CreateKey error")
		log.Println(1, err6)
		return "", err6
	}
	log.Println(1, "RunContainer: container "+resp.ID+" running")
	return resp.ID, nil
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
	containersNodes, err := etcd.ListDir("/rws/containers")
	if err != nil {
		log.Println(1, "ListAllContainers: etcd.ListDir error")
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
	if len(l) < 0{
		return "{}", nil
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
		utils.Fail(s, err, w)
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(s))
	if err != nil {
		log.Println("ContainerListHandler: response write error")
		log.Println(err)
	}
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
	dir, err := etcd.ListDir("/rws/containers/")
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
	containerString, err2 := etcd.GetKey("/rws/containers/" + containerName)
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
		utils.Fail("ContainerRunHandler: response read error", err2, w)
	}
	var c Container
	err := json.Unmarshal(bodyBytes, &c)
	if err != nil {
		utils.Fail("ContainerRunHandler: json.Unmarshal error", err, w)
		return
	}
	var ThatHost hosts.Host
	hostInfo, err := hosts.HostInfo()
	if err != nil {
		utils.Fail("ContainerRunHandler: HostInfo error", err, w)
		return
	}
	err3 := json.Unmarshal([]byte(hostInfo), &ThatHost)
	if err3 != nil {
		utils.Fail("ContainerRunHandler: json.Unmarshal error", err3, w)
	}
	if ThatHost.Disk >= c.Disk &&
		ThatHost.Cores >= c.Cores &&
		ThatHost.Memory >= c.Memory {
		id, err := RunContainer(c.Image, c.Name, c.Cmd)
		if err != nil {
			utils.Fail("ContainerRunHandler: RunContainer error", err, w)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(id))
	} else {
		utils.Fail("ContainerRunHandler: this host can't run this container", errors.New("can't run container on this host"), w)
		return
	}
}

func ContainerStopHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(1, "ContainerStopHandler")
	bodyBytes, err2 := ioutil.ReadAll(r.Body)
	if err2 != nil {
		utils.Fail("ContainerStopHandler: response read error", err2, w)
		return
	}
	var c Container
	err := json.Unmarshal(bodyBytes, &c)
	if err != nil {
		utils.Fail("ContainerStopHandler: json.Unmarshal error", err, w)
		return
	}
	dir, err4 := etcd.ListDir("/rws/containers")
	if err4 != nil {
		utils.Fail("ContainerStopHandler: etcd.ListDir error", err4, w)
		return
	}
	var cont Container
	found := false
	for _, k := range dir {
		keyName := storage.GetFileNameFromPath(k.Key)
		if keyName == c.Name {
			found = true
			contString, err5 := etcd.GetKey(k.Key)
			if err5 != nil {
				utils.Fail("ContainerStopHandler: etcd.GetKey error", err5, w)
				return
			}
			err6 := json.Unmarshal([]byte(contString), &cont)
			if err6 != nil {
				utils.Fail("ContainerStopHandler: json.Unmarshal error", err6, w)
				return
			}
			break
		}
	}
	if found == false {
		utils.Fail("ContainerStopHandler: container not found", errors.New(""), w)
		return
	}
	if cont.Host == conf.LocalHostName {
		err2 := StopContainer(cont.Name)
		if err2 != nil {
			utils.Fail("ContainerStopHandler: stopContainer utils.Failure", err2, w)
			return
		}
	} else {
		url := "http://" + cont.Host + "/container_stop/" + cont.Name
		b, err2 := json.Marshal(cont)
		if err2 != nil {
			utils.Fail("ContainerStopHandler: json Marshal error", err2, w)
			return
		}
		buf := bytes.NewBuffer(b)
		body, err3 := http.Post(url, "application/json", buf)
		if err3 == nil {
			if body.StatusCode != 200 {
				utils.Fail("ContainerStopHandler: http.Post status code error: "+string(body.StatusCode), err3, w)
				return
			}
		} else {
			utils.Fail("ContainerStopHandler: http.Post error", err3, w)
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
		utils.Fail("ContainerRemoveHandler: response read error", err2, w)
	}
	var c Container
	err := json.Unmarshal(bodyBytes, &c)
	if err != nil {
		utils.Fail("ContainerRemovehandler: json.Unmarshal error", err, w)
		return
	}
	dir, err4 := etcd.ListDir("/rws/containers")
	if err4 != nil {
		utils.Fail("ContainerRemoveHandler: etcd.ListDir error", err4, w)
		return
	}
	var cont Container
	found := false
	for _, k := range dir {
		keySplit := strings.Split(k.Key, "/")
		keyName := keySplit[len(keySplit)-1]
		if keyName == c.Name {
			found = true
			contString, err5 := etcd.GetKey(k.Key)
			if err5 != nil {
				utils.Fail("ContainerRemoveHandler: etcd.GetKey error", err5, w)
			}
			err6 := json.Unmarshal([]byte(contString), &cont)
			if err6 != nil {
				utils.Fail("ContainerRemove¡Handler: json.Unmarshal error", err6, w)
			}
		}
	}
	if found == false {
		utils.Fail("ContainerRemoveHandler: container not found", errors.New(""), w)
		return
	}
	if cont.Host == conf.LocalHostName {
		err2 := RemoveContainer(c.Name)
		if err2 == nil {
			fmt.Fprintf(w, "OK")
		} else {
			utils.Fail("ContainerStopHandler: stopContainer utils.Failure", err2, w)
			return
		}
	} else {
		url := "http://" + cont.Host + "/container_remove"
		b, err2 := json.Marshal(c)
		if err2 != nil {
			log.Println(1, err2)
			panic("json Marshal error")
		}
		buf := bytes.NewBuffer(b)
		body, err3 := http.Post(url, "application/json", buf)
		if err3 == nil {
			if body.StatusCode != 200 {
				utils.Fail("ContainerRemovepHandler: http.Post status code error: "+string(body.StatusCode), err3, w)
				return
			}
		} else {
			utils.Fail("ContainerStopHandler: http.Post error", err3, w)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	return
}
func ContainerListLocalHandler(w http.ResponseWriter, _ *http.Request) {
	s, err := ListLocalContainers()
	if err != nil {
		utils.Fail(s, err, w)
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(s))
}
