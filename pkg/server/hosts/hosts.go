package hosts

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/loqutus/rws/pkg/server/etcd"
	"github.com/loqutus/rws/pkg/server/utils"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Host struct {
	Name   string
	Port uint64
	Disk   uint64
	Memory uint64
	Cores  uint64
}

func AddHost(hostName string, hostPort uint64) error {
	log.Println(1, "Host add")
	dir, err := etcd.ListDir("/rws/hosts")
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
	HostInfo, err3 := GetHostInfo(hostName, hostPort)
	if err3 != nil {
		log.Println(1, "AddHost: host info get error")
		return err3
	}
	HostInfo.Port = hostPort
	b, err4 := json.Marshal(HostInfo)
	if err4 != nil {
		log.Println(1, "AddHost: host info json marshal error")
		return err4
	}
	HostInfoString := string(b)
	if found == false {
		err2 := etcd.CreateKey("/rws/hosts/"+hostName, HostInfoString)
		if err2 != nil {
			log.Println("AddHost: etcd.CreateKey error")
			return err2
		}
		log.Println(1, "AddHost: host "+hostName+" added")
	} else {
		log.Println(1, "AddHost: host already exists")
		return errors.New("host already exists")
	}
	return nil
}

func GetHostInfo(host string, hostPort uint64) (Host, error) {
	url := "http://" + host + ":" + strconv.FormatUint(hostPort, 10) + "/host_info"
	body, err := http.Get(url)
	if err != nil {
		log.Println(1, "GetHostInfo: get error")
		log.Println(1, "GetHostInfo: "+url)
		log.Println(1, body)
		return Host{}, err
	}
	var ThatHost Host
	err = json.NewDecoder(body.Body).Decode(&ThatHost)
	if err != nil{
		log.Println("GetHostInfo error: json decode error")
		log.Println(err)
		return Host{}, err
	}
	return ThatHost, nil
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
	err2 := AddHost(h.Name, h.Port)
	if err2 == nil {
		w.WriteHeader(http.StatusOK)
		_, err = fmt.Fprintf(w, "HostAddHandler: OK")
		if err != nil{
			log.Println("HostAddHandler: response write error")
			log.Println(err)
		}
	} else {
		utils.Fail("HostAddHandler: host create error", err2, w)
	}
}

func RemoveHost(hostName string) error {
	log.Println(1, "RemoveHost")
	dir, err := etcd.ListDir("/rws/hosts")
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
		err2 := etcd.DeleteKey("/rws/hosts/" + hostName)
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
		utils.Fail("HostRemoveHandler: response read error", err3, w)
	}
	var h Host
	err := json.Unmarshal(bodyBytes, &h)
	if err != nil {
		utils.Fail("HostRemoveHandler: json.Unmarshal error", err, w)
		return
	}
	err2 := RemoveHost(h.Name)
	if err2 == nil {
		w.WriteHeader(http.StatusOK)
		_, err = fmt.Fprintf(w, "OK")
		if err != nil{
			log.Println("HostRemoveHandler: response write error")
			log.Println(err)
		}
	} else {
		utils.Fail("HostRemoveHandler: RemoveHost error", err2, w)
		return
	}
}

func ListHosts() (string, error) {
	log.Println(1, "ListHosts")
	result := make(map[string]string)
	hosts, err := etcd.ListDir("/rws/hosts")
	if err != nil {
		log.Println(1, "etcd.ListDir error")
		log.Println(err)
		return "", err
	}
	if len(hosts) > 0 {
		for _, v := range hosts{
			result[v.Key] = v.Value
		}
		sm, err2 := json.Marshal(result)
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
		_, err = w.Write([]byte(s))
		if err != nil{
			log.Println("HostListHandler: response write error")
			log.Println(err)
		}
	} else {
		utils.Fail("ListHosts error", err, w)
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
	nameBytes, err := ioutil.ReadFile("/etc/hostname")
	if err != nil {
		return "", err
	}
	name := strings.TrimSpace(string(nameBytes))
	var c = Host{name, 0, di.Free, mi.Available, uint64(len(ci))}
	b, err := json.Marshal(c)
	return string(b), err
}

func HostInfoHandler(w http.ResponseWriter, _ *http.Request) {
	log.Println("HostInfoHandler")
	s, err := HostInfo()
	if err == nil {
		w.WriteHeader(http.StatusOK)
		_, err = fmt.Fprintf(w, s)
		if err != nil{
			log.Println("HostInfoHandler: response write error")
			log.Println(err)
		}
	} else {
		utils.Fail("HostInfo error", err, w)
	}
}