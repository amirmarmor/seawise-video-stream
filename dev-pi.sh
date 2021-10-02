#!/bin/bash

docker-compose down
docker-compose -f docker-compose-redis.yaml up -d

go run ./backend/main.go &