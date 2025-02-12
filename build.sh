#!/usr/bin/env bash
set -e

TARGET=target
DISTR_NAME=deploy

echo "Build Binary"
export CGO_ENABLED=0
export GOPATH="/home/netcrk/go"
export GOPROXY="https://artifactorycn.netcracker.com/pd.sandbox-staging.go.group"
export GOSUMDB=off

go build -o ./bin/mongo-operator -gcflags all=-trimpath=${GOPATH} -asmflags all=-trimpath=${GOPATH} ./main.go

docker build -t mongo_operator .
for id in $DOCKER_NAMES
do
    docker tag mongo_operator $id
done

SCRIPTS=scripts
DIST_FILE="${SCRIPTS}/migration-artifacts.zip"
DIST_CONTENT="migration-artifacts"

rm -rf ./${SCRIPTS}
mkdir ${SCRIPTS}
zip -qr "$DIST_FILE" "$DIST_CONTENT"

mkdir -p deployments/charts/mongodb-operator

cp -R ./charts/helm/mongodb-operator/* deployments/charts/mongodb-operator/
cp ./charts/deployment-configuration.json deployments/deployment-configuration.json
