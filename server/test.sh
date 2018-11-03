#!/usr/bin/env bash
set -x
kill $(ps aux | grep 'server' | grep -v grep | grep -v mosh | awk '{ print $2 }')
for ID in $(ps aux | grep "ssh -f" | grep  -v grep | awk '{ print $2 }'); do
  kill $ID
done
go build
for i in $(seq 2 5); do
  ssh pi$i pkill server
  rsync server pi$i:~/
  ssh -f pi$i ~/server
done
./server 2>&1 &
cd ../client
go build
for i in $(seq 1 5); do
  ./client --action host_add --name pi$i --port 8888
  ./client --action storage_upload --name $i.txt --hostname "http://pi$i:8888"
  sleep 1
done
./client --action container_run --name test --image "arv32v7/busybox"
#curl http://127.0.0.1:8888/
#kill $(ps aux | grep 'server' | grep -v grep | grep -v mosh | awk '{ print $2 }')
#for ID in $(ps aux | grep "ssh -f" | grep  -v grep | awk '{ print $2 }'); do
#  kill $ID
#done
cd ../server/