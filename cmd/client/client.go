package main

import (
	"flag"
	"fmt"
	"github.com/loqutus/rws/pkg/client/conf"
	"github.com/loqutus/rws/pkg/client/containers"
	"github.com/loqutus/rws/pkg/client/hosts"
	"github.com/loqutus/rws/pkg/client/pods"
	"github.com/loqutus/rws/pkg/client/storage"
	"strings"
)

var HostName = conf.HostName

func main() {
	// client --type storage --action upload --name file
	// client --type storage --action list
	var action, name, image, port, cmd string
	var cores, disk, memory, count uint64
	flag.StringVar(&action, "action", "", conf.Actions)
	flag.StringVar(&image, "image", "", "redis or mysql")
	flag.StringVar(&name, "name", "", "container/file/host name")
	flag.StringVar(&port, "port", "", "host port")
	flag.Uint64Var(&cores, "cores", 1, "cores for each container in Pod")
	flag.Uint64Var(&disk, "disk", 1, "disk for each container in Pod")
	flag.Uint64Var(&memory, "memory", 1, "memory for each container in Pod")
	flag.Uint64Var(&count, "count", 1, "containers cound in Pod")
	flag.StringVar(&cmd, "cmd", "", "command to run in container")
	flag.StringVar(&HostName, "hostname", "http://localhost:8888", "hostname to connect to")
	flag.Parse()
	switch action {
	case "storage_upload", "storage_download", "storage_remove", "storage_list", "storage_list_all":
		if name != "" && action != "storage_list" && action != "storage_list_all" {
			storage.Storage(action, name)
		} else if name == "" && action == "storage_list" {
			storage.Storage(action, "")
		} else {
			panic("file name required")
		}
	case "container_run", "container_stop", "container_list", "container_list_local", "container_remove":
		cmds := strings.Split(cmd, " ")
		r := containers.ContainerAction(action, image, name, cmds)
		fmt.Println(r)
	case "host_add", "host_remove", "host_list", "host_info":
		r := hosts.HostsAction(action, name, port)
		fmt.Println(r)
	case "pod_add", "pod_stop", "pod_remove", "pod_list":
		var c []containers.Container
		cmds := strings.Split(cmd, " ")
		var pod = pods.Pod{name, image, count, cores, memory, disk, cmds, c}
		r := pods.PodsAction(action, pod)
		fmt.Println(r)
	default:
		fmt.Println("unknown action " + action)
		panic(conf.Actions)
	}
}
