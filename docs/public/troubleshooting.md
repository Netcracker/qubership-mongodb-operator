This section provides information on how to locate and fix some of the common MongoDB issues.

## DR Issues - Swithover/Failover failed

This section provides information about the Swithover/Failover failed.

### Description

Swithover/Failover failed.

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **Retry switchover/falover operation:**
   - On the side that fails edit Config Map "last-applied-configuration-info" - delete value of key `summary-spec` and save.
   - Restart the operator pod on the side that fails and wait the side status becomes expected.
   - Retry the switchover/failover operation.

### Recommendations

"Not applicable"

## Common Issues - No Free Space Left on MongoDB PV

This section provides information about the MongoDB issues related to lack of free space on the MongoDB PV.

### Description 

All MongoDB pods can start rebooting if no free space is left on PV.

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **Navigate to mongo-cluster ns:**
    - Scale Down pods `datars10-0`, `datars21-0`, and `datars32-0`.
    - Connect to Nodes/NFS to access the MongoDB PVs content.
    - Remove content of folders on the 3 MongoDB PVs `/data/datars10`, `/data/datars21`, and `/data/datars32` accordingly. Now all other datars pods should be able to start.
    - Remove unnecessary data from MongoDB collections.
    - Perform a `compact` operation on collections according to the instructions in [MongoDB Disk Space Usage not changed after data deleted](#mongodb-disk-space-usage-not-changed-after-data-deleted).
    - Scale Up pods `datars10-0`, `datars21-0`, and `datars32-0`.

### Recommendations

"Not applicable"

## Common Issues - MongoDB Disk Space Usage Not Changed After Data Deleted

This section provides information about the MongoDB issues related MongoDB Disk Space Usage Not Changed After Data Deleted.

### Description

After the data is deleted from a collection, a compact command must be executed on the collection.

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **Navigate to mongo-cluster ns:**
   - Open the terminal of Mongos pod and execute mongosh admin -u root -p <root_pass> --eval="sh.status()".
       In the output, you can view a list of databases and their primary shard, for example:
       `{  "_id" : "mydatabase",  "primary" : "datars3",  "partitioned" : false,  "version" : {  "uuid" : UUID("f9627145-eab6-49c6-aac0-77e41b53d060"),  "lastMod" : 1 } }` - primary shard datars3
   - Navigate to the first pod of shard of your database, in the current example - datars3. Accordingly, navigate to datars30-0 and execute the following:
    
       ```text
        mongosh admin -u root -p <root_pass> --eval='rs.status().members.forEach(mem => {print(mem.name); print("\t"+mem.stateStr)})'
       ```

       The output provides which pod is primary and which are secondary replicas at the moment, for example:
    
       ```text
        datars30-0.datars3.mongo.svc.cluster.local:27017
                SECONDARY
        datars31-0.datars3.mongo.svc.cluster.local:27017
                SECONDARY
        datars32-0.datars3.mongo.svc.cluster.local:27017
                PRIMARY
       ```

   - Navigate to each SECONDARY pods and perform `compact`.
    
       ```text
        mongosh admin -u root -p <root_pass>
        use <db_name>
        db.runCommand({compact: "<collection_name>"})
       ```

   - Navigate to the primary pod and perform `stepDown` and `compact`.
    
       ```text
        mongosh admin -u root -p <root_pass>
        rs.stepDown()
        use <db_name>
        db.runCommand({compact: "<collection_name>"})
       ```

### Recommendations

"Not applicable"

## Common Issues - Index build stuck

This section provides information about the MongoDB issues related to Index build stuck with "Index build: waiting for next action before completing final phase".

### Description

One or more replicas in the replicaset are not in expected state (PRIMARY, SECONDARY). Starting in MongoDB 4.4, index builds on a replica set or sharded cluster build simultaneously across all data-bearing replica set members. The primary requires a minimum number of data-bearing voting members (i.e. commit quorum), including itself, that must complete the build before marking the index as ready for use. 

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **Fix all replicas that are not in expected state (PRIMARY, SECONDARY).**
   - Navigate to mongo-cluster namespace:**
   - Open the terminal of Mongos pod and execute mongosh admin -u root -p <root_pass> --eval="sh.status()".
      In the output, you can view a list of databases and their primary shard, for example:
      `{  "_id" : "mydatabase",  "primary" : "datars3",  "partitioned" : false,  "version" : {  "uuid" : UUID("f9627145-eab6-49c6-aac0-77e41b53d060"),  "lastMod" : 1 } }` - primary shard datars3
   - Navigate to the first pod of shard of database where index build is stuck, in the current example - datars3. Accordingly, navigate to datars30-0 and execute the following:
    
       ```text
        mongosh admin -u root -p <root_pass> --eval='rs.status().members.forEach(mem => {print(mem.name); print("\t"+mem.stateStr)})'
       ```
    
       The output provides which pod is primary and which are secondary replicas at the moment, for example:
    
       ```text
        datars30-0.datars3.mongo.svc.cluster.local:27017
                RECOVERING
        datars31-0.datars3.mongo.svc.cluster.local:27017
                SECONDARY
        datars32-0.datars3.mongo.svc.cluster.local:27017
                PRIMARY
       ```

   - Navigate to the RECOVERING replica (datars30-0 in the example above) clear its data and reboot the pod:
       ```text
       rm -rf /data/datars30/*
       ```
   - After pod starts wait it becomes SECONDARY 
   - Retry index build

### Recommendations

"Not applicable"

## Common Issues - Could not find host matching read preference for set datarsX

This section provides information about the MongoDB issues related to Could not find host matching read preference { mode: "primary" } for set datarsX.

### Description

One or more replicas in replicaset datarX have incorrect status.

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **Fix all replicas that are not in expected state (PRIMARY, SECONDARY).**
   - Navigate to mongo-cluster namespace
   - Navigate to any pod of replicaset `datarsX` (for example to datars30-0 if `X` is 3) and execute the following:
       ```text
        mongosh admin -u root -p <root_pass> --eval='rs.status().members.forEach(mem => {print(mem.name); print("\t"+mem.stateStr)})'
       ```
    
       The output provides which pod is primary and which are secondary replicas at the moment, for example:
    
       ```text
        datars30-0.datars3.mongo.svc.cluster.local:27017
                RECOVERING
        datars31-0.datars3.mongo.svc.cluster.local:27017
                ROLLBACK
        datars32-0.datars3.mongo.svc.cluster.local:27017
                SECONDARY
       ```
    
       **Important**: At least one replica **MUST** have status PRIMARY or SECONDARY. If there are no PRIMARY or SECONDARY replicas do not continue this guide.

   - Navigate to the each replicas with unexcpected status (not PRIMARY or SECONDARY), clear its data and reboot the pod  (datars30-0 and datars31-0 in the example above):
       ```text
        kubectl exec datars30-0 -- sh -c 'rm -rf /data/datars30/*'
        kubectl delete pod datars30-0
        kubectl exec datars31-0 -- sh -c 'rm -rf /data/datars31/*'
        kubectl delete pod datars31-0
       ```
   - After pod starts wait it becomes SECONDARY
       ```text
        datars30-0.datars3.mongo.svc.cluster.local:27017
                SECONDARY
        datars31-0.datars3.mongo.svc.cluster.local:27017
                SECONDARY
        datars32-0.datars3.mongo.svc.cluster.local:27017
                PRIMARY
       ```

### Recommendations

"Not applicable"

## Common Issues - Deploy Failed With Error in Operator Logs

This section provides information about the MongoDB issues related to Deploy Failed With ```Reconciliation exception: PVC pvc-mongodb-main-data-0 'Bound' status waiting failed``` in Operator Logs.

### Description

The Storage Provisioner is configured to create and bind PV when first consumer is created.

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **Disable waiting for PVC binding.**
    - Navigate to the deployment parameter configuration (CMDB).
    - Add the `waitPvcBound: false` parameter to the `mongodb.storage` and `backup.storage` sections.
    - Retry the deployment job.

### Recommendations

"Not applicable"

## Common Issues - Mongodb Statefulset is Not Starting With Error

This section provides information about the MongoDB issues related to Mongodb Statefulset is Not Starting With Error ```Assertion: 28595:2: No such file or directory src/mongo/db/storage/wiredtiger/wiredtiger_kv_engine.cpp 267```.

### Description 

The backend filesystem has failed, and the data for this particular instance is corrupted. 

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **You cannot recover it. However, you can resync the data from other pods.**
   - Look at the pods in your cluster. Usually, there are 9 data pods (datars10-0, datars12-0,...) and 3 cnfrs pods (cnfrs1-0,...). To restore all the data, you need at least one instance of each ReplicaSet. Out of datars10-0, 11-0, and 12-0, at least one must be alive. Same for datars2*, datars3*, and cnfrs*.
   - If at least one pod is alive for all ReplicaSets, you can start recovering the data. If at least one ReplicaSet has all pods down, the data is lost, and you cannot recover it.
   - Scale down the StatefulSet with errors. Navigate to **Applications > StatefulSets > choose a set > actions/edit yaml > change `spec:replicas` to 0 > Save**.
   - If you have access to FS, navigate and clean the folder, **<Place_where_pv_is_stored>/<statefulset_name>**. For example, for **cnfrs1** with PV backed by folder **/mount/cinder-volumes/mongo-backup**, it is **/mount/cinder-volumes/mongo-backup/cnfrs1**.

2. **If you do not have access to FS, you can do the following:** 

   - Log in through OC to OpenShift.
   - Load Statefulset Config using the following command:
   
      ```text
        oc get statefulset <statefulsetname> -o yaml > /tmp/<statefulsetname>.yaml
      ```
   
   - Backup **/tmp/<statefulsetname>.yaml** using the following command: 
   
      ```text
        cp /tmp/<statefulsetname>.yaml /tmp/<statefulsetname>.yaml_back
      ```
   
   - Edit **/tmp/<statefulsetname>.yaml_back**. Change the line that is located at `spec:template:spec:args` and displayed as `"cp /opt/mongo-secret/mongodb-keyfile /data/mongodb-keyfile && chmod 0600 /data/mongodb-keyfile........"` to `"sleep 5000"` and set `spec:replicas:1`.
   - Delete Statefulset using the following command:
   
      ```text
        oc delete statefulset <statefulsetname>
      ```
   
   - Create a new Statefulset using the following command: 
   
      ```text
        oc create -f /tmp/<statefulsetname>.yaml_back
      ```
   
   - Navigate to the OpenShift console of the newly created pod or rsh through OC:
   
      ```text
        oc rsh <statefulsetname>-0
      ```
   
   - Check `/data` for an existing MongoDB data. It should look like the following:
   
      ```text
        sh-4.2# ls /data                                                                                                                                                                    
        WiredTiger                             collection-30--7259052634765019933.wt  index-19--7259052634765019933.wt  index-40--7259052634765019933.wt                                    
        WiredTiger.lock                        collection-33--7259052634765019933.wt  index-20--7259052634765019933.wt  index-42--7259052634765019933.wt                                    
        WiredTiger.turtle                      collection-35--7259052634765019933.wt  index-22--7259052634765019933.wt  index-43--7259052634765019933.wt                                    
        WiredTiger.wt                          collection-39--7259052634765019933.wt  index-24--7259052634765019933.wt  index-44--7259052634765019933.wt                                    
        WiredTigerLAS.wt                       collection-4--7259052634765019933.wt   index-25--7259052634765019933.wt  index-46--7259052634765019933.wt                                    
        _mdb_catalog.wt                        collection-41--7259052634765019933.wt  index-26--7259052634765019933.wt  index-47--7259052634765019933.wt                                    
        collection-0--7259052634765019933.wt   collection-45--7259052634765019933.wt  index-27--7259052634765019933.wt  index-49--7259052634765019933.wt                                    
        collection-12--7259052634765019933.wt  collection-48--7259052634765019933.wt  index-29--7259052634765019933.wt  index-5--7259052634765019933.wt                                     
        collection-13--7259052634765019933.wt  collection-50--7259052634765019933.wt  index-3--7259052634765019933.wt   index-51--7259052634765019933.wt                                    
        collection-15--7259052634765019933.wt  collection-8--7259052634765019933.wt   index-31--7259052634765019933.wt  index-9--7259052634765019933.wt                                     
        collection-18--7259052634765019933.wt  diagnostic.data                        index-32--7259052634765019933.wt  journal                                                             
        collection-2--7259052634765019933.wt   index-1--7259052634765019933.wt        index-34--7259052634765019933.wt  mongod.lock                                                         
        collection-21--7259052634765019933.wt  index-14--7259052634765019933.wt       index-36--7259052634765019933.wt  mongodb-keyfile                                                     
        collection-23--7259052634765019933.wt  index-16--7259052634765019933.wt       index-37--7259052634765019933.wt  sizeStorer.wt                                                       
        collection-28--7259052634765019933.wt  index-17--7259052634765019933.wt       index-38--72590`52634765019933.wt  storage.bson      
      ```
   
   - Delete all the data using the command, ```rm -f /data/*```.
   - After the data is deleted, delete Statefulset using the following command:
   
      ```text
        oc delete statefulset <statefulsetname>
      ```
   
   - Create Statefulset with original configurations using the following command: 
   
      ```text
        oc create -f /tmp/<statefulsetname>.yaml
      ```
   
   - Scale Statefulset using the following command:
   
      ```text
        oc scale <statefulsetname> --replicas=1
      ```
   
   - Wait for the resynchronization. The pod should become **Running: 1/1** in OpenShift.

## Common Issues - Deploy Fails

This section provides information about the MongoDB issues related to Deploy Fails Because of `otherDomainName` and `customDataRSParameters` Parameters.

### Description

Known issue in the release.
```text
  * spec.mongodb.customDataRSParameters: Invalid value: "null": spec.mongodb.customDataRSParameters in body must be of type array: "null"
  * spec.schemaSettings.otherDomainName: Invalid value: "null": spec.schemaSettings.otherDomainName in body must be of type string: "null"
```

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **Set the following parameters:**

    ```text
        schemaSettings.otherDomainName: "cluster.local"
        mongodb.customDataRSParameters: '["logLevel=0"]'
    ```

### Recommendations

"Not applicable"
    
## Common Issues - Mongodb Statefulset is Not Starting With Error and You Do Not Have Enough Alive Pods to Sync From

This section provides information about the MongoDB issues related to Mongodb Statefulset is Not Starting With Error ```Assertion: 28595:2: No such file or directory src/mongo/db/storage/wiredtiger/wiredtiger_kv_engine.cpp 267``` and You Do Not Have Enough Alive Pods to Sync From.

### Description

You cannot recover all data in any easy way, except restoring from backups.

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **If you are OK with an empty, but a working database, do the following:**
   - Clean all the data in all datars and cnfrs pods. For more information, see [Mongodb-statefulset-is-not-starting-with-error-```I - [initandlisten] Assertion: 28595:2: No such file or directory src/mongo/db/storage/wiredtiger/wiredtiger_kv_engine.cpp 267```](#Mongodb Statefulset is not starting with error ```I - [initandlisten] Assertion: 28595:2: No such file or directory src/mongo/db/storage/wiredtiger/wiredtiger_kv_engine.cpp 267```)
   - Clone the repository to the machine with access to cloud.
   - Log in to a proper project through OC.
   - Specify the values for `./scripts/re-init.sh` environment variables in the beginning of the script.
   - Run `./scripts/re-init.sh`.
   - See if the run and smoke tests has finished successfully.

### Recommendations

"Not applicable"

## Common Issues - Pods Do Not Start With Error in Events

This section provides information about the MongoDB issues related to Pods Do Not Start With Error ```context deadline exceeded``` in Events.

### Description

It is a Docker issue.

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **You need to restart the Docker on the effected node. The OPs\IT can do it.**

### Recommendations

"Not applicable"

## Common Issues - Pods Stuck in Pending State

This section provides information about the MongoDB issues related to Pods Stuck in Pending State.

### Description

This happens on older, not updated installations. Old installations used init-containers to manage security keys. For more information, refer to [https://kubernetes.io/docs/concepts/workloads/pods/init-containers/](https://kubernetes.io/docs/concepts/workloads/pods/init-containers/). Despite the init container is exited after the main pod start, if it is deleted from the OpenShift node, the whole pod can fall into the **Pending** state. A cron task deletes the exited containers from the node.

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **Work Around:** 
    - Delete the **Pending** state containers one-by-one. Delete and wait until the new container becomes **Running 1\1**, then delete the next container and so on.

2. **Resolution:** 
   - Update to the newest versions of mongo-cluster.

### Recommendations

"Not applicable"

## Common Issues - Cannot Login to MongoDB

This section provides information about the MongoDB issues related to the following: Cannot Login to MongoDB, Says Unauthorized, When Trying to Login With Root in Logs ```SCRAM-SHA-1 authentication failed for root on admin from client 127.0.0.1:47050 ; UserNotFound: Could not find user root@admin```.

### Description

In the logs of mongos, there is 'can't choose primary for datars${number}. Also, all queries cannot run with error 'can't choose primary for datars${number}'

When trying to enter mongo without credentials, such as:

```
mongosh admin
```

It allows you to log in, but does not allow running any command.

MongoDB has not been initiated yet, the root user has not been created. Until you create a root user, you cannot do anything.

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **Log in to primary of the failed ReplicaSet and run.**

   - If your Replica set is not set, you need to initialize it first. Log in to any replica member and run:
   
      ```text
        rs.initiate({_id:'datars${RS_NUMBER}', members: [{_id:0, host:'datars${RS_NUMBER}0-0.datars${RS_NUMBER}.${PROJECT}.svc.cluster.local:27017'},{_id:1, host:'datars${RS_NUMBER}1-0.datars${RS_NUMBER}.${PROJECT}.svc.cluster.local:27017'},{_id:2, host:'datars${RS_NUMBER}2-0.datars${RS_NUMBER}.${PROJECT}.svc.cluster.local:27017'}]});
      ```
   
   - Login to Primary of ReplicaSet without credentials:
   
      ```text
        mongosh admin
      ```
   
      Run the following command:   
   
      ```text
        use admin;
        db.createUser({user: 'root', pwd: '${MONGO_PASSWORD}', roles: ['root']});
      ```
   
   - Log out and log in to Primary of ReplicaSet again, but with credentials:
   
      ```text
        mongosh -u root -p ${MONGO_PASSWORD} admin
      ```
   
      Run the following command to add the monitoring user: 
   
      ```text
        db.createUser({user: 'monitoring', pwd: '${MONGO_MONITORING_PASSWORD}', roles: [{role:'readWrite', db:'test'}, 'clusterMonitor']});
      ```
   
   - Delete mongos pod to force re-initialization of shard. Check the mongos logs, and see if the following message is displayed:
   
      ```text
        can't resolve host <one of your replicaset pods>
      ```
   
      You may need to force Mongos to resolve the required host.
      Log in to Mongos with the root credentials and run:
   
      ```text
        sh.status()
      ```
   
      The following is displayed:
   
      ```text
        {  "_id" : "datars2",  "host" : "datars2/datars20-0.datars2.lambda-backup-mongo.svc.cluster.local:27017,datars21-0.datars2.lambda-backup-mongo.svc.cluster.local:27017,datars22-0.datars2.lambda-backup-mongo.svc.cluster.local:27017",  "state" : 1 }
        {  "_id" : "datars3",  "host" : "datars3/datars31-0:27017",  "state" : 1 }
      ```
   
      As you see, datars3 is wrong. If you ping datars31-0 from mongos, you get an error:
   
      ```text
        sh-4.2# ping datars31-0
        ping: unknown host datars31-0
      ```
   
      However, if you ping a full host name, it works:
   
      ```text
        PING datars31-0.datars3.mongodb-cluster.svc.cluster.local (10.129.31.59) 56(84) bytes of data.
        64 bytes from 10.129.31.59: icmp_seq=1 ttl=64 time=10.2 ms
        64 bytes from 10.129.31.59: icmp_seq=2 ttl=64 time=1.01 ms
      ```
   
      So to force, you need to add a record for datars31-0 into **/etc/hosts** with IP of datars31-0.datars3.mongodb-cluster.svc.cluster.local:
   
      ```text
        10.129.31.59 datars31-0 
      ```
   
      Mongos finds a new config, re-configures itself, and uses the full path later on. There is no need for the **/etc/hosts** record anymore.
   
      ```text
        {  "_id" : "datars2",  "host" : "datars2/datars20-0.datars2.lambda-backup-mongo.svc.cluster.local:27017,datars21-0.datars2.lambda-backup-mongo.svc.cluster.local:27017,datars22-0.datars2.lambda-backup-mongo.svc.cluster.local:27017",  "state" : 1 }
        {  "_id" : "datars3",  "host" : "datars3/datars30-0.datars3.lambda-backup-mongo.svc.cluster.local:27017,datars31-0.datars3.lambda-backup-mongo.svc.cluster.local:27017,datars32-0.datars3.lambda-backup-mongo.svc.cluster.local:27017",  "state" : 1 }
      ```

### Recommendations

"Not applicable"

## Common Issues - No Backup Information in Monitoring

This section provides information about the MongoDB issues related to No Backup Information in Monitoring

### Description

OpenShift service account **monitoring-agent** does not have sufficient rights to view anything in the OpenShift Mongo project.

### Alerts

"Not applicable"

### Stack trace(s)

There are following logs in mongodb-monitoring-agent:

```text
    2018-07-17 13:21:14 ERROR [Engine] [140272126884528] - Error getting series meta for plugin: microservices_info_collector
    Traceback (most recent call last):
      File "/opt/collector_engine.py", line 139, in __updateSeries
        self.__seriesMeta[plugin_name] = plugin_instance.getSeries()
      File "/opt/plugins/microservices_info_collector.py", line 107, in getSeries
        configmap_list = self.__client.get_configmaps()
      File "/opt/utils/openshift_client.py", line 61, in get_configmaps
        resp = resp.get('items', {})
    AttributeError: 'NoneType' object has no attribute 'get'
```

### How to solve

1. Provide the OpenShift service account **monitoring-agent** sufficient rights:
    
    ```
    service_account="monitoring-agent"
    oc policy add-role-to-user view system:serviceaccount:${NAMESPACE}:${service_account} -n ${NAMESPACE}
    ```

### Recommendations

"Not applicable"

## Common Issues - Backup Failed for Some Reason

This section provides information about the MongoDB issues related to Backup Failed for Some Reason.

### Description

"Not applicable"

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **Check following:**
    Reasons can vary. You need to go to ```rc/mongodb-backup-daemon``` or ```dc/mongodb-backup-daemon``` of the failing MongoDB installation and look inside ```/backups/<failed_backup_folder>/.console```.   
    You can get _failed_backup_folder_ from an error, usually it is like `20180717T020000`.
    
    Search for line starting with:
    
    ```
    Backup failed
    ```

    You can find an error near it.

### Recommendations

"Not applicable"

## Common Issues - Mongo Pod Was Down for Some Time and Now Cannot Start

This section provides information about the MongoDB issues related to Mongo Pod Was Down for Some Time and Now Cannot Start.

### Description

You can look in the logs for errors such as:

```replication: Data too stale, halting replication```

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **You need to re-synchronize the pod**
   - For more information, see [Mongodb-statefulset-is-not-starting-with-error-```I - [initandlisten] Assertion: 28595:2: No such file or directory src/mongo/db/storage/wiredtiger/wiredtiger_kv_engine.cpp 267```](#Mongodb Statefulset is not starting with error ```I - [initandlisten] Assertion: 28595:2: No such file or directory src/mongo/db/storage/wiredtiger/wiredtiger_kv_engine.cpp 267```).

### Recommendations

"Not applicable"

## Common Issues - Same Queries Return Different Results

This section provides information about the MongoDB issues related to Same Queries Return Different Results.

### Description

Following are the two possible reasons:
- Split-brain
- One server is too stale and `ReadFromSecondary` is enabled

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. Split brain checking: 

    Go through ReplicaSets and run ```rs.status()``` on each one.   
    If you have something such as the following, you have a split-brain:
    
    ```
    {
            "set" : "datars2",
            ...
            },
            "members" : [
                    {
                            "_id" : 0,
                            "name" : "datars20-0.datars2.eco2-mano91-infra-mongodb.svc.cluster.local:27017",
                            ...
                            "stateStr" : "SECONDARY",
                            ...
                    },
                    {
                            "_id" : 1,
                            "name" : "datars21-0.datars2.eco2-mano91-infra-mongodb.svc.cluster.local:27017",
                            ...
                            "stateStr" : "PRIMARY",
                            ...
                    },
                    {
                            "_id" : 2,
                            "name" : "datars22-0.datars2.eco2-mano91-infra-mongodb.svc.cluster.local:27017",
                            ...
                            "stateStr" : "PRIMARY",
                            ...
                    }
            ],
            "ok" : 1
    }
    ```
    
    You need to sacrifice one of the primary servers and delete all the data from it.

### Recommendations

This potentially can harm your application because there are some documents that are unique for every primary.

## Common Issues - Random Pods Restart During MongoDB Backup Process

This section provides information about the MongoDB issues related to Random Pods Restart During MongoDB Backup Process.

### Description

Backup process may stress the file system. Such cases happen when the backup pod uses GlusterFS and 5GB or more data is dumped at once. This may lead to I/O waits.

![Mongo Network IO during backup](/docs/public/images/backup-troubleshoot-1.png)

![I/O waits during backup](/docs/public/images/backup-troubleshoot-2.png)

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **Limit the incoming (ingress) traffic** 
   - Limit the incoming (ingress) traffic to the backup pod so the load is decreased, but the backup process in this case could be longer.
   To set the traffic limit, specify the `BACKUP_INGRESS_BANDWIDTH` parameter with the desired value in the deploy job parameters.

### Recommendations

"Not applicable"

## Common Issues - Clean deploy failed with error message in operator logs

This section provides information about the MongoDB issues related to Clean deploy failed with error message ```command replSetInitiate requires authentication``` in operator logs.

### Description

PV's which were specified in deploy params have some data from previous installation.

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **Set the following parameters:**
    
    ```
    recycler.install: true 
    ```
    It will clean up PV's before installation. Or clean up PV's manually

### Recommendations

"Not applicable"

## Prometheus Alerts Troubleshooting - MongoDB Replication Lag

This section describes in detail the alerts, the possible reasons, and the solution related to MongoDB Replication Lag.

### Description

|Problem|Possible Reason|
|---|---|
|Replication lag is high.|The secondary nodes cannot replicate data fast enough to keep up with the rate that the data is being written to the primary node.|

This can occur for a few reasons, so it can be hard to pinpoint exactly why you are experiencing a replication lag. Some of the main culprits include network latency, disk throughput, concurrency, and large amounts of data writes to MongoDB. Your MongoDB replication lag could be caused by something as simple as network latency, packet loss within your network, or a routing issue. Any of this could be slowing down the replication from your primary node to your secondary.

One of the leading causes of replication lag in multi-tenant systems is slow disk throughput. If the filesystem on the secondary disk cannot replicate the data to the disks as fast as the primary, the secondary will have issues in keeping up. The disks may also run out of memory, I/O and CPU, keeping the data from being written to secondary node disks and letting them fall further behind the primary.

Concurrent operations can sometimes cause unintended consequences within your system. In this case, large and long running write operations lock up the system and block the replication to secondaries until complete, increasing the replication lag. Similar to concurrency, when running frequent and large write operations, the secondary node disk is unable to read the oplog as fast as the primary is being written to and falls behind on replication.

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **Investigate for possible disk I/O problems**
   - (see [Random Pods Restart](#random-pods-restart) section), network issues of the environment. 

2. **Lack of resources or unoptimized queries or incorrect DB indexes.** 
   - For more information, see [MongoDB CPU Usage](#mongodb-cpu-usage) and [MongoDB Memory Usage](#mongodb-memory-usage).

### Recommendations

"Not applicable"

## Prometheus Alerts Troubleshooting - MongoDB Cursors Timeouts

This section describes in detail the alerts, the possible reasons, and the solution related to MongoDB Cursors Timeouts.

### Description

|Problem|Possible Reason|
|---|---|
|Some cursors exceeded default MongoDB cursor timeout.|Unoptimized queries or large volume of data to gather by one batch.|

By default, the cursor timeout is 10 minutes.

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **Read all the data at one time, and then process it.**
   The disadvantages of this scheme are also obvious. If the data volume is very large, you may not be able to put it all in the memory. Even if it can all be put into the memory, the list derivation traverses all the data, and then the for loop traverses again, wasting time.

2. **Reduce the default amount of pieces of data for cursor**
   Let the cursor return less default amount of pieces of data each time, so that the time to consume this batch of data is less than 10 minutes.
   However, this scheme increases the number of database connections, thus increasing the I/O time consumption.

3. **Remove timeout**
   Let the cursor never time out. Set the parameter “no cursor” timeout = true to make the cursor never timeout.
   However, this operation is very dangerous because if your program stops unexpectedly for some reason, the cursor can no longer be closed. Unless mongodb is restarted, these cursors remain on mongodb and occupy resources.

### Recommendations

"Not applicable"

## Prometheus Alerts Troubleshooting - MongoDB CPU Usage

This section describes in detail the alerts, the possible reasons, and the solution related to MongoDB CPU Usage.

### Description

|Problem|Possible Reason|
|---|---|
|High CPU load and slow queries.|There are incorrect DB indexes or queries are not optimized.|

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **Detect the queries that lead to CPU load spikes:**
    
    Run the following function in the MongoDB console:
    
    ```db.currentOp({“secs_running”: {$gte: 3}})```
    
    It returns a list of operations in progress that take more than 3 seconds to perform. If your CPU load is 100%, you can use — filter.

2. **Inspect the `system.profile` collection**
    If the Database Profiler is configured, you can inspect the `system.profile` collection. For more information, see [https://docs.mongodb.com/manual/tutorial/manage-the-database-profiler/](https://docs.mongodb.com/manual/tutorial/manage-the-database-profiler/) 
    
    For example, the following command sets the profiling level for the current database to `1` and the slow operation threshold to 1000 milliseconds.  
    
    ```user:PRIMARY> db.setProfilingLevel(1, 1000)```
    
    Then, query for the data against this collection and analyze:
    
            db.system.profile.find().pretty()
             
            // or get 'query' operations only and specified fields
            db.system.profile.find( { op: { $eq : 'query' } } , {"millis": 1, "ns": 1, "ts": 1,"query": 1}).sort( { ts : -1 } ).pretty()

3. **Interpret the results and fix the issue.**
    
    You can view the plan summary to identify an inefficient plan or by using `explain(‘executionStats’)`.
    
    Each of the above two ways provides information about the query plan. The `db.currentOp()` method provides the planSummary field — a string that contains the query plan to help debug slow queries. The Database Profiler method provides even more data such as `keysExamined, docsExamined, nreturned, execStats`.
    All these fields provide useful information that contains the execution statistics of the query operation.
    
    For example, using the `db.currentOp()` method, the following query is obtained:
    
    ```text
        "query" : {  
           "$query" : {  
             "UniqueId" : "a6f338db7ea728e0",  
             "application_id" : 36530,  
             "class_name" : "Logs"  
           }  
         },
    ``` 
    Check the query plan using the following:
    
    ```text
        db.custom_data.find({"application_id" : 36530,"class_name" : "Logs","UniqueId" : "a6f338db7ea728e0"}).explain('executionStats')
    ```

    You get the following output:
    
    ```text
        {  
        "queryPlanner" : {  
            "winningPlan" : ...,   
            "rejectedPlans" : [...],  
        },  
        
        "executionStats" : {  
            "executionSuccess" : true,  
            "nReturned" : 1,  
            "executionTimeMillis" : 754,  
            "totalKeysExamined" : 97436,  
            "totalDocsExamined" : 97436,  
            "executionStages" : {  
                "stage" : "FETCH",  
                "filter" : {  
                    "UniqueId" : {  
                        "$eq" : "a6f338db7ea728e0"  
                    }  
                },  
                "nReturned" : 1,  
                "executionTimeMillisEstimate" : 280,  
                "works" : 97438,  
                "advanced" : 1,  
                "needTime" : 97435,  
                "needFetch" : 0,  
                "saveState" : 2283,  
                "restoreState" : 2283,  
                "isEOF" : 1,  
                "invalidates" : 0,  
                "docsExamined" : 97436,  
                "alreadyHasObj" : 0,  
                "inputStage" : {  
                    "stage" : "IXSCAN",  
                    "nReturned" : 97436,  
                    "executionTimeMillisEstimate" : 50,  
                    "works" : 97437,  
                    "advanced" : 97436,  
                    "needTime" : 0,  
                    "needFetch" : 0,  
                    "saveState" : 2283,  
                    "restoreState" : 2283,  
                    "isEOF" : 1,  
                    "invalidates" : 0,  
                    "keyPattern" : {  
                        "application_id" : 1,  
                        "class_name" : 1,  
                        "user_id" : 1,  
                        "created_at" : 1  
                    },  
                    "indexName" : "application_id_1_class_name_1_user_id_1_created_at_1",  
                    "isMultiKey" : false,  
                    "direction" : "forward",  
                    "indexBounds" : {  
                        "application_id" : [  
                            "[36530.0, 36530.0]"  
                        ],  
                        "class_name" : [  
                            "[\"Logs\", \"Logs\"]"  
                        ],  
                        "user_id" : [  
                            "[MinKey, MaxKey]"  
                        ],  
                        "created_at" : [  
                            "[MinKey, MaxKey]"  
                        ]  
                    },  
                    "keysExamined" : 97436,  
                    "dupsTested" : 0,  
                    "dupsDropped" : 0,  
                    "seenInvalidated" : 0,  
                    "matchTested" : 0  
                }  
            }     
        }
    ```
    
    The `docsExamined` key provides information about the number of writes to the hard disk and not the RAM. In the example above, it examined almost 100K(!) documents. Ideally, the 'docsExamined' value must be close to zero.

### Recommendations

"Not applicable"

## Prometheus Alerts Troubleshooting - MongoDB Memory Usage

This section describes in detail the alerts, the possible reasons, and the solution related to MongoDB Memory Usage.

### Description

|Problem|Possible Reason|
|---|---|
|Memory Usage is too high|Too much data in database. Not optimal index usage.|

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **It is possible that big amounts of data loads the mongo memory. Execute the following:**

    ```text
        mongosh admin -u {mongo_user} -p {mongo_password}
        db.stats()
        ```
        
        It provides output like this:
        ```
        {
        "db" : "admin",
        "collections" : 3,
        "views" : 0,
        "objects" : 9,
        "avgObjSize" : 315,
        "dataSize" : 2835,
        "storageSize" : 81920,
        "numExtents" : 0,
        "indexes" : 4,
        "indexSize" : 114688,
        "fsUsedSize" : 1394253824,
        "fsTotalSize" : 2046640128,
        "ok" : 1,
        "operationTime" : Timestamp(1560517541, 1),
        "$gleStats" : {
            "lastOpTime" : Timestamp(0, 0),
            "electionId" : ObjectId("7fffffff0000000000000001")
        },
        "lastCommittedOpTime" : Timestamp(1560517541, 1),
        "$clusterTime" : {
            "clusterTime" : Timestamp(1560517541, 1),
            "signature" : {
                "hash" : BinData(0,"sgndM9T3OHJMkgb8S8S4yqYPJn4="),
                "keyId" : NumberLong("6702361869169983503")
            }
        }
        }
    ```
    
    View the `indexSize` and `dataSize` provided in bytes. Up to 50% of this size can be cached in RAM. If that is the problem, consider increasing the RAM limits on `datars` StatefulSets or review the index policy.

### Recommendations

"Not applicable"

## Prometheus Alerts Troubleshooting - MongoDB Replication Status 3

This section describes in detail the alerts, the possible reasons, and the solution related to MongoDB Replication Status 3 (RECOVERING).

### Description

|Problem|Possible Reason|
|---|---|
|Members either perform startup self-checks, or transition from completing a rollback or resync.|Various reasons.|

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **Check the pod logs and find error messages to continue investigation of the root cause.**

### Recommendations

"Not applicable"

## Prometheus Alerts Troubleshooting - MongoDB Replication Status 6

This section describes the alerts, the possible reasons, and the solution related to MongoDB Replication Status 6 (UNKNOWN) in detail.

### Description

|Problem|Possible Reason|
|---|---|
|The member's state, as seen from another member of the set, is not yet known.|Various reasons.|


### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **Check pod logs and find error messages to continue investigation of the root cause.**

### Recommendations

"Not applicable"

## Prometheus Alerts Troubleshooting - MongoDB Replication Status 8

This section describes in detail the alerts, the possible reasons, and the solution related to MongoDB Replication Status 8 (DOWN).

### Description

|Problem|Possible Reason|
|---|---|
|The member, as seen from another member of the set, is unreachable. The member, as seen from another member of the set, is unreachable.|Network issues.|

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **Check pod logs and find error messages to continue investigation of the root cause.**

### Recommendations

"Not applicable"

## Prometheus Alerts Troubleshooting - MongoDB Replication Status 9

This section describes in detail the alerts, the possible reasons, and the solution related to  MongoDB Replication Status 9 (ROLLBACK).

### Description

|Problem|Possible Reason|
|---|---|
|This member is actively performing a rollback.|-|

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

"Not applicable"

### Recommendations

"Not applicable"

## Prometheus Alerts Troubleshooting - MongoDB Replication Status 10

This section describes in detail the alerts, the possible reasons, and the solution related to MongoDB Replication Status 10 (REMOVED).

### Description

|Problem|Possible Reason|
|---|---|
|This member was once in a replica set, but was subsequently removed.|-|

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

"Not applicable"

### Recommendations

"Not applicable"
