#!/usr/bin/env bash
set -e

TARGET=target
DISTR_NAME=deploy

echo "Build Binary"
export CGO_ENABLED=0
export GOPATH="/home/netcrk/go"
export GOPROXY="https://artifactorycn.netcracker.com/pd.sandbox-staging.go.group"
export GOSUMDB=off

go build -o ./bin/mongodb-services -gcflags all=-trimpath=${GOPATH} -asmflags all=-trimpath=${GOPATH} ./main.go

docker build -t mongodb-services .
for id in $DOCKER_NAMES
do
    docker tag mongodb-services $id
done

echo "test"

mkdir -p deployments/charts/mongodb-services

gzip -f -c ./charts/helm/mongodb-services/monitoring/mongo7-grafana-dashboard.json > ./charts/helm/mongodb-services/monitoring/mongo7-grafana-dashboard.json.gz
gzip -f -c ./charts/helm/mongodb-services/monitoring/mongo8-grafana-dashboard.json > ./charts/helm/mongodb-services/monitoring/mongo8-grafana-dashboard.json.gz
cp -R ./charts/helm/mongodb-services/* deployments/charts/mongodb-services/
cp ./charts/deployment-configuration.json deployments/deployment-configuration.json
