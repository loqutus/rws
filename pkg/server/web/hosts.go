package web

import (
	"encoding/json"
	"github.com/loqutus/rws/pkg/server/hosts"
	"html/template"
	"log"
	"net/http"
)


type HostsInfo struct {
	Hosts []hosts.Host
}

func HostsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("web.HostsHandler")
	var hsts map[string]string
	var HI HostsInfo
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
	for _, v := range hsts{
		var h hosts.Host
		err := json.Unmarshal([]byte(v), &h)
		if err != nil{
			log.Println("HostsHandler: json.Unmarshal h error")
			log.Println(err)
			continue
		}
		HI.Hosts = append(HI.Hosts, h)
	}
	tmpl := template.New("hosts")
	tmpl, err = tmpl.ParseFiles("/web/hosts.html", "/web/inc/header.html", "/web/inc/navbar.html")
	if err != nil {
		log.Println("template.ParseFiles error")
		log.Println(err)
	}
	err = tmpl.Execute(w, HI)
	if err != nil {
		log.Println("HostsHandler: tmpl.Execute error")
		log.Println(err)
	}
}
