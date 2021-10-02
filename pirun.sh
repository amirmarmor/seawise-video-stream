#!/bin/bash

docker-compose -f ./redis.yaml down
docker-compose -f ./redis.yaml up -d

cd backend
kill -9 $(lsof -t -i:1323)

cmd=./backend>/dev/null
$cmd &

now=$(date +"%T")
echo "[$now] Running"

