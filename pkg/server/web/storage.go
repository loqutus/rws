package web

import (
	"encoding/json"
	"github.com/loqutus/rws/pkg/server/storage"
	"html/template"
	"log"
	"net/http"
)

type FileWeb struct {
	Name     string
	Replicas uint64
	Size     string
	Host     string
}

type WebFilesInfo struct {
	Files []FileWeb
}

func StorageHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("web.FilesHandler")
	var fls []storage.File
	filesString, err := storage.StorageList()
	if err != nil {
		log.Println("web.StorageHandler: StorageList error")
		log.Println(err)
		return
	} else {
		err2 := json.Unmarshal([]byte(filesString), &fls)
		if err2 != nil {
			log.Println("web.StorageHandler: json.Unmarshal filesString error")
			log.Println(err2)
			return
		}
	}
	var WF WebFilesInfo
	for _, p := range fls {
		WF.Files = append(WF.Files, FileWeb{Name: p.Name, Host: p.Host, Size: ByteCountBinary(p.Size), Replicas: p.Replicas})
	}
	tmpl := template.New("storage")
	tmpl, err = tmpl.ParseFiles("/web/storage.html", "/web/inc/header.html", "/web/inc/navbar.html")
	if err != nil {
		log.Println("web.StorageHandler: template.ParseFiles error")
		log.Println(err)
	}
	err = tmpl.Execute(w, WF)
	if err != nil {
		log.Println("web.StorageHandler: tmpl.Execute error")
		log.Println(err)
	}
}
