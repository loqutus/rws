package web

import (
	"encoding/json"
	"github.com/loqutus/rws/pkg/server/hosts"
	"html/template"
	"log"
	"net/http"
)

type HostsInfo struct {
	Hosts []map[string]string
}

func HostsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("web.HostsHandler")
	var hsts []map[string]string
	hostsString, err := hosts.ListHosts()
	if err != nil {
		log.Println("HostsHandler: HostsList error")
		log.Println(err)
	} else {
		err2 := json.Unmarshal([]byte(hostsString), &hsts)
		if err2 != nil {
			log.Println("HostsHandler: json.Unmarshal hostsString error")
			log.Println(err2)
		}
	}
	tmpl := template.New("hosts")
	tmpl, err = tmpl.ParseFiles("/web/hosts.html", "/web/inc/header.html", "/web/inc/navbar.html")
	if err != nil {
		log.Println("template.ParseFiles error")
		log.Println(err)
	}
	HI := HostsInfo{hsts}
	err = tmpl.Execute(w, HI)
	if err != nil {
		log.Println("HostsHandler: tmpl.Execute error")
		log.Println(err)
	}
}
