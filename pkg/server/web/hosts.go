package web

import (
	"encoding/json"
	"fmt"
	"github.com/loqutus/rws/pkg/server/hosts"
	"html/template"
	"log"
	"net/http"
)

type HostWeb struct {
	Name   string
	Disk   string
	Memory string
	Cores  uint64
}

func ByteCountBinary(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

type HostsInfo struct {
	Hosts []hosts.Host
}

type HostsInfoWeb struct {
	Hosts []HostWeb
}

func HostsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("web.HostsHandler")
	var hsts map[string]string
	var HI HostsInfo
	var HW HostsInfoWeb
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
	for _, v := range hsts {
		var h hosts.Host
		err := json.Unmarshal([]byte(v), &h)
		if err != nil {
			log.Println("HostsHandler: json.Unmarshal h error")
			log.Println(err)
			continue
		}
		HI.Hosts = append(HI.Hosts, h)
	}
	for _, h := range HI.Hosts {
		HW.Hosts = append(HW.Hosts, HostWeb{Name: h.Name, Disk: ByteCountBinary(h.Disk), Memory: ByteCountBinary(h.Memory), Cores: h.Cores})
	}
	tmpl := template.New("hosts")
	tmpl, err = tmpl.ParseFiles("/web/hosts.html", "/web/inc/header.html", "/web/inc/navbar.html")
	if err != nil {
		log.Println("template.ParseFiles error")
		log.Println(err)
	}
	err = tmpl.Execute(w, HW)
	if err != nil {
		log.Println("HostsHandler: tmpl.Execute error")
		log.Println(err)
	}
}
