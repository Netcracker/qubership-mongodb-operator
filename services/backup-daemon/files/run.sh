#!/bin/sh

exec /opt/backup/backup-daemon \
  --backup-cmd "/opt/mongodb-backup/scripts.sh backup -f {{.data_folder}} {{.dbs}}" \
  --restore-cmd "/opt/mongodb-backup/scripts.sh restore -f {{.data_folder}} {{.dbs}} {{.dbmap}}" \
  --dblist-cmd "/opt/mongodb-backup/scripts.sh list-dbs -f {{.data_folder}}" \
  --tls-enabled "${INTERNAL_TLS_ENABLED}" \
  --certs-path "${INTERNAL_TLS_PATH}" \
  "$@"