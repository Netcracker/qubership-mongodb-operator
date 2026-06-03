#!/bin/bash
set -x
set -e

rm -rf /var/jenkins_home/workspace/mongodb-integration-tests/target
mkdir /var/jenkins_home/workspace/mongodb-integration-tests/target/

if [[ -z "${OPENSHIFT_TOKEN}" ]];
then
    oc login "https://${CLOUD_URL}:8443" -u "${OPENSHIFT_USER}" \
            -p "${OPENSHIFT_PASSWORD}" --insecure-skip-tls-verify

    OPENSHIFT_TOKEN=$(oc whoami -t)
else
    oc login "https://${CLOUD_URL}:8443" --token "${OPENSHIFT_TOKEN}" \
       --insecure-skip-tls-verify
fi

if [[ -z "${OPENSHIFT_WORKSPACE}" ]];
then
    export OPENSHIFT_WORKSPACE="mongodb-main"
fi

read_secret() {
  local path="$1"

  if [ -f "$path" ]; then
    cat "$path"
  else
    printf ""
  fi
}


oc project ${OPENSHIFT_WORKSPACE}

srvversion=$(oc version | tail -n1)
echo ${srvversion}

oc delete dc/mongo-tests || true

export OPENSHIFT_WORKSPACE_WA="${OPENSHIFT_WORKSPACE}"
ATTEMPTS="${WAIT_TIMEOUT:=200}"

oc process REPLICAS="1" \
    TEST_IMAGE=${TEST_IMAGE} \
    TAGS="$TAGS" \
    OPENSHIFT_WORKSPACE_WA="${OPENSHIFT_WORKSPACE}" \
    EXTERNAL_BACKUP_PATH="$EXTERNAL_BACKUP_PATH" \
    DATARS_HOST="$DATARS_HOST" \
    LEFT_NODES_PATTERN="$LEFT_NODES_PATTERN" \
    RIGHT_NODES_PATTERN="$RIGHT_NODES_PATTERN" \
    WAIT_TIMEOUT="$ATTEMPTS" \
    -f ./integration-tests/openshift/template.json |  sed 's#apps.openshift.io/v1#v1#' |oc apply -f -

SLEEP_BETWEEN_ITERATIONS=1

echo "Timeout is $((ATTEMPTS * SLEEP_BETWEEN_ITERATIONS)) seconds"

error_handling() {
    echo "Read tests logs:"

    oc logs dc/mongo-tests

    exit 1
}

_wait() {
  local check_command=$1
  local i
  for ((i=1;i<=ATTEMPTS;i++)); do
    if eval "$check_command"; then
      echo "Done"
      return 0
    fi
    echo -n "."
    # sleep one second
    sleep 1
  done
  echo
  echo >&2 "=> Giving up: $2"
  exit 1
}

_wait 'oc get pods | grep -v "deploy" | grep "mongo-tests-[0-9]" | awk '\''{print $2}'\'' | grep "1/1"' \
    "Could not start tests pod"

for i in $(seq 1 ${ATTEMPTS});
do
  echo "Sleeping the ${i} times..."
  sleep ${SLEEP_BETWEEN_ITERATIONS}

  if [ "${i}" -eq "${ATTEMPTS}" ]
  then
      echo "Time is over"
      echo "could not to get result of this job"
      echo "Job is failed"
      error_handling
  fi

  NUMBER_OF_RESULT_FILES="$(oc rsh dc/mongo-tests /bin/sh -c 'ls ./output | wc -l' || true)"

  echo "The result of command is ${NUMBER_OF_RESULT_FILES}"

  case "${NUMBER_OF_RESULT_FILES}" in
        "3")
            echo "Job is succeeded"
            break
            ;;
        *)
            echo "Tests are running"
            ;;
  esac
done


echo "Read tests logs:"
oc logs dc/mongo-tests


TEST_POD_NAME=$(oc get pods -o=jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}' | grep mongo-tests | grep -v deploy)
oc cp "${TEST_POD_NAME}:/opt/robot/output" /var/jenkins_home/workspace/mongodb-integration-tests/target/

if oc logs dc/mongo-tests | grep '| FAIL |'; then
  echo "Integration tests failed"
  exit 1
fi
