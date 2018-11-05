#!/usr/bin/env bash
set -x
set -e
CGO_ENABLED=0 go build .
docker build . -t rws
docker tag rws loqutus/rws
docker push loqutus/rws
docker-compose up -d
for i in $(seq 2 5); do
    ssh pi$i docker-compose up -d
done
cd ../client
go build
for i in $(seq 1 5); do
  ./client --action host_add --name pi$i --port 8888
  ./client --action storage_upload --name $i.txt --hostname "http://pi$i:8888"
  sleep 1
done
./client --action container_stop --name test
./client --action container_remove --name test
./client --action container_run --name test --image "arm32v6/alpine"
./client --action pod_add --name test --image "arm32v6/alpine"
./client --action pod_list
for i in $(seq 1 5); do
    docker stop -n rws
done
cd ../server/