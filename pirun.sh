#!/bin/bash

cd backend || exit

sudo rmmod uvcvideo
sudo modprobe uvcvideo nodrop=1 timeout=5000 quirks=0x80

export VERBOSE=1
export APIHOST="swCloudeLB-313106632.us-east-1.elb.amazonaws.com"
echo $APIHOST
cmd=./backend
$cmd


now=$(date +"%T")
echo "[$now] Running"
