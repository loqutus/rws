#!/usr/bin/env bash
set -x
set -e
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build .
docker build -f Dockerfile-local . -t rws-local
docker tag rws-local loqutus/rws-local
docker container prune -f
docker-compose -f docker-compose-local.yml down --remove-orphans
docker-compose -f docker-compose-local.yml up -d
sleep 1
#export ETCD_UNSUPPORTED_ARCH=arm
etcdctl mkdir /rws
etcdctl mkdir /rws/hosts
etcdctl mkdir /rws/pods
etcdctl mkdir /rws/containers
etcdctl mkdir /rws/storage
cd ../client
go test