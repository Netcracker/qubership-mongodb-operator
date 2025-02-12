This section provides information about configuring the MongoDB Cluster Installation and Maintenance Scripts.  

# List of Service Ports and Dependencies

Netcracker MongoDB Service consists of MongoDB, Mongo Backup Daemon, Monitoring Agent, and Mongo DBaaS Adapter.

The list of service ports and dependencies are described in the following table:

| Component        | Exposed Ports                                                                                                        | Dependencies (Used Ports)                                                                                        |
| ---------------- | -------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| MongoDB          | `27017/TCP` The MongoDB server port, used by client applications to access the database.                                 | Backup Daemon - `8080/TCP, 8443/TCP (TLS)`                                                                       |
| Backup Daemon    | `8080/TCP, 8443/TCP (TLS)` The port for backup, restore, and obtaining status of full and granular database backups. | MongoDB - `27017/TCP`                                                                                            |
| Monitoring Agent | 9216/TCP, 9217/TCP, 9218/TCP, 9219/TCP, 9220/TCP                                                                     | MongoDB - `27017/TCP`, Backup Daemon - `8080/TCP, 8443/TCP (TLS)`                                                |
| DBaaS Adapter    | `8080/TCP, 8443/TCP (TLS)` The DBaaS API port, used by the DBaaS aggregator to manage Mongo database.                | MongoDB - `27017/TCP`, Backup Daemon - `8080/TCP, 8443/TCP (TLS)`, DBaaS Aggregator - `8080/TCP, 8443/TCP (TLS)` |
| MongoDB operator | `8069/TCP` The MongoDB operator port, used by DR daemon to manage the DR operations.                                     | MongoDB - `27017/TCP`                                                                                            |
| DR daemon        | `8080/TCP, 8443/TCP (TLS)` The MongoDB DR daemon port, used by Site Manager to send the DR operations.                   | MongoDB operator - `8069/TCP`                                                                                    |

The most used service port is the `27017/TCP` MongoDB server port that provides the database interface for client applications.
Backup daemon port is used in backup managers to provide consistent backups within the solution.
DBaaS performs database management operations such as create, delete, and backup through the DBaaS Adapter API `8080/TCP, 8443/TCP (TLS)` port.

The following two external interfaces are required for the service:

* DBaaS Aggregator `8080/TCP, 8443/TCP (TLS)` for registering physical database cluster.

The `27017/TCP` MongoDB server port is also used internally for replication between shards and replicas. MongoDB Service does not expose dynamic ports.

# Local MongoDB Users

The following table describes the default users created during Netcracker MongoDB deploy.

| Role                 | Default username | Default Password | Description                                                                         |
| -------------------- | ---------------- | ---------------- | ----------------------------------------------------------------------------------- |
| root                 | root             | root             | The root user                                                                       |
| backup               | backup           | backup           | The MongoDB user with 'backup' privileges.                                          |
| restore              | restore          | restore          | The MongoDB user with 'restore' privileges.                                         |
| clusterMonitor       | monitoring       | monitoring       | The MongoDB user with 'monitoring' privileges.                                      |
| userAdminAnyDatabase | dbaas            | dbaas            | The user with privileges to create/delete/update DB and users by request from DBAAS |

# Replication Lag

The secondary members replicate data continuously after the initial synchronization. The secondary members copy the oplog from their synchronization from source and apply these operations in an asynchronous process.

The secondary members may automatically change their synchronization from source as required based on the changes in the ping time and the state of other members’ replication.

Under various exceptional situations, updates to a secondary’s oplog might lag behind the desired performance time. 
You can receive the information regarding current replication status by executing the `db.getReplicationInfo()` command. For information about the command, refer to [https://docs.mongodb.com/manual/reference/method/db.getReplicationInfo/#db.getReplicationInfo](https://docs.mongodb.com/manual/reference/method/db.getReplicationInfo/#db.getReplicationInfo).

The results of measuring a replication lag depending on a latency between Primary and Secondary are as follows:

| Latency, ms | Replication Lag, sec |
| ----------- | -------------------- |
| 1           | 0.3                  |
| 10          | 0.7                  |
| 100         | 1.3                  |
| 500         | 4.8                  |
| 1000        | > 1 Hour             |

Latency up to 1 second does not significantly affect replication of Gridfs files.

The following environment settings were used:

* MongoDB third party version: 3.4.19
* Deployment schema: HA
* CPU and RAM per datars: 200m CPU and 512Mi
* 1 gbps bandwidth connection

The following synthetic data was used:
* 100000 documents 1Kb size

# Sharding

Sharding is a method of distributing data across multiple machines. MongoDB uses sharding to support deployments with very large data sets and high throughput operations.

The `SHARD_COUNT` parameter specifies how many shards are to be deployed. Currently, it supports the deployment of up to 3 shards.

## Sharding Requirements

Following is the list of sharding requirements:

* The shard key consists of an immutable field or fields that exist in every document in the target collection.
* To shard a non-empty collection, the collection must have an index that starts with the shard key. For empty collections, MongoDB creates the index if the collection does not have an appropriate index for the specified shard key. 

## Sharding Strategy  

This section provides information about the different sharding strategies available.

### Hashed Sharding  

Hashed Sharding involves computing a hash of the shard key field value. Each chunk is then assigned a range based on the hashed shard key values.

While a range of shard keys may be “close”, their hashed values are unlikely to be on the same chunk. Data distribution based on the hashed values facilitates more even data distribution, especially in data sets where the shard key changes monotonically.

However, hashed distribution means that ranged-based queries on the shard key are less likely to target a single shard, resulting in more cluster wide broadcast operation.

### Ranged Sharding

Ranged sharding involves dividing data into ranges based on the shard key values. Each chunk is then assigned a range based on the shard key values.

A range of shard keys, whose values are “close” are more likely to reside on the same chunk. This allows for targeted operations as mongos can route the operations to only the shards that contain the required data.

## Performance with Sharding

This section provides detailed information about performance with sharding.

### Sharded Collection

Insert operations with sharding can be noticeably slower compared to single shard schema. It is due to the time spent on computing sharding key and distributing data among shards.

Following is a chart that shows time in seconds spent to insert 1000000 documents by 1KiB and selects all of them by non-indexed field. The select time by indexed field is almost the same for sharded and non-sharded schemas.

Hardware configuration:

* CPU: 2 cores for each datars 
* RAM: 4GiB for each datars 

![Comparing Performance of Sharded and Non-sharded Schemas](/docs/public/images/sharding_insert_select.png)

You need to be attentive to the following factors when sharding:

* A poorly chosen shard key can lead to lower performance at several instances.
* An ordered insert into a collection takes four times longer than an unordered one.
* A shard key on a value that increases or decreases monotonically is more likely to distribute inserts to a single shard within the cluster. Use hashed sharding for this type of key. 

### Sharded Databases

A sharded mongo cluster stores database in the shard selected by the Round-robin algorithm. This specifies that even without sharding a particular collection, access to the databases can be decentralized.
