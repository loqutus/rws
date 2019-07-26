package web

import (
	"encoding/json"
	"fmt"
	"github.com/loqutus/rws/pkg/server/containers"
	"github.com/loqutus/rws/pkg/server/pods"
	"html/template"
	"log"
	"net/http"
	"strings"
)

type WebPod struct {
	Name   string
	Image  string
	Count uint64
	Disk   string
	Memory string
	Cores  uint64
	Containers []containers.Container
	Cmd    string
}

type WebPodsInfo struct {
	Pods []WebPod
}

func PodsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("web.PodsHandler")
	var pds []pods.Pod
	podsString, err := pods.ListPods()
	if err != nil {
		log.Println("PodsHandler: ListPods error")
		log.Println(err)
	} else {
		err2 := json.Unmarshal([]byte(podsString), &pds)
		if err2 != nil {
			log.Println("PodsHandler: json.Unmarshal podsString error")
			log.Println(err2)
		}
	}
	var WP WebPodsInfo
	for _, p := range pds {
		WP.Pods = append(WP.Pods, WebPod{Name: p.Name, Image: p.Image, Disk: ByteCountBinary(p.Disk), Memory: ByteCountBinary(p.Memory), Cores: p.Cores, Cmd: strings.Join(p.Cmd, " "), Containers:p.Containers})
	}
	fmt.Println(podsString)
	tmpl := template.New("pods")
	tmpl, err = tmpl.ParseFiles("/web/pods.html", "/web/inc/header.html", "/web/inc/navbar.html")
	if err != nil {
		log.Println("PodsHandler: template.ParseFiles error")
		log.Println(err)
	}
	err = tmpl.Execute(w, WP)
	if err != nil {
		log.Println("PodsHandler: tmpl.Execute error")
		log.Println(err)
	}
}
