#!/bin/bash

NAME=myapp
docker rm -f $(docker ps -aq)
docker rm dev-peer0.org1.example.com-"$NAME"-1.0
DEL_IMAGE=$(docker images | grep dev)
DEL=(${DEL_IMAGE//\ / })
echo $DEL_IMAGE
docker rmi ${DEL[2]}
DEL_IMAGE=$(docker images | grep dev)
DEL=(${DEL_IMAGE//\ / })
echo $DEL_IMAGE
docker rmi ${DEL[2]}

rm -rf hfc-key-store

