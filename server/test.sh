#!/usr/bin/env bash
set -x
set -e
CGO_ENABLED=0 go build .
docker build . -t rws
docker tag rws loqutus/rws
docker push loqutus/rws
docker container prune -f
docker-compose  -f docker-compose-etcd.yml down --remove-orphans
docker-compose  -f docker-compose-etcd.yml up -d
export ETCD_UNSUPPORTED_ARCH=arm
etcdctl mkdir /rws
etcdctl mkdir /rws/hosts
etcdctl mkdir /rws/pods
etcdctl mkdir /rws/containers
etcdctl mkdir /rws/storage
for i in $(seq 1 5); do
    scp docker-compose.yml pi$i:~/
    ssh pi$i docker-compose down --remove-orphans
    ssh pi$i docker pull loqutus/rws
    ssh pi$i docker-compose up -d
done
cd ../client
go build
for i in $(seq 1 5); do
  ./client --action host_add --name 10.0.0.$i --port 8888
  ./client --action storage_upload --name $i.txt --hostname "http://10.0.0.$i:8888"
  sleep 1
done
etcdctl ls /rws/hosts/
etcdctl ls /rws/storage/
./client --action container_run --name test --image "arm32v6/alpine" --cmd "/bin/sleep 60"
etcdctl ls /rws/containers/
sleep 5
./client --action container_stop --name test
docker ps | grep sleep
./client --action container_remove --name test
docker ps -a | grep sleep
./client --action pod_add --name test --image "arm32v6/alpine" --cmd "/bin/sleep 60"
./client --action pod_list
#for i in $(seq 1 5); do
#    ssh pi$i docker-compose down
#done
#cd ../server/