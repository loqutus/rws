package main

import (
	"fmt"
	"os"
)

func storageUpload(name string) error {
	return nil
}

func storageDownload(name string) error {
	return nil
}

func storage(action, name string) {
	switch action {
	case "h":
		fmt.Println("up or down and filename")
	case "u":
		err := storageUpload(name)
		if err != nil {
			panic("storage upload failure")
		}
	case "d":
		err := storageDownload(name)
		if err != nil {
			panic("storage download failure")
		}
	}
}

func mysqlRun(name string) error {
	return nil
}

func mysqlStop(name string) error {
	return nil
}

func mysql(action, name string) {
	switch action {
	case "help":
		fmt.Println("run or stop")
	case "run":
		err := mysqlRun(name)
		if err != nil {
			panic("mysql run error")
		}
	case "stop":
		err := mysqlStop(name)
		if err != nil {
			panic("mysql stop error")
		}
	default:
		fmt.Println("run or stop")

	}
}

func redisRun(name string) error {
	return nil
}

func redisStop(name string) error {
	return nil
}

func redis(action, name string) {
	switch action {
	case "help":
		fmt.Println("run or stop")
	case "run":
		err := redisRun(name)
		if err != nil {
			panic("redis run error")
		}
	case "stop":
		err := redisStop(name)
		if err != nil {
			panic("redis stop error")
		}
	default:
		fmt.Println("run or stop")

	}
}

func printHelp() {
	fmt.Println("storage, mysql or redis")
}

func main() {
	command := os.Args[1]
	switch command {
	case "help":
		printHelp()
	case "storage":
		storage(os.Args[2], os.Args[3])
	case "mysql":
		mysql(os.Args[2], os.Args[3])
	case "redis":
		redis(os.Args[2], os.Args[3])
	default:
		printHelp()
	}
}
