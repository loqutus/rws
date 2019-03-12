package containers

import (
	"bytes"
	"fmt"
	"encoding/json"
	"github.com/loqutus/rws/pkg/client/utils"
)

type Container struct {
	Image  string
	Name   string
	Disk   uint64
	Memory uint64
	Cores  uint64
	Host   string
	ID     string
	Cmd    []string
}

func ContainerAction(action, image, name string, cmd []string) string {
	var err error
	var resp []byte
	c := Container{image, name, 1, 1, 1, "", "", cmd}
	b, err2 := json.Marshal(c)
	if err2 != nil {
		fmt.Println(err2)
		panic("json Marshal error")
	}
	buf := bytes.NewBuffer(b)
	switch action {
	case "container_list", "container_run", "container_stop", "container_remove", "container_list_local":
		resp, err = utils.Req(action, buf)
	default:
		panic("unknown action")
	}
	if err != nil {
		fmt.Println(err)
		panic("get error")
	}
	return string(resp)
}