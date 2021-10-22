#!/bin/bash

cd backend
go build -o streamer
now=$(date +"%T")
echo "[$now] build done"

