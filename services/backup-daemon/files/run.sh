#!/bin/sh

read_secret() {
  local path="$1"

  if [ -f "$path" ]; then
    cat "$path"
  fi
}

S3_KEY_ID="$(read_secret /var/run/secrets/mongodb/s3/username)"
S3_KEY_SECRET="$(read_secret /var/run/secrets/mongodb/s3/password)"

exec /opt/backup/backup-daemon \
  --s3-access-key-id "${S3_KEY_ID}"\
  --s3-access-key-secret "${S3_KEY_SECRET}"
  "$@"