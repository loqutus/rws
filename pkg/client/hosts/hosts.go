package hosts

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/loqutus/rws/pkg/client/utils"
)

type Host struct {
	Name string
	Port uint64
}

func HostsAction(action string, hostName string, hostPort uint64) string {
	var resp []byte
	var err error
	h := Host{hostName, hostPort}
	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(h)
	if err != nil {
		fmt.Println(err)
		panic("json encoding error")
	}
	switch action {
	case "host_add", "host_remove", "host_list", "host_info":
		resp, err = utils.Req(action, b)
		if err == errors.New("host already exists") {
			fmt.Println("host already exists")
		}
		if err != nil {
			fmt.Println(err)
			panic("get error")
		}
		return string(resp)
	default:
		panic("unknown action")
	}
}
