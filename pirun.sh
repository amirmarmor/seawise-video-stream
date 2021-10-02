#!/bin/bash

docker stop my-redis
docker rm my-redis
docker run --name my-redis -p "6379:6379" -d redis redis-server --appendonly yes

cd backend
kill -9 $(lsof -t -i:1323)

cmd=./backend>/dev/null
$cmd &

now=$(date +"%T")
echo "[$now] Running"

