package utils

import (
	"bytes"
	"fmt"
	"github.com/loqutus/rws/pkg/client/conf"
	"io/ioutil"
	"net/http"
)

func Req(action string, bodyBuffer *bytes.Buffer) ([]byte, error) {
	// http://localhost:8888/container_add
	url := fmt.Sprintf("%s/%s", conf.HostName, action)
	resp, err1 := http.Post(url, "application/json", bodyBuffer)
	if err1 != nil {
		fmt.Println(err1)
		panic("request error")
	}
	if resp.StatusCode != 200 {
		fmt.Println(resp.StatusCode)
		fmt.Println(resp)
		panic("request status code error")
	}
	b, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		fmt.Println(err2)
		panic("response read error")
	}
	return b, nil
}
