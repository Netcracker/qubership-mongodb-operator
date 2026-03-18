#!/bin/bash

export ROBOT_OPTIONS="--loglevel=debug --outputdir output"

if [[ "$DEBUG" == true ]]; then
  set -x
  printenv
fi

run_ttyd() {
  if [[ -z "$TTYD_PORT" ]]; then
    TTYD_PORT=8080
  fi

  exec ttyd -p ${TTYD_PORT} bash
}

# Process some known arguments to run integration tests
case $1 in
  run-robot)
    if [[ -z "$TAGS" ]]; then
      robot ./tests
    else
      robot -i ${TAGS} ./tests
    fi
    # python3 analyze_result.py
    # run_ttyd
    ;;
  run-ttyd)
    run_ttyd
    ;;
esac

echo "sleeping 1 min"
sleep 60
