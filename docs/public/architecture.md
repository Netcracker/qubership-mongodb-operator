The following topics are discussed in this chapter:

[[_TOC_]]

# MongoDB Overview

MongoDB is a popular NoSQL database management system that falls under the category of document-oriented databases. It is designed to store and manage unstructured or semi-structured data in JSON-like documents, making it highly flexible and scalable. 

MongoDB has the following features:

* Flexible schema - MongoDB's document-based data model allows you to store varying structures of documents in the same collection without requiring a fixed schema.
* JSON-like documents - Data is stored in BSON (Binary JSON) format, which is a binary representation of JSON-like documents.
* Scalability - MongoDB can scale horizontally by sharding data across multiple servers, which helps in handling large-scale applications and workloads.
* High performance - It offers fast read and write operations, making it suitable for real-time applications.
* Rich query language - MongoDB supports complex queries, indexes, and aggregations, allowing for powerful data retrieval and manipulation.
* Replication - MongoDB supports replica sets, providing automatic failover and data redundancy for high availability.
* Geospatial queries - It includes support for geospatial indexing and querying, making it suitable for location-based applications.
* Ad hoc queries - MongoDB allows you to perform dynamic queries on documents using a flexible query language.

## Glossary

Replica set - It is a group of MongoDB servers that maintain the same data set, providing high availability and data redundancy. The primary purpose of a replica set is to ensure that the data remains available and accessible even in the event of hardware failures or planned maintenance.

A replica set consists of the following components:

* Primary node that receives all write operations from clients. It is the only node in the replica set that accepts write operations. Each replica set can have only one primary node at a time.
* Secondary nodes that replicate data from the primary node and can serve read operations. They maintain a copy of the data from the primary node, and their purpose is to provide high availability and load balancing for read operations.
* Arbiter nodes (optional) that do not store data. Their primary function is to participate in the election of a new primary node in the event of the current primary node's failure. They help to ensure an odd number of voting members in the replica set.

Shard - It is an individual MongoDB instance or replica set that holds a subset of an entire dataset. Each shard is responsible for storing a specific range of data based on a sharding key, which is a field or fields used to determine how data is distributed across shards. Shards operate independently of each other, allowing for parallel read and write operations on different parts of a dataset, which significantly enhances performance and scalability.

# MongoDB Components

The following image displays the various MongoDB components:

![MongoDB Components](/docs/public/images/mongodb_components.png)

## MongoDB Operator

MongoDB Operator is a mandatory microservice written with Operator-SDK and designed specifically for Kubernetes environments.
It simplifies the deployment and management of MongoDB clusters, which are critical for distributed coordination.
In addition to deploying the MongoDB cluster, the operator also takes care of managing supplementary services, ensuring seamless integration, and efficient resource utilization.
MongoDB operator also performs upgrade scenario without MongoDB downtime.

## MongoDB

MongoDB cluster is delivered using an official docker image.

## MongoDB Backup Daemon

MongoDB Backup Daemon is a microservice that offers a convenient REST API for performing backups and restores of MongoDB databases.
It enables users to initiate full or granular backups and restores programmatically, making it easier to automate these processes.
Additionally, the daemon allows users to schedule regular backups, ensuring data protection and disaster recovery.
Furthermore, it offers the capability to store backups on remote S3 storage, providing a secure and scalable solution for long-term data retention.

## MongoDB DBaaS Adapter

MongoDB DBaaS adapter is a microservice for integration with DBaaS, which allows to manage logical databases through API.

## Robot Tests

Robot tests is a microservice that performs integration testing after all components of MongoDB deployment are installed.

# Supported Deployment Schemes

The following are the supported MongoDB deployment schemes.

## On-Prem

The On-prem deployment scheme includes Non-HA deployment scheme, HA deployment scheme, Artiber deployment scheme, Simple deployment scheme, and DR deployment scheme. 

### Non-HA Deployment Scheme

Not for production environment!
The following image illustrates the structure of the Non-HA deployment scheme.

![MongoDB Single](/docs/public/images/mongodb_single.png)

### HA Deployment Scheme

The default MongoDB HA deployment Scheme consists of the following deployments:

* MongoDB Operator 
* 1 Mongos pod, which routes client requests to the appropriate shards within a sharded cluster
* 1 Config Replica Set (CNFRS)
* 3 shard, each of them is Data Replica Sets (DATARS)
* MongoDB Backup Daemon
* MongoDB DBaaS Adapter
* MongoDB Prometheus Exporter
* Robot Tests

![MongoDB HA](/docs/public/images/mongodb_ha.png)

**Note**

* All components except for MongoDB operator are deployed by the MongoDB Operator.
* PRIMARY in replica sets can be in any nodes, including a case when 2 or 3 PRIMARY are on the same node.
* Services are omitted from the diagram as there are many of them; 1 service for each replica set, operator, DBaaS adapter, and backup daemon.

### Non-sharded Deployment Scheme

Starting with 1.18.0 release, MongoDB can be deployed as a non-sharded scheme. In this case, a single MongoDB replica set is deployed without MongoDBs' balancer and config servers. This schema is handy in case where the sharding feature is not needed. At the same time, this schema is less problem prone for DR scenario as it requires less steps to perform DR actions.

The deployed replica set provides High Availability with fault tolerance of 1; it has 1 PRIMARY and 2 SECONDARIES.

![MongoDB Simple](/docs/public/images/mongodb_simple.png)

### DR Deployment Scheme

The Disaster Recovery scheme of MongoDB deployment assumes that MongoDB is deployed to both sides on separate Kubernetes environments with pod-to-pod connectivity between them.
For more information, refer to [MongoDB Disaster Recovery](dr_operations.md) in _Cloud Platform Disaster Recovery Guide_.

![MongoDB DR](/docs/public/images/mongodb_dr.png)

## Google Cloud

Not Applicable; the default HA scheme is used for deployment to Google Cloud.

## AWS

Not Applicable; the default HA scheme is used for deployment to AWS.

## Azure

Not Applicable; the default HA scheme is used for deployment to Azure.
