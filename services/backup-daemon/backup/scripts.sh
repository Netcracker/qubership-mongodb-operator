#!/bin/bash
set -x
set -e
set -o pipefail
set -u
OPTIND=1 # Reset in case getopts has been used previously in the shell.

require() {
    local varName=$1
    if [[ -z "${!varName}" ]]; then
        echo >&2 "$1 is not defined"
        exit 1
    fi
}

containsElement() {
    local e match="$1"
    shift
    for e; do
        [[ "$e" == "$match" ]] && return 0
    done
    return 1
}

read_secret() {
  local path="$1"

  if [ -f "$path" ]; then
    cat "$path"
  fi
}

MONGO_BACKUP_USER="$(read_secret /var/run/secrets/mongodb/mongo-backup/username)"
MONGO_BACKUP_PASSWORD="$(read_secret /var/run/secrets/mongodb/mongo-backup/password)"
MONGO_RESTORE_USER="$(read_secret /var/run/secrets/mongodb/mongo-restore/username)"
MONGO_RESTORE_PASSWORD="$(read_secret /var/run/secrets/mongodb/mongo-restore/password)"
BACKUP_DAEMON_API_CREDENTIALS_USERNAME="$(read_secret /var/run/secrets/mongodb/backup-api/username)"
BACKUP_DAEMON_API_CREDENTIALS_PASSWORD="$(read_secret /var/run/secrets/mongodb/backup-api/password)"

MONGO_DUMP="mongodump --gzip"
MONGORESTORE="mongorestore"
MONGO_CMD="mongo"
MONGOIMPORT="mongoimport"
MONGOEXPORT="mongoexport"

TLS_ENABLED=${TLS_ENABLED:-false}
DEBUG=${DEBUG:-false}

if [[ "${TLS_ENABLED}" = true ]]; then
    MONGO_DUMP="${MONGO_DUMP} --ssl --sslCAFile=${TLS_ROOTCERT}"
    MONGORESTORE="${MONGORESTORE} --ssl --sslCAFile=${TLS_ROOTCERT}"
    MONGO_CMD="${MONGO_CMD} --tls --tlsCAFile=${TLS_ROOTCERT}"
    MONGOIMPORT="${MONGOIMPORT} --ssl --sslCAFile=${TLS_ROOTCERT}"
    MONGOEXPORT="${MONGOEXPORT} --ssl --sslCAFile=${TLS_ROOTCERT}"

    # remove before using production certs
    MONGO_DUMP="${MONGO_DUMP} --sslAllowInvalidCertificates"
    MONGORESTORE="${MONGORESTORE} --sslAllowInvalidCertificates"
    MONGO_CMD="${MONGO_CMD} --tlsInsecure"
    MONGOIMPORT="${MONGOIMPORT} --sslAllowInvalidCertificates"
    MONGOEXPORT="${MONGOEXPORT} --sslAllowInvalidCertificates"
fi

NUM_PARALLEL_CONNECTIONS=${NUM_PARALLEL_CONNECTIONS:-4}
GRANULAR_NUM_PARALLEL_CONNECTIONS=${GRANULAR_NUM_PARALLEL_CONNECTIONS:-4}

MONGO_MARKER_DB=${MONGO_MARKER_DB:-admin}
MONGO_MARKER_COLLECTION=${MONGO_MARKER_COLLECTION:-cloudBackupMarkers}

