package main

import (
	"fmt"
	"testing"
)

func TestStorageUpload(t *testing.T) {
	err := storageUpload("test")
	if err != nil {
		fmt.Println(err)
		t.Errorf("error")
	}
}
