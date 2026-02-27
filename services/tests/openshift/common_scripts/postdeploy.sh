#!/bin/bash
set -x 

delete_dc(){
  oc delete dc/mongo-tests || true
}

if [[ -z $RUN_INTEGRATION_TESTS ]] || [[ $RUN_INTEGRATION_TESTS != "true" ]]; then
  delete_dc
  exit 0
fi

ATTEMPTS="${WAIT_TIMEOUT:=200}"
SLEEP_BETWEEN_ITERATIONS=1

echo "Timeout is $((ATTEMPTS * SLEEP_BETWEEN_ITERATIONS)) seconds"

for i in $(seq 1 ${ATTEMPTS});
do
  echo "Sleeping the ${i} times..."
  sleep ${SLEEP_BETWEEN_ITERATIONS}

  if [ "${i}" -eq "${ATTEMPTS}" ]
  then
      echo "Time is over"
      oc logs dc/mongo-tests
      delete_dc
      exit 1
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

if oc logs dc/mongo-tests | grep '| FAIL |'; then
  echo "Integration tests failed"
  delete_dc
  exit 1
fi

delete_dc