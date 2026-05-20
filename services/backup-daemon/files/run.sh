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
  --restore-cmd "/opt/mongodb-backup/scripts.sh restore -f {{.data_folder}} {{.dbs}} {{.dbmap}}" \
  --dblist-cmd "/opt/mongodb-backup/scripts.sh list-dbs -f {{.data_folder}}" \
  --tls-enabled "${INTERNAL_TLS_ENABLED}" \
  --certs-path "${INTERNAL_TLS_PATH}" \
  --s3-access-key-id "${S3_KEY_ID}"\
  --s3-access-key-secret "${S3_KEY_SECRET}"
  "$@"