#!/usr/bin/env bash
set -e
set -x

echo "Build Binary"
export CGO_ENABLED=0
export GOPATH="/home/netcrk/go"
echo "GOPATH=${GOPATH}"
export GOPROXY="https://artifactorycn.netcracker.com/pd.sandbox-staging.go.group"

go build -o ./bin/dbaas-mongo-adapter -gcflags all=-trimpath=${GOPATH} -asmflags all=-trimpath=${GOPATH} ./

docker build -t dbaas_mongo_adapter .
for id in $DOCKER_NAMES
do
    docker tag dbaas_mongo_adapter $id
done

echo "test"