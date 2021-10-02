#!/bin/bash

docker-compose -f ./redis.yaml down
kill -9 $(lsof -t -i:1323)

now=$(date +"%T")
echo "[$now] Stopped"

