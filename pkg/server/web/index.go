package web

import (
	"encoding/json"
	"github.com/loqutus/rws/pkg/server/containers"
	"github.com/loqutus/rws/pkg/server/hosts"
	"github.com/loqutus/rws/pkg/server/pods"
	"github.com/loqutus/rws/pkg/server/storage"
	"html/template"
	"log"
	"net/http"
)

type IndexInfo struct {
	hostsCount      int
	podscount       int
	containersCount int
	filesCount      int
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("web.IndexHandler")
	var hostsCount, podsCount, containersCount, filesCount int
	hsts := make(map[string]string)
	hostsString, err := hosts.ListHosts()
	if err != nil {
		log.Println("IndexHandler: HostsList error")
		log.Println(err)
	} else {
		err2 := json.Unmarshal([]byte(hostsString), &hsts)
		if err2 != nil {
			log.Println("IndexHandler: json.Unmarshal hostsString error")
			log.Println(err2)
		} else {
			hostsCount = len(hsts)
		}
	}
	pds := make(map[string]string)
	podsString, err := pods.ListPods()
	if err != nil {
		log.Println("IndexHandler: PodsList error")
		log.Println(err)
	} else {
		err2 := json.Unmarshal([]byte(podsString), &pds)
		if err2 != nil {
			log.Println("IndexHandler: json.Unmarshal PodsList error")
			log.Println(err2)
		} else {
			podsCount = len(pds)
		}
	}
	var cnts interface{}
	containersString, err := containers.ListAllContainers()
	if err != nil {
		log.Println("IndexHandler: ListAllContainers error")
		log.Println(err)
	} else {
		err2 := json.Unmarshal([]byte(containersString), &cnts)
		if err2 != nil {
			log.Println("IndexHandler: json.Unmarshal ContainersString error")
			log.Println(err2)
			log.Println(containersString)
		} else {
			containersCount = len(cnts)
		}
	}
	fls := make(map[string]string)
	filesString, err := storage.StorageList()
	if err != nil {
		log.Println("IndexHandler: StorageList error")
		log.Println(err)
	} else {
		err2 := json.Unmarshal([]byte(filesString), &fls)
		if err2 != nil {
			log.Println("IndexHandler: json.Unmarshal StorageList error")
			log.Println(err2)
		} else {
			filesCount = len(fls)
		}
	}
	II := IndexInfo{hostsCount, podsCount, containersCount, filesCount}
	tmpl, err := template.ParseFiles("web/index.html")
	if err != nil {
		log.Println("IndexHandler: template.ParseFiles error")
		log.Println(err)
	}
	err = tmpl.Execute(w, II)
	if err != nil {
		log.Println("IndexHandler: tmpl.Execute error")
		log.Println(err)
	}
}
