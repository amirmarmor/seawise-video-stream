#!/bin/bash

docker-compose -f ./redis.yaml down
docker-compose -f ./redis.yaml up -d

cd backend || exit
# shellcheck disable=SC2046
kill -9 $(lsof -t -i:1323)
sudo rmmod uvcvideo
sudo modprobe uvcvideo nodrop=1 timeout=5000 quirks=0x80

cmd=./backend>/dev/null
$cmd &

now=$(date +"%T")
echo "[$now] Running"

