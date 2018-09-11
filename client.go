package main

import (
	"fmt"
	"os"
)

func storage_upload(filename string) error {
	return nil
}

func storage_download(filename string) error {
	return nil
}

func storage(action, filename string) {
	switch action {
	case "h":
		fmt.Println("up or down and filename")
	case "u":
		err := storage_upload(filename)
		if err != nil {
			panic("storage upload failure")
		}
	case "d":
		err := storage_download(filename)
		if err != nil {
			panic("storage download failure")
		}
	}
}

func container_run(name, image string) (string, error) {
	return "", nil
}

func container_stop(name string) error {
	return nil
}

func container(action, name, image string) {
	switch action {
	case "h":
		fmt.Println("run, stop, name and image if run")
	case "r":
		str, err := container_run(name, image)
		if err != nil {
			panic("container run error")
		}
		fmt.Println(str)
	case "s":
		err := container_stop(name)
		if err != nil {
			panic("container stop error")
		}
	}
}

func main() {
	command := os.Args[1]
	switch command {
	case "h":
		fmt.Println("s for storage, c for container, d for database, r for redis, l for load balancer, h for help")
	case "s":
		storage(os.Args[2], os.Args[3])
	case "c":
		container(os.Args[2], os.Args[3], "")
	default:
		fmt.Println("specify s,c,d,r or l")
	}
}
