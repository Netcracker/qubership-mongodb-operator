This section provides information about backup procedures using API for MongoDB.

# Prerequisites

* Make sure there is enough free space on backup PV to store uncompressed database files.
* The RAM limit of mongo-backup-daemon pod should be at least 512Mi.
* The mongos pod should be up and running with status "1/1".

For `POST` operations, you must specify the user id and password from the `BACKUP_DAEMON_API_CREDENTIALS_USERNAME` and `BACKUP_DAEMON_API_CREDENTIALS_PASSWORD` environmental parameters so that you can use REST API to run backup tasks.

# Run the Full Manual Backup

This returns backup folder (vault) as a plain-text response. You can later use this vault name to get the backup status (7):

```
curl -XPOST  -u  username:password localhost:8080/backup
```

### Run the Incremental Manual Backup

This request returns backup folder (vault) as a plain-text response. You can later use this vault name to get a backup status:

```
curl -XPOST -u username:password localhost:8080/incremental/backup
```

**Note**: Set the `MONGO_BACKUP_USER` and `MONGO_BACKUP_PASSWORD` parameters to `MONGO_ROOT_USER` and `MONGO_ROOT_PASSWORD` respectively.
Set `ENABLE_FULL_RESTORE` parameter to `true`.

## Run the Manual Backup for Some Subset of DBs (granular backup)

This section provide details about manual backup. 

## Run Manual Backup, Passing DBs

