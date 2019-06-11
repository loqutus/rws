#!/usr/bin/env bash
set -x
set -e
CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -o ../build/package/server ../cmd/server
cd ../
docker build --no-cache -f build/package/Dockerfile . -t rws
docker tag rws-local loqutus/rws-local
docker push rws
docker container prune -f
cd deployments
docker-compose -f docker-compose.yml down --remove-orphans
docker-compose -f docker-compose.yml up -d
sleep 1
etcdctl mkdir /rws
etcdctl mkdir /rws/hosts
etcdctl mkdir /rws/pods
etcdctl mkdir /rws/containers
etcdctl mkdir /rws/storage
docker logs -f deployments_rws_1
