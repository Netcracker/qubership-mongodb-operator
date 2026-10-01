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

set_default() {
    # sets default for params
    var=$1
    value=$2

    if [[ $OVERRIDE_ENV == "true" ]] || [[ -z ${!var+x} ]] || [[ -z ${!var} ]] ; then
    	if [[ ${value} == "--empty" ]]; then
    		export "${var}"=
    	else
        	export "${var}"="${value}"
        fi
    fi
}

export OPENSHIFT_WORKSPACE_WA="${OPENSHIFT_WORKSPACE}"

if [[ -n $RUN_INTEGRATION_TESTS ]] && [[ $RUN_INTEGRATION_TESTS == "true" ]]; then
    oc policy add-role-to-user admin system:serviceaccount:$OPENSHIFT_WORKSPACE:default
	export REPLICAS=1

    OVERRIDE_ENV=false
    # If set to True ignores API Credentials and sets backup to work unauth
	# Else sets default credentials in case of it is not overridden
	set_default BACKUP_DAEMON_API_UNAUTH "False"
    if [[ "${BACKUP_DAEMON_API_UNAUTH,,}" == "true" ]]; then 
		export BACKUP_DAEMON_API_CREDENTIALS_PASSWORD=""
		export BACKUP_DAEMON_API_CREDENTIALS_USERNAME=""
	else
		set_default BACKUP_DAEMON_API_CREDENTIALS_PASSWORD "backup"
		set_default BACKUP_DAEMON_API_CREDENTIALS_USERNAME "backup"
	fi
fi
