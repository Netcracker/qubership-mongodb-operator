This section describes the various commands to work with MongoDB.

# Table of Contents

- [Table of Contents](#table-of-contents)
- [MongoDB](#mongodb)
  - [Login to MongoDB](#login-to-mongodb)
  - [Change or Create Database](#change-or-create-database)
  - [Show Collections](#show-collections)
  - [Get Databases Size](#get-databases-size)
  - [Get Collections Size](#get-collections-size)
  - [Get Current Active Operations](#get-current-active-operations)
  - [Create User](#create-user)
  - [Create Root User](#create-root-user)
  - [Block User](#blocking-a-compromised-or-unused-account-in-mongodb)
  - [Get ReplicaSet Status](#get-replicaset-status)
    - [Get ReplicaSet Config](#get-replicaset-config)
  - [Change ReplicaSet Config](#change-replicaset-config)
  - [Get Replication Lag of a Node](#get-replication-lag-of-a-node)
  - [Get Replication Lag of Other Nodes](#get-replication-lag-of-other-nodes)
  - [Add a Node to RS](#add-a-node-to-rs)
  - [Remove a Node from RS](#remove-a-node-from-rs)
  - [Get Overall Sharded Cluster Status](#get-overall-sharded-cluster-status)
  - [Get Database Distribution over Replica Sets](#get-database-distribution-over-replica-sets)
- [Change Password in MongoDB](#change-password-in-mongodb)
  - [Password Requirements](#password-requirements)
  - [Change Password in Mongos](#change-password-in-mongos)
  - [Change Root Password in Shards](#change-root-password-in-shards)
  - [Change Backup Password in Services](#change-backup-password-in-services)
  - [Change Restore Password in Services](#change-restore-password-in-services)
  - [Change Monitoring Password in Services](#change-monitoring-password-in-services)
  - [Change DBaaS Password in Services](#change-dbaas-password-in-services)
- [Change Password of DBaaS Adapter REST API](#change-password-of-dbaas-adapter-rest-api)
- [Change Password of DBaaS Aggregator REST API](#change-password-of-dbaas-aggregator-rest-api)
- [Change Password of InfluxDB](#change-password-of-influxdb)
- [Remove MongoDB](#remove-mongodb)
  - [Remove MongoDB from DBAAS Aggregator](#remove-mongodb-from-dbaas-aggregator)
  - [Uninstall Helm Release](#uninstall-helm-release)
  - [Cleanup Namespace:](#cleanup-namespace)
- [Vault Integration](#vault-integration)
  - [Credential Rotation](#credential-rotation)
- [If Operator Needs to be Restarted](#if-operator-needs-to-be-restarted)


# MongoDB

## Login to MongoDB

* To log in to a MongoDB shell, execute the following command in the mongos-XXX pod:

   ``` 
  mongo admin -u <MONGO_ROOT_USER> -p <MONGO_ROOT_PASS> 
   ```   

## Change or Create Database

The command creates a database, if it does not exist already, and uses the database name.
	
To change or create a database, use the following command:
   
  ```
use <database_name>
  ```
    
## Show Collections
    
To show collections, use the following command:

  ```
db.getCollectionNames()
  ```
   
## Get Databases Size

To get the total DBs size in bytes, execute the following command in the mongos-XXX pod:

```
db.adminCommand("listDatabases").databases.map( function(db) { return db.sizeOnDisk;} ).reduce(function(sum, value) {  return sum + value;}, 0);
```

To get the size of all databases in bytes, execute the following command in the mongos-XXX pod:

  ```
db.adminCommand("listDatabases").databases.sort(
    function(l, r) {return r.sizeOnDisk - l.sizeOnDisk}).forEach(
        function(d) {print(d.name + " - " + d.sizeOnDisk)});
  ```

## Get Collections Size

Collection size is represented by the storageSize value. For more information, refer to [https://www.mongodb.com/docs/manual/reference/command/collStats/#mongodb-data-collStats.storageSize](https://www.mongodb.com/docs/manual/reference/command/collStats/#mongodb-data-collStats.storageSize).

To get the size of all collections in all databases in bytes, execute the following command in the mongos-XXX pod:

```
db.adminCommand({ listDatabases: 1, nameOnly: true} )["databases"].forEach( 
    function(database){ 
        var name = database["name"];  
        print(name);  
        db.getSiblingDB(name).getCollectionNames().forEach( 
            function(collname) { print("\t"+collname); print("\t\t " + db.getSiblingDB(name).getCollection(collname).storageSize());  }); 
    });
```

To get the size of all collections in a specific database in bytes, execute the following command in the mongos-XXX pod:

```
dbName="admin";
db.getSiblingDB(dbName).getCollectionNames().forEach( 
            function(collname) { print("\t"+collname); print("\t\t " + db.getSiblingDB(dbName).getCollection(collname).storageSize());  }); 
```

Where, dbName is the name of the database.

## Get Current Active Operations

To get the current active operations, log in to mongo in the mongos-XXX pod and execute the following command:

```
db.currentOp(
   {
     "active" : true,
     "secs_running" : { "$gt" : 1 },
     "ns" : /^test_db\./
   }
)
```
where, dbName is the name of the database


Where:
* `/^test_db\./` is a regex to filter databases to list the active operations.
* `{ "$gt" : 1 }` is a filter to list the active operations that are taking longer than 1 second.

## Create User

To create a new user, use the following command:

```
use admin;
db.createUser({user: '${MONGO_USER}', pwd: '${MONGO_PASSWORD}', roles: ['${ROLE1}','${ROLE2}',...]});
```

## Create Root User

To create a root user, use the following command:

```
use admin;
db.createUser({user: 'root', pwd: '${MONGO_PASSWORD}', roles: ['root']});
```

## Blocking a Compromised or Unused Account in MongoDB

To identify the user that needs to be blocked, you can list all MongoDB users in the admin database:
```
use admin
db.system.users.find()
```

MongoDB does not directly support "disabling" an account in the traditional sense. However, you can block an account by either:

• Removing User Roles: You can revoke all roles from the user, making them unable to perform any operations.(Temporary Block)
```
db.revokeRolesFromUser("<username>", [ /* list of all roles */ ])
```

• Dropping the User: If you want to permanently block the user, you can delete the user account.(Permanent Block)
```
db.dropUser("<username>")
```

## Get ReplicaSet Status 

You can run only on datars\cnfrs pods.

To get the ReplicaSet status, use the following command:

```
rs.status()
```

### Get ReplicaSet Config 

You can run only on datars\cnfrs pods.

To get ReplicaSet config, use the following command:
 
```
rs.conf()
```

## Change ReplicaSet Config 

You can run only on datars\cnfrs pods.

For more information on RS configuration, refer to the following links:    

* [https://docs.mongodb.com/manual/reference/replica-configuration/](https://docs.mongodb.com/manual/reference/replica-configuration/)
* [https://docs.mongodb.com/manual/reference/method/rs.reconfig/](https://docs.mongodb.com/manual/reference/method/rs.reconfig/)

To change the config, use the following command:

```
# get config
cfg = rs.conf();
# change any field of config, for example change priority of first member:
cfg.members[1].priority = 2;
# reconfigure
rs.reconfig(cfg);
```

## Get Replication Lag of a Node 

You can run only on datars\cnfrs pods.

For more information, refer to [https://docs.mongodb.com/manual/reference/method/rs.printReplicationInfo/](https://docs.mongodb.com/manual/reference/method/rs.printReplicationInfo/).

To get the replication lag of a node, use the following command:

```
rs.printReplicationInfo()
```

## Get Replication Lag of Other Nodes

You can run only on datars\cnfrs pods.

For more information, refer to [https://docs.mongodb.com/manual/reference/method/rs.printSlaveReplicationInfo/](https://docs.mongodb.com/manual/reference/method/rs.printSlaveReplicationInfo/).

To get the replication lag of other nodes, use the following command:

```
rs.printSlaveReplicationInfo()
```

## Add a Node to RS

You can run only on datars\cnfrs pods.

To add a node to RS, use the following command:

```
rs.add("url_of_node:port")
```

## Remove a Node from RS

You can run only on datars\cnfrs pods.

To remove a node from RS, use the following command:

```
rs.remove("url_of_node:port")
```

## Get Overall Sharded Cluster Status

To get the status of a sharded cluster, use the following command:

```
sh.status()
```

## Get Database Distribution over Replica Sets

To get specific database distribution over replica sets, use the following command:

```
    use <db_name>
	db.stats()
```

The following output example shows that 30 collections of database mano_ne1 are located on datars2:

```
...
		"datars2/datars20-0.datars2.mongo-cluster.svc.cluster.local:27017,datars21-0.datars2.mongo-cluster.svc.cluster.local:27017,datars22-0.datars2                                                                                                                                                                                                                                                                                                                                                                                  
.mongo-cluster.svc.cluster.local:27017" : {                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    
                        "db" : "mano_ne1",                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     
                        "collections" : 30,
...
```

To get all databases distribution over replica sets, use the following command in the terminal of the primary cnfrs:

```
use config 
db.databases.find()    
```

The following output example shows that the test and `system-monitor_dumps` databases are located on datars1:

```
{ "_id" : "test", "primary" : "datars1", "partitioned" : false }                                                                                             
{ "_id" : "system-monitor_dumps", "primary" : "datars1", "partitioned" : false } 
```

# Change Password in MongoDB

This section provides information on how to change the password for the users in MongoDB.

In Mongo deployment, there are following default user/password pairs:

* Root user: root/root
* Backup user: backup/backup
* Restore user: root/root
* Monitoring user: monitoring/monitoring
* DBaaS adapter user: dbaas/dbaas

## Password Requirements

Passwords must be at least 8 characters long and contain at least one letter, one digit, and one special character.

## Change Password in Mongos

To change password in Mongos:

1. Open the OpenShift or Kubernetes UI Console.
1. Select the MongoDB project for OpenShift and the MongoDB namespace for Kubernetes.
1. Navigate to the terminal of the mongos pod.
1. Execute the following command:
   
   ```
   mongosh admin -u <current_root_user> -p <current_root_password>
   ```
   
   The default user is root. The default password is root. If you have changed the password, use the current password.
1. To update the password, execute the following command:
   
   ```
   db.changeUserPassword('<current_user>','<new_password>')
   ```
   
1. Close the terminal.

## Change Root Password in Shards

If you are using a Non-HA scheme of deployment, you can skip this step.

1. Open the OpenShift or Kubernetes UI Console.
1. Select the MongoDB project for OpenShift and the MongoDB namespace for Kubernetes.
1. Navigate to the terminal of any `datars1X-0` pod.
1. Execute the following command:
   
   ```
   mongosh admin -u <current_root_user> -p <current_root_password>
   ```
  
   The default user is root. The default password is root.
1. Run the following command to determine PRIMARY:
   
   ```
   rs.status()
   ```
   
   The output should be as follows:
  
   ```
   {
           "set" : "datars1",
           "date" : ISODate("2019-09-02T13:06:07.454Z"),
           "myState" : 1,
           "term" : NumberLong(1),
           "syncingTo" : "",
           "syncSourceHost" : "",
           "syncSourceId" : -1,
           "heartbeatIntervalMillis" : NumberLong(2000),
           "optimes" : {
                   "lastCommittedOpTime" : {
                           "ts" : Timestamp(1567429550, 3),
                           "t" : NumberLong(1)
                   },
                   "appliedOpTime" : {
                           "ts" : Timestamp(1567429550, 3),
                           "t" : NumberLong(1)
                   },
                   "durableOpTime" : {
                           "ts" : Timestamp(1567429550, 3),
                           "t" : NumberLong(1)
                   }
           },
           "members" : [
                   {
                           "_id" : 0,
                           "name" : "datars10-0.datars1.mongo-cluster.svc.cluster.local:27017",
                           "health" : 1,
                           "state" : 1,
                           "stateStr" : "PRIMARY",
                           "uptime" : 446706,
                           "optime" : {
                                   "ts" : Timestamp(1567429550, 3),
                                   "t" : NumberLong(1)
                           },
                           "optimeDate" : ISODate("2019-09-02T13:05:50Z"),
                           "syncingTo" : "",
                           "syncSourceHost" : "",
                           "syncSourceId" : -1,
                           "infoMessage" : "",
                           "electionTime" : Timestamp(1566982943, 1),
                           "electionDate" : ISODate("2019-08-28T09:02:23Z"),
                           "configVersion" : 1,
                           "self" : true,
                           "lastHeartbeatMessage" : ""
                   },
                   {
                           "_id" : 1,
                           "name" : "datars11-0.datars1.mongo-cluster.svc.cluster.local:27017",
                           "health" : 1,
                           "state" : 2,
                           "stateStr" : "SECONDARY",
                           "uptime" : 446634,
                           "optime" : {
                                   "ts" : Timestamp(1567429550, 3),
                                   "t" : NumberLong(1)
                           },
                           "optimeDurable" : {
                                   "ts" : Timestamp(1567429550, 3),
                                   "t" : NumberLong(1)
                           },
                           "optimeDate" : ISODate("2019-09-02T13:05:50Z"),
                           "optimeDurableDate" : ISODate("2019-09-02T13:05:50Z"),
                           "lastHeartbeat" : ISODate("2019-09-02T13:06:05.934Z"),
                           "lastHeartbeatRecv" : ISODate("2019-09-02T13:06:06.029Z"),
                           "pingMs" : NumberLong(0),
                           "lastHeartbeatMessage" : "",
                           "syncingTo" : "datars10-0.datars1.mongo-cluster.svc.cluster.local:27017",
                           "syncSourceHost" : "datars10-0.datars1.mongo-cluster.svc.cluster.local:27017",
                           "syncSourceId" : 0,
                           "infoMessage" : "",
                           "configVersion" : 1
                   },
                   {
                           "_id" : 2,
                           "name" : "datars12-0.datars1.mongo-cluster.svc.cluster.local:27017",
                           "health" : 1,
                           "state" : 2,
                           "stateStr" : "SECONDARY",
                           "uptime" : 446634,
                           "optime" : {
                                   "ts" : Timestamp(1567429550, 3),
                                   "t" : NumberLong(1)
                           },
                           "optimeDurable" : {
                                   "ts" : Timestamp(1567429550, 3),
                                   "t" : NumberLong(1)
                           },
                           "optimeDate" : ISODate("2019-09-02T13:05:50Z"),
                           "optimeDurableDate" : ISODate("2019-09-02T13:05:50Z"),
                           "lastHeartbeat" : ISODate("2019-09-02T13:06:05.856Z"),
                           "lastHeartbeatRecv" : ISODate("2019-09-02T13:06:05.856Z"),
                           "pingMs" : NumberLong(5),
                           "lastHeartbeatMessage" : "",
                           "syncingTo" : "datars10-0.datars1.mongo-cluster.svc.cluster.local:27017",
                           "syncSourceHost" : "datars10-0.datars1.mongo-cluster.svc.cluster.local:27017",
                           "syncSourceId" : 0,
                           "infoMessage" : "",
                           "configVersion" : 1
                   }
           ],
           "ok" : 1
   }
   ```

1. Find the primary member by line, `"stateStr" : "PRIMARY"`. In the above example, it is `"name" : "datars10-0.datars1.mongo-cluster.svc.cluster.local:27017"`.
1. Navigate to the terminal of the primary datars1X-0.
1. Execute the following command:
   
   ```
   mongosh admin -u <current_root_user> -p <current_root_password>
   ```
   
   The default user is root. The default password is root.    
  
1. To update the password, execute the following command:
   
   ```
   db.changeUserPassword('<current_user>','<new_password>')
   ```
   
1. Close the terminal.

Repeat the above steps for `datars2X-0`, `datars3X-0` (`datars4X-0`, `datars5X-0`, and `datars6X-0` if you are using DR scheme), and `cnfrsX-0`. 

## Change Backup Password in Services

If you have changed the password of a backup user in MongoDB, you need to update it in the services.

To change the backup password in services, implement the following steps:

**Update Secret**

For OpenShift:

1. Open the OpenShift UI Console.
1. Select the **MongoDB** project.
1. Navigate to **Resources > Secrets**.
1. From the **Secrets** drop-down list, select **mongodb-backup-credentials**.
1. From the **Actions** drop-down list, select **Edit YAML**.
1. Replace the value of **data.password** with a new password encoded to base64.
1. Click **Save**.

For Kubernetes:

1. Open the Kubernetes UI Console.
1. Select the **MongoDB** namespace.
1. Navigate to **Config and Storage > Secrets**.
1. From the **Secrets** drop-down list, select **mongodb-backup-credentials**.
1. In the upper right corner, click the pencil icon to edit.
1. Click the **YAML** tab.
1. Replace the value of **data.password** with the new password encoded to base64.
1. Click **Update**.

**Restart Backup Daemon**

For OpenShift:

1. Navigate to **Applications > Pods**.
1. From the pod drop-down list, select **mongodb-backup-daemon**.
1. From the **Actions** drop-down list, select **Delete**.
1. In the pop-up window, click **Delete**. Wait until the redeployment is done.

For Kubernetes:

1. Navigate to **Workloads > Pods**.
1. From the pod drop-down list, select **mongodb-backup-daemon**.
1. In the upper right corner, click the trash icon to delete.
1. In the pop-up window, click **Delete**. Wait until the redeployment is done.

**Restart DBaaS Adapter**

For OpenShift:

1. Navigate to **Applications > Deployments**.
1. Click **dbaas-mongo-adapter**.
1. From the **Actions** drop-down list, select **Delete**.
1. In the pop-up window, click **Delete**. Wait until the redeployment is done.

For Kubernetes:

1. Navigate to **Workloads > Deployments**.
1. Click **dbaas-mongo-adapter**.
1. In the upper right corner, click the trash icon to delete.
1. In the pop-up window, click **Delete**. Wait until the redeployment is done.

## Change Restore Password in Services

If you have changed the password of a restore user in MongoDB, you need to update it in the services.

To change the restore password in services, implement the following steps:

**Update Secret**

For OpenShift:

1. Open the OpenShift UI Console.
1. Select the **MongoDB** project.
1. Navigate to **Resources > Secrets**.
1. From the **Secrets** drop-down list, select **mongodb-restore-credentials**.
1. From the **Actions** drop-down list, select **Edit YAML**.
1. Replace the value of **data.password** with the new password encoded to base64.
1. Click **Save**.

For Kubernetes:

1. Open the Kubernetes UI Console.
1. Select the **MongoDB** namespace.
1. Navigate to **Config and Storage > Secrets**.
1. From the **Secrets** drop-down list, select **mongodb-restore-credentials**.
1. In the upper right corner, click the pencil icon to edit.
1. Click the **YAML** tab.
1. Replace the value of **data.password** with the new password encoded to base64.
1. Click **Update**.

**Restart Backup Daemon**

For OpenShift:

1. Navigate to **Applications > Pods**.
1. From the pod list, select **mongodb-backup-daemon**.
1. From the **Actions** drop-down list, select **Delete**.
1. In the pop-up window, click **Delete**. Wait until the redeployment is done.

For Kubernetes:

1. Navigate to **Workloads > Pods**.
1. From the pod drop-down list, select **mongodb-backup-daemon**.
1. In the upper right corner, click the trash icon to delete.
1. In the pop-up window, click **Delete**. Wait until the redeployment is done.

**Restart DBaaS Adapter**

For OpenShift:

1. Navigate to **Applications > Deployments**.
1. Click **dbaas-mongo-adapter**.
1. From the **Actions** drop-down list, select **Delete**.
1. In the pop-up window, click **Delete**. Wait until the redeployment is done.

For Kubernetes:

1. Navigate to **Workloads > Deployments**.
1. Click **dbaas-mongo-adapter**.
1. In the upper right corner, click the trash icon to delete.
1. In the pop-up window, click **Delete**. Wait until the redeployment is done.

## Change Monitoring Password in Services

If you have changed the password of a monitoring user in MongoDB, you need to update it in the services.

To change the monitoring password in services, implement the following steps:

**Update Secret**

For OpenShift:

1. Open the OpenShift UI Console.
1. Select the **MongoDB** project.
1. Navigate to **Resources > Secrets**.
1. From the **Secrets** drop-down list, select **mongodb-monitoring-credentials**.
1. From the **Actions** drop-down list box, select **Edit YAML**.
1. Replace the value of **data.password** with the new password encoded to base64.
1. Click **Save**.

For Kubernetes:

1. Open the Kubernetes UI Console.
1. Select the **MongoDB** namespace.
1. Navigate to **Config and Storage > Secrets**.
1. From the **Secrets** drop-down list, select **mongodb-monitoring-credentials**.
1. In the upper right corner, click the pencil icon to edit.
1. Click the **YAML** tab.
1. Replace the value of **data.password** with the new password encoded to base64.
1. Click **Update**.

**Restart Monitoring Agent**

For OpenShift:

1. Navigate to **Applications > Deployments**.
1. Click **mongodb-monitoring-agent**.
1. From the **Actions** drop-down list, select **Delete**.
1. In the pop-up window, click **Delete**. Wait until the redeployment is done.

For Kubernetes:

1. Navigate to **Workloads > Deployments**.
1. Click **mongodb-monitoring-agent**.
1. In the upper right corner, click the trash icon to delete.
1. In the pop-up window, click **Delete**. Wait until the redeployment is done.

## Change DBaaS Password in Services

If you have changed the password of a DBaaS user in MongoDB, you need to update it in the services.

To change the DBaaS password in services, implement the following steps:

**Update Secret**

For OpenShift:

1. Open the OpenShift UI Console.
1. Select the **MongoDB** project.
1. Navigate to **Resources > Secrets**.
1. From the **Secrets** drop-down list, select **mongo-dbaas-admin-credentials-secret**.
1. From the **Actions** drop-down list, select **Edit YAML**.
1. Replace the value of **data.password** with the new password encoded to base64.
1. Click **Save**.

For Kubernetes:

1. Open the Kubernetes UI Console.
1. Select the **MongoDB** namespace.
1. Navigate to **Config and Storage > Secrets**.
1. From the **Secrets** drop-down list, select **mongo-dbaas-admin-credentials-secret**.
1. In the upper right corner, click the pencil icon to edit.
1. Click the **YAML** tab.
1. Replace the value of **data.password** with the new password encoded to base64.
1. Click **Update**.

**Restart DBaaS Adapter**

For OpenShift:

1. Navigate to **Applications > Deployments**.
1. Click **dbaas-mongo-adapter**.
1. From the **Actions** drop-down list, select **Delete**.
1. In the pop-up window, click **Delete**. Wait until the redeployment is done.

For Kubernetes:

1. Navigate to **Workloads > Deployments**.
1. Click **dbaas-mongo-adapter**.
1. In the upper right corner, click the trash icon to delete.
1. In the pop-up window, click **Delete**. Wait until the redeployment is done.

# Change Password of DBaaS Adapter REST API

To change the password of the DBaaS adapter REST API, implement the following steps:

**Update Secret**

For OpenShift:

1. Open the OpenShift UI Console.
1. Select the **MongoDB** project.
1. Navigate to **Resources > Secrets**.
1. From the **Secrets** drop-down list, select **dbaas-aggregator-credentials**.
1. From the **Actions** drop-down list, select **Edit YAML**.
1. Replace the value of **data.password** with the new password encoded to base64.
1. Click **Save**.

For Kubernetes:

1. Open the Kubernetes UI Console.
1. Select the **MongoDB** namespace.
1. Navigate to **Config and Storage > Secrets**.
1. From the **Secrets** drop-down list, select **dbaas-aggregator-credentials**.
1. In the upper right corner, click the pencil icon to edit.
1. Click the **YAML** tab.
1. Replace the value of **data.password** with the new password encoded to base64.
1. Click **Update**.

**Restart DBaaS Adapter**

For OpenShift:

1. Navigate to **Applications > Deployments**.
1. Click **dbaas-mongo-adapter**.
1. From the **Actions** drop-down list, select **Delete**.
1. In the pop-up window, click **Delete**. Wait until the redeployment is done.

For Kubernetes:

1. Navigate to **Workloads > Deployments**.
1. Click **dbaas-mongo-adapter**.
1. In the upper right corner, click the trash icon to delete.
1. In the pop-up window, click **Delete**. Wait until the redeployment is done.

# Change Password of DBaaS Aggregator REST API

To change the password of DBaaS aggregator REST API, follow the below sequence:

**Update Secret**

For OpenShift:

1. Open the OpenShift UI Console.
1. Select the **MongoDB** project.
1. Navigate to **Resources > Secrets**.
1. From the **Secrets** drop-down list, select **dbaas-aggregator-registration-credentials**.
1. From the **Actions** drop-down list, select **Edit YAML**.
1. Replace the value of **data.password** with the new password encoded to base64.
1. Click **Save**.

For Kubernetes:

1. Open the Kubernetes UI Console.
1. Select the **MongoDB** namespace.
1. Navigate to **Config and Storage > Secrets**.
1. From the **Secrets** drop-down list, select **dbaas-aggregator-registration-credentials**.
1. In the upper right corner, click the pencil icon to edit.
1. Click the **YAML** tab.
1. Replace the value of **data.password** with the new password encoded to base64.
1. Click **Update**.

**Restart DBaaS Adapter**

For OpenShift:

1. Navigate to **Applications > Deployments**.
1. Click **dbaas-mongo-adapter**.
1. From the **Actions** drop-down list, select **Delete**.
1. In the pop-up window, click **Delete**. Wait until the redeployment is done.

For Kubernetes:

1. Navigate to **Workloads > Deployments**.
1. Click **dbaas-mongo-adapter**.
1. In the upper right corner, click the trash icon to delete.
1. In the pop-up window, click **Delete**. Wait until the redeployment is done.

# Change Password of InfluxDB

To change the password of InfluxDB, you need to update the secret.

For OpenShift:

1. Open the OpenShift UI Console.
1. Select the **MongoDB** project.
1. Navigate to **Applications > Deployments**.
1. Click **mongodb-monitoring-agent**.
1. Click the **Environment** tab.
1. Enter the new password in the **INFLUXDB_PASSWORD** field.
1. Click **Save**.

For Kubernetes:

1. Open the Kubernetes UI Console.
1. Select the **MongoDB** namespace.
1. Navigate to **Config and Storage > Secrets**.
1. From the **Secrets** drop-down list, select **mongodb-monitoring-agent**.
1. In the upper right corner, click the pencil icon to edit.
1. Click the **YAML**.
1. Find the `INFLUXDB_PASSWORD` item under `spec.template.spec.containers.env`.
1. Replace the value for the `INFLUXDB_PASSWORD` item with the new password.
1. Click **Update**.

# Remove MongoDB

The process of removing MongoDB is described below.

## Remove MongoDB from DBAAS Aggregator

For OpenShift:

1. Open the OpenShift UI Console.
1. Select the **MongoDB** project.
1. Navigate to **Applications > Pods**.
1. Click the **dbaas-mongo-adapter** pod.
1. Click the **Terminal** tab.
1. Paste the following command to delete a physical database.

```
curl -v -XDELETE -u ${DBAAS_AGGREGATOR_REGISTRATION_USERNAME}:${DBAAS_AGGREGATOR_REGISTRATION_PASSWORD} ${DBAAS_AGGREGATOR_REGISTRATION_ADDRESS}/api/v1/dbaas/mongodb/physical_databases/${DBAAS_AGGREGATOR_PHYSICAL_DATABASE_IDENTIFIER}
```

Or through CLI:

```
oc rsh <mongodb dbaas adapter pod> /bin/sh
curl -v -XDELETE -u ${DBAAS_AGGREGATOR_REGISTRATION_USERNAME}:${DBAAS_AGGREGATOR_REGISTRATION_PASSWORD} ${DBAAS_AGGREGATOR_REGISTRATION_ADDRESS}/api/v1/dbaas/mongodb/physical_databases/${DBAAS_AGGREGATOR_PHYSICAL_DATABASE_IDENTIFIER}
```

For Kubernetes:

1. Open the Kubernetes UI Console.
1. Select the **MongoDB** namespace.
1. Navigate to **Pods**.
1. Click the **dbaas-mongo-adapter** pod.
1. In the upper right corner, click the terminal icon to enter into a shell.
1. Paste the following command to delete a physical database.

```
curl -v -XDELETE -u ${DBAAS_AGGREGATOR_REGISTRATION_USERNAME}:${DBAAS_AGGREGATOR_REGISTRATION_PASSWORD} ${DBAAS_AGGREGATOR_REGISTRATION_ADDRESS}/api/v1/dbaas/mongodb/physical_databases/${DBAAS_AGGREGATOR_PHYSICAL_DATABASE_IDENTIFIER}
```

Or through CLI:

```
kubectl exec -t -i <mongodb dbaas adapter pod> /bin/sh
curl -v -XDELETE -u ${DBAAS_AGGREGATOR_REGISTRATION_USERNAME}:${DBAAS_AGGREGATOR_REGISTRATION_PASSWORD} ${DBAAS_AGGREGATOR_REGISTRATION_ADDRESS}/api/v1/dbaas/mongodb/physical_databases/${DBAAS_AGGREGATOR_PHYSICAL_DATABASE_IDENTIFIER}
```

## Uninstall Helm Release

1. The list of current installations.

```
helm list
```

2. Delete the mongodb helm installation.

```
helm uninstall <name>
```

## Cleanup Namespace:

For OpenShift and Kubernetes:

1. Log in to the cloud.
2. Delete MongoDB.

``` 
[oc | kubectl ] delete all --all -n <mongo namespace>
```

3. Delete all pvc.

```
[oc | kubectl ]  delete pvc --all -n <mongo namespace>
```

4. Delete all mongodb configmaps.

```
[oc | kubectl ]  delete configmap --all -n <mongo namespace>
```

# If Operator Needs to be Restarted

The Mongodb operator is designed in a way that restart will not execute any steps if Mongoservice CR has no changes.    
To force Mongodb operator to re-execute all steps:

1. Edit the Config Map "last-applied-configuration-info" - delete value of key `summary-spec` and save.
2. Restart the operator pod.
