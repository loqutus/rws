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

func StorageUploadHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(1, "StorageUploadHandler")
	fileName := GetFileNameFromPath(r.URL.Path)
	log.Println(1, "StorageUploadHandler: "+fileName)
	dir, err := etcd.ListDir("/rws/storage")
	if err != nil {
		utils.Fail("StorageUploadHandler: EtcdListDir error", err, w)
	}
	for _, file := range dir {
		keyName := GetFileNameFromPath(file.Key)
		if keyName == fileName {
			utils.Fail("StorageUploadHandler: file already exists", errors.New("file already exists"), w)
			return
		}
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.Fail("StorageUploadHandler: request reading error", err, w)
		return
	}
	fmt.Println("-----")
	fmt.Println(body)
	FileSize := len(body)
	FilePathName := conf.DataDir + "/" + fileName
	di, err2 := disk.Usage("/")
	if err2 != nil {
		utils.Fail("StorageUploadHandler: disk usage get error", err2, w)
		return
	}
	if di.Free > uint64(FileSize) {
		err3 := ioutil.WriteFile(FilePathName, []byte(body), 0644)
		if err3 != nil {
			utils.Fail("StorageUploadHandler: file write error", err3, w)
			return
		}
		f := File{fileName, conf.LocalHostName, uint64(FileSize), 1}
		fileBytes, err7 := json.Marshal(f)
		if err7 != nil {
			utils.Fail("StorageUploadHandler: json.Marshal error", err7, w)
			return
		}
		err8 := etcd.CreateKey("/rws/storage/"+fileName, string(fileBytes))
		if err8 != nil {
			utils.Fail("StorageUploadHandler: etcdCreateKey error", err8, w)
			return
		}
		log.Println(1, "StorageUploadHandler: file "+FilePathName+" uploaded")
		w.WriteHeader(http.StatusOK)
		return
	} else {
		hostsListString, err5 := hosts.ListHosts()
		if err5 != nil {
			utils.Fail("StorageUploadHandler: ListHosts error", err5, w)
			return
		}
		var hostsList []hosts.Host
		err4 := json.Unmarshal([]byte(hostsListString), &hostsList)
		if err4 != nil {
			utils.Fail("StorageUploadHandler: JsonUnmarshal error", err4, w)
			return
		}
		for _, host := range hostsList {
			url := "http://" + host.Name + "/host_info"
			body2, err5 := http.Get(url)
			if err5 != nil {
				utils.Fail("StorageUploadHandler: get error", err5, w)
				continue
			}
			bodyBytes, err := ioutil.ReadAll(body2.Body)
			if err != nil {
				utils.Fail("StorageUploadHandler: request reading error", err, w)
				return
			}
			var thatHost hosts.Host
			err6 := json.Unmarshal([]byte(bodyBytes), &thatHost)
			if err6 != nil {
				utils.Fail("StorageUploadHandler: json.Unmarshal error", err6, w)
				return
			}
			if uint64(FileSize) < thatHost.Disk {
				log.Println(1, "StorageUploadHandler: uploading to "+host.Name)
				url := fmt.Sprintf("%s/storage_upload/%s", host.Name, FilePathName)
				dat, err6 := http.Post(url, "application/octet-stream", r.Body)
				if err6 != nil {
					log.Println("StorageUploadHandle: post error: " + url)
					log.Println(dat)
					continue
				}
				log.Println(1, "StorageUploadHandler: "+fileName+" uploaded")
				w.WriteHeader(http.StatusOK)
				_, err = w.Write([]byte("OK"))
				if err != nil {
					log.Println("StorageUploadHandler: response write error")
					log.Println(err)
				}
				return
			} else {
				log.Println(1, "StorageUploadHandler: not enough free space on "+host.Name)
				continue
			}
		}
		log.Println(1, "StorageUploadHandler: unable to upload file "+fileName)
		log.Println(1, http.StatusInternalServerError)
		utils.Fail("StorageUploadHandler: unable to upload file "+fileName, errors.New("StorageUploadHandler: no hosts found to upload file"), w)
		return
	}
}

func StorageDownloadHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(1, "StorageDownloadHandler")
	fileName := GetFileNameFromPath(r.URL.Path)
	log.Println(1, "StorageDownloadHandler: "+fileName)
	dir, err := etcd.ListDir("/rws/storage")
	if err != nil {
		utils.Fail("StorageDownloadHandler: EtcdListDir error", err, w)
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
		utils.Fail("StorageDownloadHandler: file not found", errors.New("file not found"), w)
		return
	}
	fileString, err9 := etcd.GetKey("/rws/storage/" + fileName)
	if err9 != nil {
		utils.Fail("StorageDownloadHandler: EtcdGetKey error", err9, w)
		return
	}
	var file File
	err10 := json.Unmarshal([]byte(fileString), &file)
	if err10 != nil {
		utils.Fail("StorageDownloadHandler: json.Unmarshal error", err10, w)
		return
	}
	if file.Host == conf.LocalHostName {
		dat, err1 := ioutil.ReadFile("data/" + fileName)
		if err1 != nil {
			utils.Fail("StorageDownloadHandler: file read error", err1, w)
			return
		}
		_, err2 := w.Write(dat)
		if err2 != nil {
			utils.Fail("StorageDownloadHandler: request write error", err2, w)
			return
		}
		w.WriteHeader(http.StatusOK)
		log.Println(1, "StorageDownloadHandler: file "+fileName+" downloaded")
		return
	} else {
		url := "http://" + file.Host + "/storage_download/" + file.Name
		body, err3 := http.Get(url)
		if err3 != nil {
			utils.Fail("StorageDownloadHandler: file get error", err3, w)
			return
		}
		bodyBytes, err4 := ioutil.ReadAll(body.Body)
		if err4 != nil {
			utils.Fail("StorageDownloadHandler: body read error", err4, w)
			return
		}
		_, err6 := w.Write(bodyBytes)
		if err6 != nil {
			utils.Fail("StorageDownloadHandler: request write error", err6, w)
			return
		}
		return
	}
}

func StorageRemoveHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(1, "StorageRemoveHandler")
	fileName := GetFileNameFromPath(r.URL.Path)
	log.Println(1, "StorageRemoveHandler: "+fileName)
	dir, err := etcd.ListDir("/rws/storage")
	if err != nil {
		utils.Fail("StorageDownloadHandler: EtcdListDir error", err, w)
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
		utils.Fail("StorageDownloadHandler: file not found", errors.New("file not found"), w)
		return
	}
	fileString, err := etcd.GetKey("/rws/storage/" + fileName)
	if err != nil {
		utils.Fail("StorageRemoveHandler: EtcdGetKey error", err, w)
		return
	}
	var file File
	err3 := json.Unmarshal([]byte(fileString), &file)
	if err3 != nil {
		utils.Fail("StorageRemoveHandler: json.Unmarshal error", err3, w)
		return
	}
	if file.Host == conf.LocalHostName {
		err := os.Remove("data/" + fileName)
		if err != nil {
			utils.Fail("StorageRemoveHandler: file remove error", err, w)
			return
		}
		log.Println(1, "StorageRemoveHandler: file "+fileName+" removed locally")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	} else {
		url := "http://" + file.Host + "/storage_remove/" + fileName
		resp, err3 := http.Get(url)
		if err3 != nil {
			utils.Fail("StorageRemoveHandler: file remove get error", err3, w)
			return
		}
		if resp.StatusCode != http.StatusOK {
			utils.Fail("StorageRemoveHandler: file remove get error", err3, w)
			return
		}
		log.Println(1, "StorageRemoveHandler: file "+fileName+" removed from host "+file.Host)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}
	err2 := etcd.DeleteKey("/rws/storage/" + fileName)
	if err2 != nil {
		utils.Fail("StorageRemoveHandler: etcdDeleteKey error", err2, w)
	}
	log.Println(1, "StorageDeleteHandler: "+fileName+" deleted")
	return
}

func StorageList() (string, error) {
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

func StorageListHandler(w http.ResponseWriter, _ *http.Request) {
	log.Println(1, "StorageListHandler")
	s, err := StorageList()
	if err != nil {
		utils.Fail("StorageListHandler: StorageList error", err, w)
		return
	}
	_, err2 := w.Write([]byte(s))
	if err2 != nil {
		utils.Fail("StorageListHandler: request write error", err2, w)
		return
	}
	return
}

func StorageFileSizeHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(1, "StorageFileSizeHandler: storage file size")
	fileName := GetFileNameFromPath(r.URL.Path)
	log.Println(1, "StorageFileSizeHandler: "+fileName)
	found := false
	dir, err := etcd.ListDir("/rws/storage")
	if err != nil {
		utils.Fail("StorageFileSizeHandler: EtcdListDir error", err, w)
		return
	}
	for _, Key := range dir {
		keyName := GetFileNameFromPath(Key.Key)
		if keyName == fileName {
			found = true
		}
	}
	if found == false {
		utils.Fail("StorageFileSizeHandler: file not found", err, w)
		return
	}
	var f File
	key, err := etcd.GetKey("/rws/storage/" + fileName)
	if err != nil {
		utils.Fail("StorageFileSizeHandler: EtcdGetKey error", err, w)
	}
	err2 := json.Unmarshal([]byte(key), &f)
	if err2 != nil {
		utils.Fail("StorageFileSizeHandler: json.Unmarshal error", err2, w)
	}
	_, err3 := w.Write([]byte(strconv.Itoa(int(f.Size))))
	if err3 != nil{
		utils.Fail("StorageFileSizeHandler: request write error", err2, w)
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

