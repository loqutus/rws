package info

import (
	"github.com/loqutus/rws/pkg/server/containers"
	"github.com/loqutus/rws/pkg/server/hosts"
	"github.com/loqutus/rws/pkg/server/pods"
	"github.com/loqutus/rws/pkg/server/storage"
)

type Info struct {
	Storage    []storage.File
	Hosts      []hosts.Host
	Pods       []pods.Pod
	Containers []containers.Container
}
