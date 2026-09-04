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

set -x
set -e

for docker_image_name in ${DOCKER_NAMES}; do
  docker build \
    --file=Dockerfile \
    --pull \
    -t ${docker_image_name} \
    .
done

# Integration tests artifacts
BUILD_SH_TARGET_DIR=target

DEPLOY_ARTIFACT_NAME=mongodb-integration-tests-artifacts

DEPLOY_JOB_PATH_INTEGRATION_TESTS=jenkins/jobs/mongodb-integration-tests

mkdir -p ${BUILD_SH_TARGET_DIR}


zip -r \
  ${BUILD_SH_TARGET_DIR}/${DEPLOY_ARTIFACT_NAME}.zip \
  openshift ${DEPLOY_JOB_PATH_INTEGRATION_TESTS} integration-tests
