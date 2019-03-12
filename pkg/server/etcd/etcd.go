package etcd

import (
	etcdClient "go.etcd.io/etcd/client"
	"golang.org/x/net/context"
	"log"
	"time"
)

const Host = "http://etcd:2379"

var Client etcdClient.Client

func init() {
	etcdCfg := etcdClient.Config{
		Endpoints: []string{Host},
		Transport: etcdClient.DefaultTransport,
	}
	var err error
	Client, err = etcdClient.New(etcdCfg)
	if err != nil {
		log.Println(err)
		panic("etcd client initialization error")
	}
}

func CreateKey(name, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	kAPI := etcdClient.NewKeysAPI(Client)
	_, err := kAPI.Create(ctx, name, value)
	return err
}

func SetKey(name, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	kAPI := etcdClient.NewKeysAPI(Client)
	_, err := kAPI.Set(ctx, name, value, nil)
	return err
}

func DeleteKey(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	kAPI := etcdClient.NewKeysAPI(Client)
	_, err := kAPI.Delete(ctx, name, nil)
	return err
}

func GetKey(name string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	kAPI := etcdClient.NewKeysAPI(Client)
	resp, err := kAPI.Get(ctx, name, nil)
	if err != nil {
		return "", err
	}
	return resp.Node.Value, nil
}

func ListDir(name string) (etcdClient.Nodes, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	kAPI := etcdClient.NewKeysAPI(Client)
	resp, err := kAPI.Get(ctx, name, nil)
	if err != nil {
		return nil, err
	}
	return resp.Node.Nodes, nil
}
