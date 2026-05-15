#!/usr/bin/env bash
set -e

docker build -t myimage .
for id in $DOCKER_NAMES
do
    docker tag myimage $id
done

