The following topics are discussed in this chapter:

[[_TOC_]]

# Overview

MongoDB can be deployed in the Disaster Recovery (DR) scheme with clusters in _active_ and _standby_ modes using the configuration described in the [DR Schema](installation_guide.md#dr-schema) section of the MongoDB Cluster Installation Procedure chapter in _Cloud Platform Installation_.    
For more information about the deployment scheme overview, refer to [MongoDB Cluster Deployment Model](architecture.md#dr-deployment-scheme) in _Cloud Platform Installation_.    

# DR Scenarios

The following sections describe the various DR scenarios in MongoDB.

## Switchover

The sequence of a switchover operation for MongoDB is as follows:

1. Switch the current active cluster to standby.
    * Reconfigure CNFRS and DATARS replica sets so that they cannot become `PRIMARY`.
    * Scale-down supplementary services and `mongos`.    
2. Switch the current standby cluster to active.
    * Reconfigure CNFRS and DATARS replica sets so that they can become `PRIMARY`.
    * Scale-up supplementary services and `mongos`.

![Switchover](/docs/public/images/switchover.png)

## Failover

The sequence of a failover operation for MongoDB is as follows:

1. Switch the current standby cluster to active.
    * Reconfigure CNFRS and DATARS replica sets so that they can become `PRIMARY`.
    * Scale-up supplementary services and `mongos`.

![Failover](/docs/public/images/failover.png)

# REST API

The following are the details of the REST API. 

## Get Status

Request:

```bash
curl -GET http://mongodb-operator.<NAMESPACE>:8080/sitemanager
```

Where:

* `<NAMESPACE>` is the OpenShift/Kubernetes project or namespace of the MongoDB cluster.

`GET` `sitemanager` endpoint returns json with `mode` and its `status` of the current cluster.

Response:

```json
{"mode": "active", "status": "done"}
```

* `mode`
    * `active` - Cluster is in the active mode.
    * `standby` - Cluster is in the standby mode.
    * `disabled` - MongoDB cluster is disabled.
* `status`
    * `running` - The switchover is in progress.
    * `done` - The switchover is successful.
    * `failed` - The switchover has failed.  

## Health

Request:

```bash
curl -GET http://mongodb-operator.<NAMESPACE>:8080/healthz
```

Where:

* `<NAMESPACE>` is the OpenShift/Kubernetes project or namespace of the MongoDB cluster.

`GET` `sitemanager` endpoint returns json with `status` of the current cluster.

Response:

```json
{"status": "up"}
```

* `status`
    * `up` - MongoDB cluster is ready.
    * `down` - Some of the MongoDB cluster components are down or replication is partially broken, but the database is running.
    * `degraded` - MongoDB cluster is not ready.

## Switch Mode

Request:

```bash
curl -XPOST -d '{"mode": "<MODE>"}' -H "Content-Type: application/json" \
http://mongodb-operator.<NAMESPACE>:8080/sitemanager
```

Where:

* `<NAMESPACE>` is the OpenShift/Kubernetes project or namespace of the MongoDB cluster.
* `<MODE>` is the mode to be applied to the MongoDB cluster side. The possible mode values are as follows:
    * `active` - Switch ON active mode.
    * `standby` - Switch ON standby mode.
    * `disabled` - Disable the MongoDB cluster.
