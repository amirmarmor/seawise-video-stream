#!/bin/bash

cd backend || exit

sudo rmmod uvcvideo
sudo modprobe uvcvideo nodrop=1 timeout=5000 quirks=0x80

export VERBOSE=1
cmd=./backend>/dev/null
$cmd &


now=$(date +"%T")
echo "[$now] Running"
