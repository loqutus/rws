package hosts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/loqutus/rws/pkg/client/utils"
)

type Host struct {
	Name string
}

func HostsAction(action, hostName, hostPort string) string {
	var resp []byte
	var err error
	h := Host{hostName + ":" + hostPort}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(h)
	switch action {
	case "host_add", "host_remove", "host_list", "host_info":
		resp, err = utils.Req(action, b)
		if err != nil {
			fmt.Println(err)
			panic("get error")
		}
		return string(resp)
	default:
		panic("unknown action")
	}
}
