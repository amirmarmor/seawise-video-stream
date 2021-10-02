#!/bin/bash

docker-compose -f ./redis.yaml down
docker-compose -f ./redis.yaml up -d

cd backend
kill -9 $(lsof -t -i:1323)
rmmod uvcvideo
modprobe uvcvideo nodrop=1 timeout=5000 quirks=0x80

cmd=./backend>/dev/null
$cmd &

now=$(date +"%T")
echo "[$now] Running"

