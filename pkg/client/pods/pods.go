package pods

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/loqutus/rws/pkg/client/containers"
	"github.com/loqutus/rws/pkg/client/utils"
)

type Pod struct {
	Name       string
	Image      string
	Count      uint64
	Cores      uint64
	Memory     uint64
	Disk       uint64
	Cmd        []string
	Containers []containers.Container
}

func PodsAction(action string, pod Pod) string {
	b, err := json.Marshal(pod)
	if err != nil {
		fmt.Println("json marshal error")
		panic(err)
	}
	buf := bytes.NewBuffer(b)
	switch action {
	case "pod_add", "pod_remove", "pod_list", "pod_info":
		resp, err := utils.Req(action, buf)
		if err != nil {
			fmt.Println("post error")
			panic(err)
		}
		return string(resp)
	default:
		panic("unknown action")
	}
}
