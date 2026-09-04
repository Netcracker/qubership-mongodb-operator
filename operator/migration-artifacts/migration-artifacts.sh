#!/usr/bin/env bash
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


#Required for DP job to obtain parameters
eval $(sed -e 's/:[^:\/\/]/="/g;s/$/"/g;s/ *=/=/g' <<<"$DEPLOYMENT_PARAMETERS" | grep "ENABLE_MIGRATION\|CUSTOM_RESOURCE_NAME")

ENABLE_MIGRATION=${ENABLE_MIGRATION:-true}
SERVICE_NAME=${SERVICE_NAME:-mongodb-operator}
SELECTOR=${SELECTOR:-app=mongo-cluster}

echo "ENABLE_MIGRATION: ${ENABLE_MIGRATION}"
if [[ ${ENABLE_MIGRATION} != "true" ]]; then
  exit 0
fi

if command -v kubectl &>/dev/null; then
  kubectl="kubectl"
else
  source ${WORKSPACE}/oc_version_used.sh
  kubectl="${OCBINVERP}"
fi

if command -v helm &>/dev/null; then
  helm="helm"
else
  helm="helm3"
fi

echo "Start migration procedure"

if ! ($helm list | grep ${SERVICE_NAME}); then
  echo "There are no ${SERVICE_NAME} helm releases. Please perform manual migration"
  exit 0
fi

if ! $kubectl get mongoservices operator; then
  echo "mongoservice does not exist. No migration needed."
  exit 0
fi


echo "Removing the linkage between deployments and services to keep them running during migration"
$kubectl get statefulsets -o name | xargs -I {} $kubectl patch {} -p '{"metadata":{"ownerReferences":null}}'

# $helm uninstall ${SERVICE_NAME}
# $helm uninstall ${SERVICE_NAME}-${NAMESPACE}
# echo "releases has been deleted, wait 10s for terminating resources"
# sleep 10

echo "End migration procedure"