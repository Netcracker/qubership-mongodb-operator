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
