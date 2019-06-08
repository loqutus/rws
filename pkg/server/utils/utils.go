package utils

import (
	"fmt"
	"log"
	"net/http"
)

func Fail(str string, err error, w http.ResponseWriter) {
	log.Println(1, str)
	log.Println(1, err.Error())
	w.WriteHeader(http.StatusInternalServerError)
	_, err = w.Write([]byte(str))
	if err != nil {
		fmt.Println("response write error")
		log.Println(err)
		return
	}
	_, err = w.Write([]byte(err.Error()))
	if err != nil {
		fmt.Println("response write error")
		log.Println(err)
		return
	}
	w.WriteHeader(500)
}
