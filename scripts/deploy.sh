#!/usr/bin/env bash
set -x
set -e
start=`date +%s`
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
s(){
    scp docker-compose.yml pi$1:~/
    ssh pi$1 docker-compose -f docker-compose.yml down --remove-orphans
    ssh pi$1 docker pull loqutus/rws
    ssh pi$1 docker-compose -f docker-compose.yml up -d
}
s 2 &
s 3 &
s 4 &
s 5 &
wait
sleep 2
cd ../cmd/client
go test
end=`date +%s`
echo $((end-start))
docker logs -f deployments_rws_1
