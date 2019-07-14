package main

import (
	"github.com/loqutus/rws/pkg/server/conf"
	"github.com/loqutus/rws/pkg/server/containers"
	"github.com/loqutus/rws/pkg/server/hosts"
	"github.com/loqutus/rws/pkg/server/pods"
	"github.com/loqutus/rws/pkg/server/scheduler"
	"github.com/loqutus/rws/pkg/server/storage"
	"github.com/loqutus/rws/pkg/server/web"
	"log"
	"net/http"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile | log.Lmicroseconds)
	log.Println("starting server")
	go scheduler.Scheduler()
	http.HandleFunc("/storage_upload/", storage.UploadHandler)
	http.HandleFunc("/storage_download/", storage.DownloadHandler)
	http.HandleFunc("/storage_remove/", storage.RemoveHandler)
	http.HandleFunc("/storage_list", storage.ListHandler)
	http.HandleFunc("/storage_file_size/", storage.FileSizeHandler)
	http.HandleFunc("/container_run", containers.ContainerRunHandler)
	http.HandleFunc("/container_stop", containers.ContainerStopHandler)
	http.HandleFunc("/container_list", containers.ContainerListHandler)
	http.HandleFunc("/container_list_local", containers.ContainerListLocalHandler)
	http.HandleFunc("/container_remove", containers.ContainerRemoveHandler)
	http.HandleFunc("/pod_add", pods.PodAddHandler)
	http.HandleFunc("/pod_stop", pods.PodStopHandler)
	http.HandleFunc("/pod_list", pods.PodListHandler)
	http.HandleFunc("/pod_remove", pods.PodRemoveHandler)
	http.HandleFunc("/host_add", hosts.HostAddHandler)
	http.HandleFunc("/host_remove", hosts.HostRemoveHandler)
	http.HandleFunc("/host_list", hosts.HostListHandler)
	http.HandleFunc("/host_info", hosts.HostInfoHandler)
	http.HandleFunc("/web", web.IndexHandler)
	http.HandleFunc("/web/hosts", web.HostsHandler)
	if err := http.ListenAndServe(conf.Addr, nil)
	err != nil {
		panic(err)
	}
}