You can also pass a list of collections for every DB backed, and specify query for every collection if needed using the following query: [https://docs.mongodb.com/manual/reference/program/mongodump/#cmdoption-mongodump-query](https://docs.mongodb.com/manual/reference/program/mongodump/#cmdoption-mongodump-query).

This query returns the backup folder (vault) as a plain-text response. You can later use this vault name to get a backup status (7).

For DBs use the following command:

```
curl -XPOST -u username:password -v -H "Content-Type: application/json" -d '{"dbs":["db_name1","db_name2"]}' localhost:8080/backup
```

For DBs with collections use the following command:

```
curl -XPOST -u username:password -v -H "Content-Type: application/json" -d '{"dbs":["db_name1",{"db_name2":{"collections":["first","second"]}]}' localhost:8080/backup
```

For DBs with collections and queries use the following command:

```
curl -XPOST -u username:password -v -H "Content-Type: application/json" -d '{"dbs":["db_name1",{"db_name2":{"collections":["first",{"second":{"test1":"1"}}]}]}' localhost:8080/backup
```

## Run Manual Backup That Will Not Be Deleted Ever

If you do not want your backup to be evicted, add `allow_eviction":"False"` in your request. It works both for full and granular backups: 

```
curl -XPOST -u username:password -v -H "Content-Type: application/json" -d '{"allow_eviction":"False","dbs":["arg1","arg2"]}' localhost:8080/backup
```

### Run Manual Backup that will be stored at NFS

You need to use the `externalBackupPath` parameter.
This returns backup folder (vault) as a plain-text response. You can later use this vault name to get a backup status (7).

```
curl -XPOST -u username:password -v -H "Content-Type: application/json" -d '{"externalBackupPath": "YYYYMMDDHHmmSS/mongo"}' localhost:8080/backup
```

## Run Manual Eviction

For manual eviction, use the following command:

```
curl -XPOST -u username:password localhost:8080/evict
```

## Remove Specific Backup by ID

To remove a specific back up by ID, use the following command:

```
curl -XPOST -u username:password localhost:8080/evict/backupid
```

It returns `200` for `OK`.

## Get Health

The Get Health method returns JSON with the following information:

```
"status": status of backup daemon   
"backup_queue_size": backup daemon queue size (if > 0 then there are 1 or tasks waiting for execution)
 "storage": storage info:
  "total_space": total storage space in bytes
  "dump_count": number of backup
  "free_space": free space left in bytes
  "size": used space in bytes
  "total_inodes": total number of inodes on storage
  "free_inodes": free number of inodes on storage
  "used_inodes": used number of inodes on storage
  "last": last backup metrics
    "metrics['exit_code']": exit code of script 
    "metrics['exception']": python exception if backup failed
    "metrics['spent_time']": spent time
    "metrics['size']": backup size in bytes
    "failed": is failed or not
    "locked": is locked or not
    "id": vault name of backup
    "ts": timestamp of backup  
  "lastSuccessful": last succesfull backup metrics
    "metrics['exit_code']": exit code of script 
    "metrics['spent_time']": spent time
    "metrics['size']": backup size in bytes
    "failed": is failed or not
    "locked": is locked or not
    "id": vault name of backup
    "ts": timestamp of backup\
```

To get JSON health, use the following command:

```
curl -XGET localhost:8080/health
```

For incremental backups, use the following command:

```
curl -XGET localhost:8080/incremental/health
```

## Run Recovery

You must specify JSON with vault. You must specify databases in the DBs list. You cannot run a recovery until you specify the databases.

To run the full recovery, you need to use database-specific methods. You cannot run full recovery from API.

You receive `task_id` as a response. Use it in (7) to get the status of recovery:

```
curl -XPOST -u username:password -v -H "Content-Type: application/json" -d  '{"vault":"20170913T1114", "dbs":["db1","db2"]}' localhost:8080/restore
```

If you need to copy a database, you can use the `changeDbName` arg in JSON.

An example is given below:

```
curl -XPOST -u username:password -v -H "Content-Type: application/json" -d  '{"vault":"20170913T1114", "dbs":["db1","db2","db3""], "changeDbNames":{"db1":"new_db1_name","db2":"new_db2_name"}}' localhost:8080/restore
```

This saves `db1` and `db2` as they are on a DB server, and restore `db1` and `db2` into new (or existing) databases called `new_db1_name` and `new_db2_name`. Database `db3` will be rewritten, because it's not in `changeDbNames` list

The following is an example for restore from incremental backup:

**Note**: Set the `MONGO_BACKUP_USER` and `MONGO_BACKUP_PASSWORD` parameters to `MONGO_ROOT_USER` and `MONGO_ROOT_PASSWORD` respectively.
Set `ENABLE_FULL_RESTORE` parameter to `true`.

```
curl -XPOST -u username:password -v -H "Content-Type: application/json" -d  '{"vault":"20170913T1114"}' localhost:8080/incremental/restore
```

The following is an example of restore from backup that is stored on NFS:

```
curl -XPOST -u username:password -v -H "Content-Type: application/json" -d '{"externalBackupPath": "YYYYMMDDHHmmSS/mongo"}' localhost:8080/restore
```

## Full Restore Procedure

This section provides instruction on how to restore all MongoDB databases from full backup.

**Note**: If any password is changed since full backup, after restore it might lead to inconsistency of the cluster.

### Prerequisites

Ensure that MongoDB cluster is up and running. 

### Procedure

For full restore:

1. Connect to the terminal of MongoDB Backup Daemon.
1. List the available backups using the following command: 
   
   ```curl localhost:8080/listbackups```

1. Select backup to restore and check if it is full back up using the following command:
  
   ```curl localhost:8080/listbackups/<backup_id>```
   
   The output should contain `"is_granular": false`.

1. Run full restore using the following command:
   
   ```/opt/mongodb-backup/scripts.sh restore -f /backup-storage/<backup_id>``` 

1. If a message asking to replace files appears, press "A" to allow replace all files.

If restore is successful, the following message appears: `databases successfully restored`

## Get Backup/Recovery Status

You receive the following HTTP responses: `200` for `OK`, `206` for `Still in process` and `500` for `NOT OK`, use the following command to get recovery status:

```
curl -XGET localhost:8080/jobstatus/<task_id>
```

or 

```
curl -XGET localhost:8080/jobstatus/<vault_name>
```

For incremental backups, use the following command:

```
curl -XGET localhost:8080/incremental/jobstatus/<task_id>
```

or 

```
curl -XGET localhost:8080/incremental/jobstatus/<vault_name>
```

Also, you receive a JSON string as plain-text with the following information:

* `status` - Successful/Queued/Processing/Failed
* `message` - Optional field, only if error, contains description of error
* `vault` - The vault name to use in recovery
* `type` - Backup/restore
* `err` - None if no error, last 5 lines of log if `status=Failed`
* `task_id` - The task_id of the task

An example is given below:  

```
{"status": "Successful", "vault": "20170927T1122", "type": "backup", "err": "None", "task_id": "a592eeb6-abac-4d98-b638-75a820e333b1"}
```

## List Backups

To list the backups, use the following command:

```
curl -XGET localhost:8080/listbackups
```

For incremental backups, use the following command:

```
curl -XGET localhost:8080/incremental/listbackups
```

This command returns a JSON list of backup names.

## Get Backup Information

To get the backup information, use the following command:

```
curl -XGET localhost:8080/listbackups/<vault_id>
```

For incremental backups:

```
curl -XGET localhost:8080/incremental/listbackups/<vault_id>
```

This command returns a JSON string with status about particular backup:

* `ts` - UNIX timestamp of backup
* `spent_time` - Time spent on backup (in ms)
* `db_list` - List of backed up databases
* `id` - The vault name
* `size` - The size of backup in bytes
* `evictable` - Whether backup is evictable
* `locked` - Whether backup is locked (either process isn't finished, or it failed somehow)
* `exit_code` - The exit code of backup script
* `failed` - Whether backup failed or not
* `valid` - Backup is valid or not

An example is given below:

```
{"ts": 1514282821000, "spent_time": "5066ms", "db_list": "Sorry, no information on databases available", "id": "20171226T100701", "size": "36647b", "evictable": true, "locked": false, "exit_code": 0, "failed": false, "valid": true}
```

### Upload/Download API

To get backup archive use the command below.

For full backups, use the following command:

```
curl -XGET localhost:8080/backup/<backup_id>
```

For incremental backups, use the following command:

```
curl -XGET localhost:8080/incremental/backup/<backup_id>
```

This command returns a status `200 OK` and archive with current backup that was specified in backup_id.

To upload backup archive, use the following command:

```
curl -XPOST localhost:8080/restore/backup
```

Then attach backup archive with name that would be ID of this backup.
If the backup was successfully loaded, then it returns the `200` status.