cluster_backup() {
    require MONGO_AUTH_DB
    export FAILED=0
    vaultfolder=${vault##*/}

    [[ -z "${MONGO_BACKUP_DB}" ]] && mongoHost=mongos || mongoHost="${MONGO_BACKUP_DB}"

    # ===== clean up databases JSON =====
    if [[ -n ${databases:+x} ]]; then
        clean_databases=$(echo "$databases" | sed "s/^'//;s/'$//" | sed 's/\\"/"/g')
    fi
    # ==================================================

    if [[ -z ${databases:+x} ]]; then
        echo "=> Start backup mongo cluster to: $vault..."
        ${MONGO_DUMP} \
            --username="${MONGO_BACKUP_USER}" \
            --password="${MONGO_BACKUP_PASSWORD}" \
            --authenticationDatabase="${MONGO_AUTH_DB}" \
            --host="${mongoHost}" \
            --numParallelCollections="${NUM_PARALLEL_CONNECTIONS}" \
            --gzip \
            --out="$vault" || export FAILED=1
    else
        # if also DB_NAMES are specified
        echo "=> Start partial backup mongo cluster to: $vault ..."

        local number_of_dbs
        number_of_dbs=$(echo "$clean_databases" | jq -r 'length')
        if [[ $? != 0 ]]; then
            echo "!! Couldn't parse DBs JSON"
            exit 1
        fi

        for ((DB_NUMBER = 0; DB_NUMBER < number_of_dbs; DB_NUMBER++)); do
            if [[ $(echo "$clean_databases" | jq -r ".[$DB_NUMBER] | type") == "string" ]]; then
                db=$(echo "$clean_databases" | jq -cr ".[$DB_NUMBER]")
                collections=""
            elif [[ $(echo "$clean_databases" | jq -r ".[$DB_NUMBER] | type") == "object" ]]; then
                db=$(echo "$clean_databases" | jq -cr ".[$DB_NUMBER] | keys[0]")
                collections=$(echo "$clean_databases" | jq -cr ".[$DB_NUMBER][\"$db\"].collections")
            else
                echo "!! Couldn't parse DBs JSON"
                exit 1
            fi

            local number_of_cols
            number_of_cols=$(echo "$collections" | jq -r length)

            # if 'collection' field is empty, backing up the whole db
            if [[ -z ${number_of_cols} ]]; then
                result=$(${MONGO_DUMP} -v \
                    --username="${MONGO_BACKUP_USER}" \
                    --password="${MONGO_BACKUP_PASSWORD}" \
                    --authenticationDatabase="${MONGO_AUTH_DB}" \
                    --dumpDbUsersAndRoles \
                    --host="$mongoHost" \
                    --numParallelCollections="${GRANULAR_NUM_PARALLEL_CONNECTIONS}" \
                    --gzip \
                    --out="$vault" \
                    --db="$db" 2>&1) || FAILED=1
                if [[ ${FAILED} -gt 0 || $(echo "$result" | head -n1 | grep "dumping up to 0 collections in parallel") != "" ]]; then
                    echo "!! Database $db failed to backup"
                    FAILED=1
                fi
            else
                # backing up only set of collections
                for ((COL_NUMBER = 0; COL_NUMBER < number_of_cols; COL_NUMBER++)); do
                    if [[ $(echo "$collections" | jq -r ".[$COL_NUMBER]| type") == "string" ]]; then
                        col=$(echo "$collections" | jq -r ".[$COL_NUMBER]")
                        query=""
                    elif [[ $(echo "$collections" | jq -r ".[$COL_NUMBER]| type") == "object" ]]; then
                        col=$(echo "$collections" | jq -r ".[$COL_NUMBER] | keys[0]")
                        query=$(echo "$collections" | jq -cr ".[$COL_NUMBER][\"$col\"]")
                    else
                        echo "!! Couldn't parse Collections JSON for DB $db"
                        exit 1
                    fi

                    echo "=> Backing up: db: $db, collection: $col, query: $query"

                    [[ -n ${query} ]] && query="-q $query"

                    result=$(${MONGO_DUMP} -v \
                        --username="${MONGO_BACKUP_USER}" \
                        --password="${MONGO_BACKUP_PASSWORD}" \
                        --authenticationDatabase="${MONGO_AUTH_DB}" \
                        --host="$mongoHost" \
                        --numParallelCollections="${GRANULAR_NUM_PARALLEL_CONNECTIONS}" \
                        --gzip \
                        --out="$vault" \
                        --db="$db" \
                        -c "$col" \
                        ${query} 2>&1) || FAILED=1

                    if [[ ${FAILED} -gt 0 || $(echo "$result" | head -n1 | grep "dumping up to 0 collections in parallel") != "" ]]; then
                        echo "!! Database $db failed to backup"
                        FAILED=1
                    fi
                done
            fi
        done
    fi

    if [[ ${FAILED} -eq 1 ]]; then
        echo "=> Backup failed. Look at $vault/.console for logs"
        exit 1
    else
        echo "=> Backup successfully dumped (gzip enabled, no manual compression needed)"
    fi

    echo "=> Backup process successfully finished"
}

restore_user_databases() {
    require MONGO_AUTH_DB
    export FAILED=0

    if [[ ! -d "$vault" ]]; then
        echo >&2 "Vault folder does not exist by given path: $vault"
        exit 1
    fi

    if ls "$vault"/*.zip 1>/dev/null 2>&1; then
        archiveFormat="zip"
    elif ls "$vault"/*.tgz 1>/dev/null 2>&1; then
        archiveFormat="tgz"
    else
        archiveFormat="gzip-folder"
    fi

    use_gzip=false
    if [[ "$archiveFormat" == "gzip-folder" ]]; then
        while read -r dir; do
            if find "$dir" -maxdepth 1 -type f -name "*.bson.gz" | grep -q .; then
                use_gzip=true
                break
            fi
        done < <(find "$vault" -type d ! -name admin ! -name config)
    fi
    echo "use_gzip is set to: $use_gzip"

    echo -n "=> Test dump archive..."
    if [[ "$archiveFormat" == "tgz" ]]; then
        gzip -t "$vault"/*.tgz && echo "OK"
    elif [[ "$archiveFormat" == "zip" ]]; then
        unzip -t "$vault"/*.zip && echo "OK"
    else
        echo "OK"
    fi

    [[ -z "${MONGO_SOURCE_DB}" ]] && mongoHost=mongos || mongoHost="${MONGO_SOURCE_DB}"
    echo "=> Mongo host to restore - $mongoHost"

    # Decompress archive if needed
    if [[ "$archiveFormat" == "tgz" ]]; then
        echo -n "=> Decompressing archive..."
        tar -xzf "$vault"/*.tgz --directory=/ && echo "OK"
    elif [[ "$archiveFormat" == "zip" ]]; then
        echo -n "=> Decompressing archive..."
        unzip -o "$vault"/*.zip -d / && echo "OK"
    fi

    # ===== clean up databases JSON =====
    if [[ -n ${databases:+x} ]]; then
        clean_databases=$(echo "$databases" | sed "s/^'//;s/'$//" | sed 's/\\"/"/g')
    fi
    # ==================================================

    # Pre-check that all requested databases exist in backup
    if [[ -n ${databases:+x} ]]; then
        backed_dbs_array="$(list_databases)"
        local number_of_dbs
        number_of_dbs=$(echo "$clean_databases" | jq -r 'length')
        for ((DB_NUMBER = 0; DB_NUMBER < number_of_dbs; DB_NUMBER++)); do
            db=$(echo "$clean_databases" | jq -cr ".[$DB_NUMBER]")
            if ! containsElement "$db" ${backed_dbs_array}; then
                echo "❌ Database $db not found in backup. Aborting restore."
                exit 1
            fi
        done
    fi

    if [[ -z ${databases:+x} ]]; then
        echo "=> Restore all databases:"
        echo "=> No specific databases passed; restoring all available databases from backup"
        for database in $(ls -d $vault/* | grep -v -e '\.tgz$' | grep -v -e '\.zip$'); do
            local db_name=$(basename "$database")
            [[ "$db_name" == "config" || "$db_name" == "admin" ]] && continue
            echo "=>        $database"
            ${MONGORESTORE} --drop \
                --username="${MONGO_RESTORE_USER}" \
                --password="${MONGO_RESTORE_PASSWORD}" \
                --authenticationDatabase="${MONGO_AUTH_DB}" \
                --host="$mongoHost" \
                --numParallelCollections="${NUM_PARALLEL_CONNECTIONS}" \
                --db="$db_name" \
                $( [[ "$use_gzip" == true ]] && echo "--gzip" ) \
                "$database" || FAILED=1
        done
    else
        echo "=> Restoring selected databases..."
        local number_of_dbs
        number_of_dbs=$(echo "$clean_databases" | jq -r 'length')
        for ((DB_NUMBER = 0; DB_NUMBER < number_of_dbs; DB_NUMBER++)); do
            database=$(echo "$clean_databases" | jq -cr ".[$DB_NUMBER]")
            db_name=$(basename "$database")
            [[ "$db_name" == "config" || "$db_name" == "admin" ]] && continue
            echo "=> Restoring database: $database"
            if [[ ! -z ${dbmap:+x} ]]; then
                new_db_name=$(echo "${dbmap}" | jq ".[\"$db_name\"]" -r)
                if [[ $new_db_name != "null" ]]; then
                    db_name=$new_db_name
                fi
            fi
            ${MONGORESTORE} --drop \
                --username="${MONGO_RESTORE_USER}" \
                --password="${MONGO_RESTORE_PASSWORD}" \
                --authenticationDatabase="${MONGO_AUTH_DB}" \
                --host="$mongoHost" \
                --numParallelCollections="${GRANULAR_NUM_PARALLEL_CONNECTIONS}" \
                --db="$db_name" \
                $( [[ "$use_gzip" == true ]] && echo "--gzip" ) \
                "$vault/$database" || FAILED=1
        done
    fi

    # Cleanup
    if [[ "$archiveFormat" != "gzip-folder" ]]; then
        for i in $(ls "$vault" | grep -v -e '\.tgz$' | grep -v -e '\.zip$'); do
            rm -rf "$vault/$i"
        done
    fi

    rm -rf "$vault/.lock"

    if [[ $FAILED -eq 1 ]]; then
        echo "❌ One or more databases failed to restore."
        exit 1
    fi
    echo "✅ Databases successfully restored"
}


restore_admin_database() {
    require CONFIG_NODES

    if [[ ! -d "$vault" ]]; then
        echo >&2 "Vault folder does not exist by given path: $vault"
        exit 1
    fi

    if ls "$vault"/*.zip 1>/dev/null 2>&1; then
        archiveFormat="zip"
    elif ls "$vault"/*.tgz 1>/dev/null 2>&1; then
        archiveFormat="tgz"
    else
        archiveFormat="gzip-folder"
    fi

    is_gzip_backup=false
    if [[ "$archiveFormat" == "gzip-folder" ]]; then
        if find "$vault/admin" -type f -name "*.bson.gz" | grep -q .; then
            is_gzip_backup=true
        fi
    fi

    # Validate archive if applicable
    if [[ "$archiveFormat" == "tgz" ]]; then
        echo -n "=> Test .tgz archive..."
        gzip -t "$vault"/*.tgz && echo "OK"
    elif [[ "$archiveFormat" == "zip" ]]; then
        echo -n "=> Test .zip archive..."
        unzip -t "$vault"/*.zip && echo "OK"
    fi

    # Decompress if archive
    if [[ "$archiveFormat" == "tgz" ]]; then
        echo -n "=> Decompressing .tgz archive..."
        tar -xzf "$vault"/*.tgz --directory=/ && echo "OK"
    elif [[ "$archiveFormat" == "zip" ]]; then
        echo "=> Decompressing .zip archive..."
        unzip -o "$vault"/*.zip -d /
    fi

    echo "=> Restore admin database..."
    restore_cmd="$MONGORESTORE --drop \
        --host=\"$CONFIG_NODES\" \
        --db=admin \
        --batchSize=1"

    $restore_cmd $( [[ "$is_gzip_backup" == true ]] && echo "--gzip" ) "$vault/admin" || {
        echo "!! Failed to restore admin database."
        exit 1
    }

    for i in $(ls "$vault" | grep -v -e '\.tgz$' | grep -v -e '\.zip$'); do
        rm -rf "$vault/$i"
    done

    echo "=> Admin database successfully restored"
}

set_marker() {
    require MONGO_AUTH_DB

    [[ -z "${MONGO_BACKUP_DB}" ]] && mongoHost=mongos || mongoHost="${MONGO_BACKUP_DB}"

    if [[ -z "${marker_data:+x}" ]]; then
        echo >&2 "Marker data is required (-d flag)"
        exit 1
    fi

    clean_marker=$(echo "$marker_data" | sed "s/^'//;s/'$//" | sed 's/\\"/"/g')
    singleton_marker=$(echo "${clean_marker}" | jq '. + {"_singleton": 1}')

    echo "=> Writing marker to ${MONGO_MARKER_DB}.${MONGO_MARKER_COLLECTION} on ${mongoHost}..."
    echo "${singleton_marker}" | ${MONGOIMPORT} \
        --username="${MONGO_BACKUP_USER}" \
        --password="${MONGO_BACKUP_PASSWORD}" \
        --authenticationDatabase="${MONGO_AUTH_DB}" \
        --host="${mongoHost}" \
        --db="${MONGO_MARKER_DB}" \
        --collection="${MONGO_MARKER_COLLECTION}" \
        --mode=upsert \
        --upsertFields=_singleton \
        --type=json || {
        echo >&2 "!! Failed to write marker"
        exit 1
    }

    echo "=> Marker written successfully"
}

get_marker() {
    require MONGO_AUTH_DB

    [[ -z "${MONGO_BACKUP_DB}" ]] && mongoHost=mongos || mongoHost="${MONGO_BACKUP_DB}"

    result=$(${MONGOEXPORT} \
        --username="${MONGO_BACKUP_USER}" \
        --password="${MONGO_BACKUP_PASSWORD}" \
        --authenticationDatabase="${MONGO_AUTH_DB}" \
        --host="${mongoHost}" \
        --db="${MONGO_MARKER_DB}" \
        --collection="${MONGO_MARKER_COLLECTION}" \
        --type=json 2>/dev/null)

    if [[ -z "${result}" ]]; then
        echo >&2 "!! No marker found"
        exit 1
    fi

    marker_value=$(echo "${result}" | jq -r '.marker')
    if [[ -z "${marker_value}" || "${marker_value}" == "null" ]]; then
        echo >&2 "!! Marker entry found but 'marker' field is missing"
        exit 1
    fi

    echo "${marker_value}"
}

show_help() {
    echo "Backup daemon script used for backing up mongo dbs"
    echo "Usage:"
    echo "scripts.sh action [-h] -f vault [-d databases] [-m dbmap]"
    echo "			action: backup/restore/restore-admin/set-marker/get-marker"
    echo "					backup - backups all dbs or specified number of dbs (via -d key)"
    echo "					restore - restores all dbs or specified number of dbs (via -d key). Can rename some dbs using -m key"
    echo "					restore-admin - restores admin database with users"
    echo "					list-dbs - shows backed up databases from vault"
    echo "					set-marker - writes/overwrites the single marker document to MongoDB native storage (via -d JSON)"
    echo "					get-marker - fetches the single marker document from MongoDB native storage"
    echo "			-h: print this help"
    echo "			-f: specify full path to vault to backup into\restore from"
    echo "			-d: JSON of databases (or JSON marker document for set-marker)"
    echo "			-m: JSON of databases rename like {\"old_db_name1\":\"new_db_name1\",\"old_db_name2\":\"new_db_name2\"}"
}

list_databases() {
    trimmedvault=$(echo "$vault" | sed 's~^/~~;s~/$~~')

    if ls "$vault"/*.zip 1>/dev/null 2>&1; then
        archiveFormat="zip"
    elif ls "$vault"/*.tgz 1>/dev/null 2>&1; then
        archiveFormat="tgz"
    else
        archiveFormat="gzip-folder"
    fi

    if [[ "$archiveFormat" == "tgz" ]]; then
        tar -tvf "${vault}"/*.tgz | grep -e '^d' | awk '{print $6}' |
            sed "s~^$trimmedvault/~~" | awk -F/ '{print $1}' | sort -u
    elif [[ "$archiveFormat" == "zip" ]]; then
        unzip -lqq "${vault}"/*.zip |
            awk '{ if($4 !~ /\/\..*/ && $4 ~ /[^\/]$/) {split($4,a,"/"); len=length(a); if(b[a[len - 1]]++ == 0) print a[len - 1]}}'
    else
        # gzip-folder: new format from mongodump --gzip
        find "$vault" -mindepth 1 -maxdepth 1 -type d -exec basename {} \; | grep -v -e '^admin$' -e '^config$'
    fi
}


