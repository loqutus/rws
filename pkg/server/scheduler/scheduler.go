package scheduler

import (
	"encoding/json"
	"github.com/loqutus/rws/pkg/server/containers"
	"github.com/loqutus/rws/pkg/server/etcd"
	"github.com/loqutus/rws/pkg/server/hosts"
	"github.com/loqutus/rws/pkg/server/pods"
	"log"
	"time"
)

func Scheduler() {
	for {
		log.Println("scheduler: check pods and containers")
		dir, err := etcd.ListDir("/rws/pods")
		if err != nil {
			log.Println("EtcdListDir error")
			log.Println(err)
			time.Sleep(60 * time.Second)
			continue
		}
		var podsSlice []pods.Pod
		for _, pod := range dir {
			var p pods.Pod
			err2 := json.Unmarshal([]byte(pod.Value), &p)
			if err2 != nil {
				log.Println("scheduler: json.Unmarshal error")
				log.Println(err2)
				time.Sleep(60 * time.Second)
				continue
			}
			podsSlice = append(podsSlice, p)
		}
		dir2, err2 := etcd.ListDir("/rws/containers")
		if err2 != nil {
			log.Println("EtcdListDir error")
			log.Println(err2)
			time.Sleep(60 * time.Second)
			continue
		}
		var containersSlice []containers.Container
		for _, cont := range dir2 {
			var c containers.Container
			err3 := json.Unmarshal([]byte(cont.Value), &c)
			if err3 != nil {
				log.Println("scheduler: json.Unmarshal error")
				log.Println(err3)
				time.Sleep(60 * time.Second)
				continue
			}
			containersSlice = append(containersSlice, c)
		}
		if len(podsSlice) == 0 {
			log.Println("scheduler: no pods found")
			time.Sleep(60 * time.Second)
			continue
		}
		var hostsSlice []hosts.Host
		dir2, err5 := etcd.ListDir("/rws/hosts")
		if err5 != nil {
			log.Println("scheduler: EtcdListDir error")
			log.Println(err5)
			time.Sleep(60 * time.Second)
			continue
		}
		for _, host := range dir2 {
			var h hosts.Host
			err2 := json.Unmarshal([]byte(host.Value), &h)
			if err2 != nil {
				log.Println("scheduler: json unmarshal error")
				log.Println(err2)
				time.Sleep(60 * time.Second)
				continue
			}
			hostsSlice = append(hostsSlice, h)
		}
		if len(hostsSlice) == 0 {
			log.Println("scheduler: no hosts found")
			time.Sleep(60 * time.Second)
			continue
		}
		for _, p := range podsSlice {
			log.Println("scheduler: Pod " + p.Name + " should have " + string(p.Count) + " containers")
			var foundContainers uint64
			for _, h := range hostsSlice {
				hostRunningContainers, err4 := containers.GetHostContainers(h.Name)
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
				for _, host := range hostsSlice {
					id, err := containers.RunContainer(p.Image, p.Name, p.Cmd)
					if err != nil {
						log.Println("scheduler: RunContainer error")
						log.Println(err)
						time.Sleep(60 * time.Second)
						continue
					}
					var c = containers.Container{p.Image, p.Name, p.Disk, p.Memory, p.Cores, host.Name, id, p.Cmd}
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
			err5 := etcd.SetKey("/rws/pods/"+p.Name, string(podMarshalled))
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
