#!/bin/bash

docker-compose -f ./redis.yaml down
docker-compose -f ./redis.yaml -d up

cd backend
kill -9 $(lsof -t -i:1323)

cmd=./backend>/dev/null
$cmd &

now=$(date +"%T")
echo "[$now] Running"

