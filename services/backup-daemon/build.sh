#!/usr/bin/env bash
set -e

# #Making sure working directory is the same as the current file location
# cd "$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

# export DIR=$(pwd)

# echo "Current dir is ${DIR}"

# export MONGO_TGZ_URL="${MONGO_TGZ_URL:-https://artifactorycn.netcracker.com/nc.thirdparty.files/mongodb/mongodb/mongodb-linux-x86_64-rhel70-3.4.19.tgz}"

# echo "Mongo version url: ${MONGO_TGZ_URL}"

# sed -i 's~MONGO_TGZ_URL~'${MONGO_TGZ_URL}'~g' Dockerfile

docker build -t myimage .
for id in $DOCKER_NAMES
do
    docker tag myimage $id
done

