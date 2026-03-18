#!/bin/bash
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

echo "test"