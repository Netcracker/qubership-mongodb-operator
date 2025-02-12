This section describes the MongoDB dashboards, metrics, and their significance.

# Dashboard for Prometheus Metrics

An overview of the MongoDB dashboard is shown in the following image.

![Dashboard Overview](/docs/public/images/grafana-prometheus-overview.png)

## Metrics

This section describes the metrics and their meanings.

### Cluster Overview

* `Shards` - Displays the total number of shards in the Cluster.
* `Databases` - Displays the total number of databases.
* `Sharding databases` - Displays the total number of databases in the cluster.
* `Collections` - Displays the total number of collections.
* `Sharded collections` - Displays the total number of collections with sharding enabled.
* `Mongos status` - Displays the Mongos pod status.
* `CNFRS status` - Displays the CNFRS pod status.
* `DATARS status` - Displays the DATARS pod status.

### All Pods: Disk

* `Disk I/O Utilization` - Displays the data read/write in bytes.
* `Disk I/O Utilization By Pods` - Displays the data read/write in bytes by pods.
* `IOps` - Displays the count of writes/reads completed.
* `IOps By Pods` - Displays the count of writes/reads completed by pods.

### All Pods: Network

* `Receive/Transmit Bandwidth` - Displays the overall incoming and outgoing network traffic in bytes per second.
* `Receive/Transmit Bandwidth Usage By Pods` - Displays the overall incoming and outgoing network traffic in bytes per second by pods.
* `Rate of Received/Transmitted Packets` - Displays the overall incoming and outgoing packets per second for the namespace.
* `Rate of Received/Transmitted Packets By Pods` - Displays the overall incoming and outgoing packets per second by pods for the namespace.

### CPU/RAM usage

* `DATARS RAM Usage` - Displays the DATARS pod memory usage.
* `CNFRS RAM Usage` - Displays the CNFRS pod memory usage.
* `DATARS CPU Usage` - Displays the DATARS pod CPU usage.
* `CNFRS CPU Usage` - Displays the CNFRS pod CPU usage.

### Data size

* `DB data size` - Displays the DB data size per DATARS.
* `Top 5 largest DB by data size` - Displays the top 5 largest DB by data size.
* `DB index size` - Displays the DB index size per DATARS.

### Query Metrics

* `Query Operations` - Displays the rate of query operations by query types.
* `Connections` - Displays the available and current connections count per replica set.
* `Page faults` - Displays the page faults.

### Resource Metrics

* `Oplog Size` - Displays the oplog size.
* `Oplog Recovery Window` - Displays the oplog recovery window.
* `Memory` - Displays the average resident and virtual memory.
* `Secondaries Replication Lag` - Displays the secondaries replication lag.

### DBaaS Adapter

* `Status` - Displays the DBaaS Adapter pod status.
* `CPU usage` - Displays the CPU usage of the DBaaS Adapter pod in the Namespace. The red lines show the limits of the CPU's usage.
  The values are indicated in Millicores. It is a special Kubernetes metric where the CPU core is split into 1000 units.
* `Memory usage` - Displays the RAM usage of the DBaaS Adapter pod in the Namespace. The red lines show the limits of the RAM usage.
* `GC Duration` - Displays the duration of the GC.
* `Threads` - Displays the number of user and daemon threads.
* `Requests Count` - Displays the total number of requests the DBaaS Adapter has received per response status.
* `DBaaS API Requests Duration` - Displays the request's duration per each DBaaS Adapter endpoint.

### Backup Daemon

* `Status` - Displays the Backup Daemon pod status.
* `CPU usage` - Displays the CPU usage of the Backup Daemon pod in the Namespace. The red lines show the limits of the CPU's usage.
  The values are indicated in Millicores. It is a special Kubernetes metric where the CPU core is split into 1000 units.
* `Memory usage` - Displays the RAM usage of the Backup Daemon pod in the Namespace. The red lines show the limits of the RAM usage.
* `Last Backup Status` - Displays the last backup status.
* `Last Backup Size` - Displays the last backup size.
* `Dumps count` - Displays the amount of backups done.
* `Storage Space Usage` - Displays the space used by backups against the total space.
* `Storage Inodes Usage` - Displays the amount of used inodes against the total inodes.
