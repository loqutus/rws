#!/usr/bin/env bash
set -x
set -e
CGO_ENABLED=0 go build .
docker build . -t rws
docker tag rws loqutus/rws
docker push loqutus/rws
docker container prune -f
docker-compose down --remove-orphans
docker-compose  -f docker-compose-etcd.yml down --remove-orphans
docker-compose  -f docker-compose-etcd.yml up -d
sleep 5
export ETCD_UNSUPPORTED_ARCH=arm
etcdctl mkdir /rws
etcdctl mkdir /rws/hosts
etcdctl mkdir /rws/pods
etcdctl mkdir /rws/containers
etcdctl mkdir /rws/storage
docker-compose up -d
for i in $(seq 2 5); do
    scp docker-compose.yml pi$i:~/ &
done
wait
for i in $(seq 2 5); do
    ssh pi$i docker-compose down --remove-orphans &
done
wait
for i in $(seq 2 5); do
    ssh pi$i docker pull loqutus/rws &
done
wait
for i in $(seq 2 5); do
    ssh pi$i docker-compose up -d &
done
wait
cd ../client
go test
#for i in $(seq 1 5); do
#  ./client --action host_add --name 10.0.0.$i --port 8888
#  ./client --action storage_upload --name $i.txt --hostname "http://10.0.0.$i:8888"
#  sleep 1
#done
#./client --action container_run --name test --image "arm32v6/alpine" --cmd "/bin/sleep 60"
#./client --action container_list
#./client --action container_stop --name test
#./client --action container_remove --name test
#./client --action pod_add --name test_pod --image "arm32v6/alpine" --cmd "/bin/sleep 60"
#./client --action pod_list