# execute command

# get action
: ${1?"Missed argument: operation: backup/restore-admin/restore/set-marker/get-marker"}

action=$1
shift

for arg in "$@"; do
    shift
    case "$arg" in
    "-start_ts") set - "$@" "-s" ;;
    *) set - "$@" "$arg" ;;
    esac
done

#set -x

while getopts "h?f:d:m:s:" option; do
    case "${option}" in
    h | \?)
        show_help
        exit 1
        ;;
    f)
        export vault="$OPTARG"
        ;;
    d)
        export databases="$OPTARG"
        export marker_data="$OPTARG"
        ;;
    m)
        export dbmap="$OPTARG"
        ;;
    s)
        export startTimestamp="$OPTARG"
        ;;
    esac
done

case "${action}" in
"set-marker" | "get-marker") ;;
*) : ${vault?"Missed argument: backup folder: set it using -f flag"} ;;
esac

case "${action}" in
"backup") cluster_backup ;;
"inc-backup") cluster_incremental_backup ;;
"restore-admin") restore_admin_database ;;
"restore") restore_user_databases ;;
"inc-restore") incremental_restore ;;
"list-dbs") list_databases ;;
"set-marker") set_marker ;;
"get-marker") get_marker ;;
"-h") show_help ;;
*) show_help >&2 && exit 1 ;;
esac
