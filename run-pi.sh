#!/bin/bash

docker-compose-redis down
docker-compose-redis up

go run ./backend/main.go &

