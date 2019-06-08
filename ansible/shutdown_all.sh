#!/usr/bin/env bash
set -x
for i in $(seq 2 5); do
  ssh root@pi$i shutdown -h now
done
sleep 5
sudo shutdown -h now
