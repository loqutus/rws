package web

import (
	"encoding/json"
	"github.com/loqutus/rws/pkg/server/containers"
	"html/template"
	"log"
	"net/http"
	"strings"
)

type WebContainer struct {
	Image  string
	Name   string
	Disk   string
	Memory string
	Cores  uint64
	Host   string
	ID     string
	Cmd    string
}

type WebContainersInfo struct {
	Containers []WebContainer
}

func ContainersHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("web.ContainersHandler")
	var cnts []containers.Container
	containersString, err := containers.ListAllContainers()
	if err != nil {
		log.Println("ContainersHandler: ListAllContainers error")
		log.Println(err)
	} else {
		err2 := json.Unmarshal([]byte(containersString), &cnts)
		if err2 != nil {
			log.Println("ContainersHandler: json.Unmarshal containersString error")
			log.Println(err2)
		}
	}
	var WC WebContainersInfo
	for _, c := range cnts {
		WC.Containers = append(WC.Containers, WebContainer{Name: c.Name, Image: c.Image, Disk: ByteCountBinary(c.Disk), Memory: ByteCountBinary(c.Memory), Cores: c.Cores, Host: c.Host, ID: c.ID[0:5], Cmd: strings.Join(c.Cmd, " ")})
	}
	tmpl := template.New("containers")
	tmpl, err = tmpl.ParseFiles("/web/containers.html", "/web/inc/header.html", "/web/inc/navbar.html")
	if err != nil {
		log.Println("ContainerHandler: template.ParseFiles error")
		log.Println(err)
	}
	err = tmpl.Execute(w, WC)
	if err != nil {
		log.Println("ContainersHandler: tmpl.Execute error")
		log.Println(err)
	}
}
