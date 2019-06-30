package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/loqutus/rws/pkg/server/conf"
	"github.com/loqutus/rws/pkg/server/etcd"
	"github.com/loqutus/rws/pkg/server/hosts"
	"github.com/loqutus/rws/pkg/server/utils"
	"github.com/shirou/gopsutil/disk"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type File struct {
	Name     string
	Host     string
	Size     uint64
	Replicas uint64
}

func GetFileNameFromPath(p string) string {
	PathSplit := strings.Split(p, "/")
	fileName := PathSplit[len(PathSplit)-1]
	return fileName
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(1, "storage.UploadHandler")
	fileName := GetFileNameFromPath(r.URL.Path)
	log.Println(1, "storage.UploadHandler: "+fileName)
	dir, err := etcd.ListDir("/rws/storage")
	if err != nil {
		utils.Fail("storage.UploadHandler: EtcdListDir error", err, w)
	}
	var found = false
	for _, file := range dir {
		keyName := GetFileNameFromPath(file.Key)
		if keyName == fileName {
			found = true
			break
		}
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.Fail("storage.UploadHandler: request reading error", err, w)
		return
	}
	FileSize := len(body)
	FilePathName := conf.DataDir + "/" + fileName
	di, err2 := disk.Usage("/")
	if err2 != nil {
		utils.Fail("storage.UploadHandler: disk usage get error", err2, w)
		return
	}
	if di.Free > uint64(FileSize) {
		if found == true {
			err := os.Remove(fileName)
			if err != nil {
				utils.Fail("storage.UploadHandler: file remove error", err, w)
				return
			}
		}
		err3 := ioutil.WriteFile(FilePathName, []byte(body), 0644)
		if err3 != nil {
			utils.Fail("storage.UploadHandler: file write error", err3, w)
			return
		}
		f := File{fileName, conf.LocalHostName, uint64(FileSize), 1}
		fileBytes, err7 := json.Marshal(f)
		if err7 != nil {
			utils.Fail("storage.UploadHandler: json.Marshal error", err7, w)
			return
		}
		err8 := etcd.CreateKey("/rws/storage/"+fileName, string(fileBytes))
		if err8 != nil {
			utils.Fail("storage.UploadHandler: etcdCreateKey error", err8, w)
			return
		}
		log.Println(1, "storage.UploadHandler: file "+FilePathName+" uploaded")
		w.WriteHeader(http.StatusOK)
		return
	} else {
		hostsListString, err5 := hosts.ListHosts()
		if err5 != nil {
			utils.Fail("storage.UploadHandler: ListHosts error", err5, w)
			return
		}
		var hostsList []hosts.Host
		err4 := json.Unmarshal([]byte(hostsListString), &hostsList)
		if err4 != nil {
			utils.Fail("storage.UploadHandler: JsonUnmarshal error", err4, w)
			return
		}
		for _, host := range hostsList {
			url := "http://" + host.Name + "/host_info"
			body2, err5 := http.Get(url)
			if err5 != nil {
				utils.Fail("storage.UploadHandler: get error", err5, w)
				continue
			}
			bodyBytes, err := ioutil.ReadAll(body2.Body)
			if err != nil {
				utils.Fail("storage.UploadHandler: request reading error", err, w)
				return
			}
			var thatHost hosts.Host
			err6 := json.Unmarshal([]byte(bodyBytes), &thatHost)
			if err6 != nil {
				utils.Fail("storage.UploadHandler: json.Unmarshal error", err6, w)
				return
			}
			if uint64(FileSize) < thatHost.Disk {
				log.Println(1, "storage.UploadHandler: uploading to "+host.Name)
				url := fmt.Sprintf("%s/storage_upload/%s", host.Name, FilePathName)
				dat, err6 := http.Post(url, "application/octet-stream", r.Body)
				if err6 != nil {
					log.Println("StorageUploadHandle: post error: " + url)
					log.Println(dat)
					continue
				}
				log.Println(1, "storage.UploadHandler: "+fileName+" uploaded")
				w.WriteHeader(http.StatusOK)
				_, err = w.Write([]byte("OK"))
				if err != nil {
					log.Println("storage.UploadHandler: response write error")
					log.Println(err)
				}
				return
			} else {
				log.Println(1, "storage.UploadHandler: not enough free space on "+host.Name)
				continue
			}
		}
		log.Println(1, "storage.UploadHandler: unable to upload file "+fileName)
		log.Println(1, http.StatusInternalServerError)
		utils.Fail("storage.UploadHandler: unable to upload file "+fileName, errors.New("storage.UploadHandler: no hosts found to upload file"), w)
		return
	}
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(1, "DownloadHandler")
	fileName := GetFileNameFromPath(r.URL.Path)
	log.Println(1, "DownloadHandler: "+fileName)
	dir, err := etcd.ListDir("/rws/storage")
	if err != nil {
		utils.Fail("DownloadHandler: EtcdListDir error", err, w)
		return
	}
	found := false
	for _, file := range dir {
		keyName := GetFileNameFromPath(file.Key)
		if keyName == fileName {
			found = true
			break
		}
	}
	if found == false {
		utils.Fail("DownloadHandler: file not found", errors.New("file not found"), w)
		return
	}
	fileString, err9 := etcd.GetKey("/rws/storage/" + fileName)
	if err9 != nil {
		utils.Fail("DownloadHandler: EtcdGetKey error", err9, w)
		return
	}
	var file File
	err10 := json.Unmarshal([]byte(fileString), &file)
	if err10 != nil {
		utils.Fail("DownloadHandler: json.Unmarshal error", err10, w)
		return
	}
	if file.Host == conf.LocalHostName {
		dat, err1 := ioutil.ReadFile("data/" + fileName)
		if err1 != nil {
			utils.Fail("DownloadHandler: file read error", err1, w)
			return
		}
		_, err2 := w.Write(dat)
		if err2 != nil {
			utils.Fail("DownloadHandler: request write error", err2, w)
			return
		}
		w.WriteHeader(http.StatusOK)
		log.Println(1, "DownloadHandler: file "+fileName+" downloaded")
		return
	} else {
		url := "http://" + file.Host + "/storage_download/" + file.Name
		body, err3 := http.Get(url)
		if err3 != nil {
			utils.Fail("DownloadHandler: file get error", err3, w)
			return
		}
		bodyBytes, err4 := ioutil.ReadAll(body.Body)
		if err4 != nil {
			utils.Fail("DownloadHandler: body read error", err4, w)
			return
		}
		_, err6 := w.Write(bodyBytes)
		if err6 != nil {
			utils.Fail("DownloadHandler: request write error", err6, w)
			return
		}
		return
	}
}

func RemoveHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(1, "RemoveHandler")
	fileName := GetFileNameFromPath(r.URL.Path)
	log.Println(1, "RemoveHandler: "+fileName)
	dir, err := etcd.ListDir("/rws/storage")
	if err != nil {
		utils.Fail("DownloadHandler: EtcdListDir error", err, w)
		return
	}
	found := false
	for _, f := range dir {
		keyName := GetFileNameFromPath(f.Key)
		if keyName == fileName {
			found = true
			break
		}
	}
	if found == false {
		utils.Fail("DownloadHandler: file not found", errors.New("file not found"), w)
		return
	}
	fileString, err := etcd.GetKey("/rws/storage/" + fileName)
	if err != nil {
		utils.Fail("RemoveHandler: EtcdGetKey error", err, w)
		return
	}
	var file File
	err3 := json.Unmarshal([]byte(fileString), &file)
	if err3 != nil {
		utils.Fail("RemoveHandler: json.Unmarshal error", err3, w)
		return
	}
	if file.Host == conf.LocalHostName {
		err := os.Remove("data/" + fileName)
		if err != nil {
			utils.Fail("RemoveHandler: file remove error", err, w)
			return
		}
		log.Println(1, "RemoveHandler: file "+fileName+" removed locally")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	} else {
		url := "http://" + file.Host + "/storage_remove/" + fileName
		resp, err3 := http.Get(url)
		if err3 != nil {
			utils.Fail("RemoveHandler: file remove get error", err3, w)
			return
		}
		if resp.StatusCode != http.StatusOK {
			utils.Fail("RemoveHandler: file remove get error", err3, w)
			return
		}
		log.Println(1, "RemoveHandler: file "+fileName+" removed from host "+file.Host)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}
	err2 := etcd.DeleteKey("/rws/storage/" + fileName)
	if err2 != nil {
		utils.Fail("RemoveHandler: etcdDeleteKey error", err2, w)
	}
	log.Println(1, "StorageDeleteHandler: "+fileName+" deleted")
	return
}

