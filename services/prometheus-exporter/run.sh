#!/bin/bash
# Copyright 2024-2025 NetCracker Technology Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

if [[ ! -z $DEBUG ]] && [[ ! -z ${DEBUG+x} ]]; then
	set -x
fi

read_secret() {
  local path="$1"
  local fallback="$2"

  if [ -f "$path" ]; then
    echo "Reading secret from file: $path" >&2
    cat "$path"
  else
    echo "Secret file not found: $path, using fallback value" >&2
    echo "$fallback"
  fi
}

PROM_EXPORTER_USER="$(read_secret /var/run/secrets/mongodb/prom-exporter/username admin)"
PROM_EXPORTER_PASSWORD="$(read_secret /var/run/secrets/mongodb/prom-exporter/password admin)"


export HTTP_AUTH="$PROM_EXPORTER_USER:$PROM_EXPORTER_PASSWORD"

EXPORT_MONGOS="${EXPORT_MONGOS:-true}"
MONGO_MONITORING_USER="$(read_secret /var/run/secrets/mongodb/mongo-monitoring/username monitoring)"
MONGO_MONITORING_PASSWORD="$(read_secret /var/run/secrets/mongodb/mongo-monitoring/password monitoring)"

MONGOS_URI="${MONGOS_URI:-mongodb://$MONGO_MONITORING_USER:$MONGO_MONITORING_PASSWORD@mongos:27017}"

TLS_ENABLED=${TLS_ENABLED:-false}

collectors="\
  --collector.replicasetstatus \
  --collector.replicasetconfig \
  --collector.topmetrics \
  --collector.diagnosticdata \
  --collector.currentopmetrics \
  --collector.fcv \
  --collector.profile"

mongosCollectors="\
  --collector.diagnosticdata \
  --collector.replicasetstatus \
  --collector.replicasetconfig \
  --collector.dbstats \
  --collector.dbstatsfreestorage \
  --collector.topmetrics \
  --collector.currentopmetrics \
  --collector.indexstats \
  --collector.collstats=false \
  --collector.profile \
  --collector.fcv \
  --collector.shards \
  --collector.pbm"
      

if [[ ${TLS_ENABLED} = true ]]; then
    MONGOS_URI=("$MONGOS_URI/?tls=true&tlsCAFile=$TLS_ROOTCERT")

	# remove before using production certs
	MONGOS_URI=("$MONGOS_URI&tlsInsecure=true")
fi

shard_members=$(echo "$MONGO_URI" | jq -r .shardMembers)
number_of_shards=$(echo "$MONGO_URI" | jq -r .shardMembers | jq -r 'length')
if [[ $? != 0 ]]; then
	echo "!! Couldn't parse shard_members"
	exit 1
fi

cnfrs_member=$(echo "$MONGO_URI" | jq -r .cnfrsMembers)

if [[ $EXPORT_MONGOS = true ]]; then
	echo "Starting exporter for mongos"
	exporter_cmd="/opt/mongodb_exporter --mongodb.uri=$MONGOS_URI --compatible-mode $mongosCollectors --web.listen-address=:9216 --mongodb.direct-connect=false"
	if [[ "$cnfrs_member" == "" || "$cnfrs_member" == "null" ]] && [[ "$shard_members" == "" || "$shard_members" == "null" ]]; then
	eval $exporter_cmd
	else
	eval $exporter_cmd &
	fi
fi

cnfrs_member=$(echo "$MONGO_URI" | jq -r .cnfrsMembers)
if [[ $cnfrs_member != "" && $cnfrs_member != "null" ]]; then 
	echo "Starting exporter for $cnfrs_member on port 9217"
	mongodb_uri="mongodb://$MONGO_MONITORING_USER:$MONGO_MONITORING_PASSWORD@$cnfrs_member"
	if [[ ${TLS_ENABLED} = true ]]; then
	  mongodb_uri="$mongodb_uri/?tls=true&lsCAFile=$TLS_ROOTCERT&tlsInsecure=true"
	fi
	/opt/mongodb_exporter --mongodb.uri="$mongodb_uri" --compatible-mode --collect-all --mongodb.direct-connect=false --web.listen-address=:9217 &
else
	echo "CNFRS members list is empty"
fi

if [[ $shard_members != "" ]]; then
	port=9218
	for((i=0;i<number_of_shards;i++)); do
	  shard=$(echo "$shard_members" | jq -r .[$i])
	  if [[ $? != 0 ]]; then
		echo "!! Couldn't get shard member from shard_members"
		exit 1
	  fi
	  echo "Starting exporter for $shard on port" $((port+i))
	  mongodb_uri="mongodb://$MONGO_MONITORING_USER:$MONGO_MONITORING_PASSWORD@$shard"
	  if [[ ${TLS_ENABLED} = true ]]; then
		mongodb_uri="$mongodb_uri/?tls=true&tlsCAFile=$TLS_ROOTCERT&tlsInsecure=true"
	  fi
	  if [[ $number_of_shards != $((i+1)) ]]; then
	  /opt/mongodb_exporter --mongodb.uri="$mongodb_uri" --compatible-mode $collectors --mongodb.direct-connect=false --web.listen-address=:$((port+i)) &
	  else
	  /opt/mongodb_exporter --mongodb.uri="$mongodb_uri" --compatible-mode $collectors --mongodb.direct-connect=false --web.listen-address=:$((port+i))
	  fi
	done
else
	echo "DATARS members list is empty"
fi