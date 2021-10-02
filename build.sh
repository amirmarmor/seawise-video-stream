#!/bin/bash

cd backend
go build -o backend
now=$(date +"%T")
echo "[$now] build done"