func StorageList() (string, error) {
	log.Println(1, "StorageList")
	filesNodes, err := etcd.ListDir("/rws/storage")
	if err != nil {
		return "", errors.New("StorageList: EtcdListDir error")
	}
	var l []File
	for _, Key := range filesNodes {
		var x File
		err := json.Unmarshal([]byte(Key.Value), &x)
		if err != nil {
			log.Println(1, "StorageList: json unmarshal error")
			return "", err
		}
		l = append(l, x)
	}
	b, err2 := json.Marshal(l)
	if err2 != nil {
		return "", err2
	}
	return string(b), nil
}

func ListHandler(w http.ResponseWriter, _ *http.Request) {
	log.Println(1, "ListHandler")
	s, err := StorageList()
	if err != nil {
		utils.Fail("ListHandler: StorageList error", err, w)
		return
	}
	_, err2 := w.Write([]byte(s))
	if err2 != nil {
		utils.Fail("ListHandler: request write error", err2, w)
		return
	}
	return
}

func FileSizeHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(1, "FileSizeHandler: storage file size")
	fileName := GetFileNameFromPath(r.URL.Path)
	log.Println(1, "FileSizeHandler: "+fileName)
	found := false
	dir, err := etcd.ListDir("/rws/storage")
	if err != nil {
		utils.Fail("FileSizeHandler: EtcdListDir error", err, w)
		return
	}
	for _, Key := range dir {
		keyName := GetFileNameFromPath(Key.Key)
		if keyName == fileName {
			found = true
		}
	}
	if found == false {
		utils.Fail("FileSizeHandler: file not found", err, w)
		return
	}
	var f File
	key, err := etcd.GetKey("/rws/storage/" + fileName)
	if err != nil {
		utils.Fail("FileSizeHandler: EtcdGetKey error", err, w)
	}
	err2 := json.Unmarshal([]byte(key), &f)
	if err2 != nil {
		utils.Fail("FileSizeHandler: json.Unmarshal error", err2, w)
	}
	_, err3 := w.Write([]byte(strconv.Itoa(int(f.Size))))
	if err3 != nil {
		utils.Fail("FileSizeHandler: request write error", err2, w)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}

func GetFileSize(filename, host, port string) (uint64, error) {
	log.Println(1, "GetFileSize")
	url := fmt.Sprintf("http://%s:%s/storage_file_size/%s", host, port, filename)
	body, err := http.Get(url)
	if err != nil {
		log.Println(1, "get error")
		log.Println(1, body.Body)
		return 0, err
	}
	if body.StatusCode != 200 {
		log.Println(1, body.StatusCode)
		log.Println(1, "status code error")
		return 0, err
	}
	b, err2 := ioutil.ReadAll(body.Body)
	if err2 != nil {
		log.Println(1, err2)
		log.Println(1, "response read error")
		return 0, err2
	}
	i, err3 := strconv.Atoi(string(b))
	if err != nil {
		log.Println(1, "atoi error")
		log.Println(1, err3)
		return 0, err3
	}
	return uint64(i), nil
}
