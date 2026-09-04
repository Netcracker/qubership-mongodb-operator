#!/bin/sh
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


read_secret() {
  local path="$1"

  if [ -f "$path" ]; then
    cat "$path"
  fi
}

S3_KEY_ID="$(read_secret /var/run/secrets/mongodb/s3/S3_KEY_ID)"
S3_KEY_SECRET="$(read_secret /var/run/secrets/mongodb/s3/S3_KEY_SECRET)"

exec /opt/backup/backup-daemon \
  --s3-access-key-id "${S3_KEY_ID}"\
  --s3-access-key-secret "${S3_KEY_SECRET}"
  "$@"