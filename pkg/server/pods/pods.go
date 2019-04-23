package pods

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dchest/uniuri"
	"github.com/loqutus/rws/pkg/server/containers"
	"github.com/loqutus/rws/pkg/server/etcd"
	"github.com/loqutus/rws/pkg/server/hosts"
	"github.com/loqutus/rws/pkg/server/utils"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type Pod struct {
	Name       string
	Image      string
	Count      uint64
	Cores      uint64
	Memory     uint64
	Disk       uint64
	Cmd        []string
	Containers []containers.Container
}

func GetHostPods(host string) ([]Pod, error) {
	url := fmt.Sprintf("http://" + host + "/pod_list")
	body, err := http.Get(url)
	if err != nil {
		log.Println(1, "GetHostPods: get error")
		log.Println(1, body)
		return []Pod{}, err
	}
	var ThatHostPods []Pod
	err = json.NewDecoder(body.Body).Decode(&ThatHostPods)
	if err != nil {
		log.Println("GetHostPods: json decode error")
		log.Println(err)
		return []Pod{}, err
	}
	for _, pod := range ThatHostPods {
		ThatHostPods = append(ThatHostPods, pod)
	}
	return ThatHostPods, nil
}

func PodAddHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(1, "PodAddHandler")
	bodyBytes, err2 := ioutil.ReadAll(r.Body)
	if err2 != nil {
		utils.Fail("PodAddHandler: response read error", err2, w)
		return
	}
	var p Pod
	err := json.Unmarshal(bodyBytes, &p)
	if err != nil {
		utils.Fail("PodAddHandler: json.Unmarshal error", err, w)
		return
	}
	dir, err := etcd.ListDir("/rws/pods")
	if err != nil {
		utils.Fail("PodAddHandler: Etcd.ListDir error", err, w)
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
		utils.Fail("PodAddHandler: pod already exists", errors.New("pod already exists"), w)
 		return
	}
	hostsDir, err := etcd.ListDir("/rws/hosts/")
	if err != nil {
		utils.Fail("PodAddHandler: Etcd.ListDir error", err, w)
		return
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
			utils.Fail("PodAddHandler: http.get error", err, w)
			continue
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			utils.Fail("PodAddHandler: ioutil.ReadAll error", err, w)
			continue
		}
		var ThatHost hosts.Host
		err3 := json.Unmarshal(body, &ThatHost)
		if err3 != nil {
			utils.Fail("PodAddHandler: json.Unmarshal error", err3, w)
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
			c := containers.Container{p.Image, pName, p.Disk, p.Memory, p.Cores, keyName, "", p.Cmd}
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
		utils.Fail("PodAddHandler: json.Marshal error", err, w)
	}
	err7 := etcd.CreateKey("/rws/pods/"+p.Name, string(s))
	if err7 != nil {
		utils.Fail("PodAddHandler: etcd.SetKey error", err7, w)
	}
	log.Println("PodAddHandler: all pod containers running")
	w.Write([]byte("OK"))
	return
}

func PodStopHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("PodStopHandler")
	bodyBytes, err2 := ioutil.ReadAll(r.Body)
	if err2 != nil {
		utils.Fail("PodAdHandler: response read error", err2, w)
		return
	}
	var p Pod
	err := json.Unmarshal(bodyBytes, &p)
	if err != nil {
		utils.Fail("PodAddHandler: json.Unmarshal error", err, w)
		return
	}
	dir, err2 := etcd.ListDir("/rws/hosts")
	if err2 != nil {
		utils.Fail("PodStopHandler: Etcd.ListDir error", err2, w)
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
			var RemoteContainers []containers.Container
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
	pods, err := etcd.ListDir("/rws/pods")
	if err != nil {
		log.Println("Etcd.ListDir error")
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
		utils.Fail("PodsList error", err, w)
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
	err2 := etcd.DeleteKey("/rws/pods/" + p.Name)
	if err2 != nil {
		utils.Fail("etcd.DeleteKey error", err2, w)
	}
	return
}
