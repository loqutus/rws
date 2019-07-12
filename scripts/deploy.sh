#!/usr/bin/env bash
set -x
set -e
CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -o ../build/package/server ../cmd/server
cd ../
docker build --no-cache -f build/package/Dockerfile . -t rws
docker tag rws loqutus/rws
docker push loqutus/rws
docker container prune -f
cd deployments
docker-compose -f docker-compose.yml down --remove-orphans
docker-compose -f docker-compose-etcd.yml down --remove-orphans
docker-compose -f docker-compose-etcd.yml up -d
etcdctl mkdir /rws
etcdctl mkdir /rws/hosts
etcdctl mkdir /rws/pods
etcdctl mkdir /rws/containers
etcdctl mkdir /rws/storage
docker-compose -f docker-compose.yml up -d
for i in $(seq 2 5); do
    scp docker-compose.yml pi$i:~/
    ssh pi$i docker-compose -f docker-compose.yml down --remove-orphans
    ssh pi$i docker-compose -f docker-compose.yml up -d
done
sleep 2
cd ../cmd/client
go test
docker logs -f deployments_rws_1
