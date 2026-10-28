---
name: troubleshoot-mongodb
description: Diagnose and resolve MongoDB operator issues including sharded cluster failures, replica-set replication problems, backup/restore failures, DBaaS adapter backup-registration failures, deployment errors, DR switchover issues, WiredTiger corruption and cache pressure, SCRAM/keyfile authentication errors, connection limits, and resource exhaustion. Covers MongodbDeployment and MongodbSupplService CRDs, mongod/mongos pods, backup daemon, monitoring exporter, and DBaaS adapter integration. Use for qubership-mongodb-operator, MongoDB-as-a-Service on Kubernetes/OpenShift.
---

# MongoDB Operator Troubleshooting Skill

Use this skill when investigating issues related to:

- **qubership-mongodb-operator** (MongoDB operator managing sharded/HA MongoDB clusters)
- **MongodbDeployment** CRD failures (cluster provisioning, replica-set configuration, storage, rolling updates)
- **MongodbSupplService** CRD issues (backup daemon, monitoring, DBaaS adapter)
- **MongoDB replication** failures (RECOVERING, UNKNOWN, DOWN, ROLLBACK, REMOVED states, replication lag)
- **WiredTiger** issues (corruption assertions, `WT_CACHE_FULL`, disk space)
- **Backup/restore** failures (backup daemon PVC, S3 credentials, `.console` logs)
- **Deployment errors** (Helm/CRD validation, PVC binding, leftover resources)
- **Disaster Recovery** (active-standby switchover/failover, clean-install ordering, domain name swap)
- **Authentication** (root/monitoring users, SCRAM-SHA failures, keyfile rotation)
- **Connection problems** (`Too many open files`, mongos routing errors)
- **DBaaS adapter** integration (backup registration, bulk operations, missing metrics)

## Component Overview

### Operator Architecture

**qubership-mongodb-operator** is a Kubernetes operator providing MongoDB-as-a-Service:

- **`MongodbDeployment`** CRD — managed by the core operator (`mongodb-operator` chart). Defines the MongoDB cluster topology, resources, storage, DR mode, TLS, and WiredTiger cache.
- **`MongodbSupplService`** CRD — managed by the services operator (`mongodb-services` chart). Defines the backup daemon, monitoring exporter, and DBaaS adapter.

**Key Components:**
- **mongod** — MongoDB database process running in `datarsN*` (data shards) and `cnfrs*` (config servers) StatefulSet pods.
- **mongos** — Query router for sharded clusters.
- **backup daemon** — Performs logical backups (`mongodb-backup-daemon`).
- **monitoring exporter** — Prometheus metrics exporter for MongoDB.
- **DBaaS adapter** — Database lifecycle management API.
- **recycler** — Optional init job that cleans PVs before a clean install.

### Critical Configuration Paths

**Helm Chart Locations:**
- `operator/charts/helm/mongodb-operator/` — Core operator Helm chart (`MongodbDeployment` CR)
- `services/service/charts/helm/mongodb-services/` — Services layer Helm chart (`MongodbSupplService` CR)

**Key Helm Values / CR Fields:**
- `schemaSettings.sharded` / `schemaSettings.schemaType` — Topology (`ha`, sharded count, replica sizes)
- `schemaSettings.thisDomainName` / `schemaSettings.otherDomainName` — DR cluster domain names (must be swapped between sides)
- `disasterRecovery.mode` — `active`, `standby`, or `disable`
- `mongodb.customDataRSParameters` — Custom `key=value` mongod parameters (JSON-encoded string)
- `mongodb.dataWiredTigerCacheGb` / `mongodb.cnfWiredTigerCacheGb` / `mongodb.singleWiredTigerCacheGb` — WiredTiger cache sizes
- `mongodb.dataResources.limits.*` / `mongodb.cnfResources.limits.*` / `mongodb.mongosResources.limits.*` — Resource limits
- `mongodb.storage.waitPvcBound` — Wait for MongoDB data PVC binding
- `backup.storage.waitPvcBound` — Wait for backup daemon PVC binding
- `recycler.install` — Clean PVs before a fresh install
- `tls.mode` — TLS mode (`disabled`, `allowTLS`, `preferTLS`, `requireTLS`)
- `prometheusExporter.alerts.common.*` — Alert thresholds

### Forbidden Actions

**Never** recommend these without explicit user confirmation:

- `rs.reconfig(...)` on an operator-managed replica set — the operator overwrites it on the next reconcile.
- `db.adminCommand({setParameter: ...})` for a persistent setting — lost on pod restart; use `mongodb.customDataRSParameters` instead.
- `rm -rf /data/...` or `rm -f /data/*` to clear a pod's data unless at least one member of the same replica set is `PRIMARY` or `SECONDARY` at that moment.
- `kubectl delete pvc` for MongoDB data volumes — risks permanent data loss; use operator-managed backup/restore or recycler instead.
- Manually editing the `last-applied-configuration-info` ConfigMap outside the documented switchover retry steps.

## Common Failure Modes

### 1. Deployment Failures

#### CRD Validation Failures

**Symptom:** `spec.mongodb.customDataRSParameters: Invalid value: "null"` or `spec.schemaSettings.otherDomainName: Invalid value: "null"`

**Cause:** Helm values left empty/null in an older release.

**Resolution:** Set valid defaults in the `mongodb-operator` Helm chart values (`MongodbDeployment` CR):

```yaml
schemaSettings:
  otherDomainName: "cluster.local"
mongodb:
  customDataRSParameters: '["logLevel=0"]'
```

See the detailed section *Common Issues - Deploy Fails* in `references/troubleshooting.md`.

#### PVC Binding Timeout

**Symptom:** Operator log shows `Reconciliation exception: PVC pvc-mongodb-main-data-0 'Bound' status waiting failed`

**Cause:** Storage provisioner creates/binds PV only when the first consumer is created.

**Resolution:** Disable waiting for PVC binding:
- For MongoDB data PVCs: set `mongodb.storage.waitPvcBound: false`
- For backup daemon PVC: set `backup.storage.waitPvcBound: false`

See the detailed section *Common Issues - Deploy Failed With Error in Operator Logs* in `references/troubleshooting.md`.

### 2. Replica Set / Replication Issues

#### Replica Stuck in Recovering or Unhealthy States

**Symptom:** Prometheus alert `MongoDB replication Status 3 (RECOVERING)`, `MongoDBRecoveringTooLong`, `UNKNOWN`, `DOWN`, `ROLLBACK`, or `REMOVED`; or log lines like `replication: Data too stale, halting replication`.

**Diagnostic Steps:**

1. Check the affected member's mongod log:
   ```bash
   kubectl logs -n <namespace> <pod-name> -c mongod | grep -iE "replication|recover|rollback|stale"
   ```
2. Inspect replica-set state:
   ```bash
   kubectl exec -n <namespace> <pod-name> -c mongod -- \
     mongosh admin -u root -p <root_pass> --eval='rs.status().members.forEach(m => {print(m.name + "\t" + m.stateStr)})'
   ```
3. Verify at least one member of the same replica set is `PRIMARY` or `SECONDARY` before any destructive recovery step.

**Common Causes:**
- Member was down too long and oplog no longer covers the outage → resync required.
- WiredTiger corruption (`Assertion: 28595:2: No such file or directory src/mongo/db/storage/wiredtiger/wiredtiger_kv_engine.cpp 267`).
- Network partition / split-brain (two primaries in one replica set).

**Resolution:**
- Data too stale → clear the member's data and restart the pod to resync.
- WiredTiger assertion → scale down the affected StatefulSet, clear the data directory, scale back up.
- Split-brain → sacrifice one primary by clearing its data and resyncing.

See the detailed sections in `references/troubleshooting.md`:
- *Common Issues - Mongo Pod Was Down for Some Time and Now Cannot Start*
- *Common Issues - Mongodb Statefulset is Not Starting With Error*
- *Common Issues - Same Queries Return Different Results*
- *Common Issues - Could not find host matching read preference for set datarsX*

#### Replication Lag

**Symptom:** Alert `MongoDB replication lag` fires.

**Common Causes:**
- Disk IOPS saturation on secondaries.
- Network latency/packet loss.
- Large or long-running write operations blocking replication.
- Insufficient CPU/memory limits.

**Resolution:**
- Investigate disk I/O and backup-related I/O waits.
- Review CPU/memory usage and slow queries.

See the detailed sections in `references/troubleshooting.md`:
- *Prometheus Alerts Troubleshooting - MongoDB Replication Lag*
- *Common Issues - Random Pods Restart During MongoDB Backup Process*
- *Prometheus Alerts Troubleshooting - MongoDB CPU Usage*
- *Prometheus Alerts Troubleshooting - MongoDB Memory Usage*

### 3. Storage / Disk Issues

#### No Free Space Left on MongoDB PV

**Symptom:** All MongoDB pods start rebooting; disk usage at 100%.

**Resolution:**
1. Scale down one member of each affected replica set.
2. Free space on the PV (remove old data or unnecessary collections).
3. Run `compact` on collections.
4. Scale the pods back up.

See the detailed sections in `references/troubleshooting.md`:
- *Common Issues - No Free Space Left on MongoDB PV*
- *Common Issues - MongoDB Disk Space Usage Not Changed After Data Deleted*

#### Disk Space Not Reclaimed After Delete

**Symptom:** Data deleted but PV usage unchanged.

**Cause:** MongoDB does not release disk space to the filesystem until `compact` runs.

**Resolution:** Run `db.runCommand({compact: "<collection_name>"})` on each replica-set member, stepping down the primary first.

See the detailed section *Common Issues - MongoDB Disk Space Usage Not Changed After Data Deleted* in `references/troubleshooting.md`.

### 4. Performance / Resource Issues

#### WiredTiger Cache Pressure (`WT_CACHE_FULL`)

**Symptom:** Read latency spikes, `WT_CACHE_FULL` in mongod log, or alert `MongoWiredTigerCachePressure`.

**Cause:** WiredTiger cache is undersized for the working set.

**Resolution:**
1. Raise the typed cache value for the affected component in the `mongodb-operator` chart values:
   - `mongodb.dataWiredTigerCacheGb` — data shards
   - `mongodb.cnfWiredTigerCacheGb` — config servers
   - `mongodb.singleWiredTigerCacheGb` — standalone deployments
2. Raise the matching pod memory limit (`mongodb.dataResources.limits.memory`, `mongodb.cnfResources.limits.memory`) so the cache fits in heap.

See the detailed section *Common Issues - MongoDB Read Latency Spikes / WT_CACHE_FULL (Not Corruption)* in `references/troubleshooting.md`.

#### High CPU / Memory Usage

**Symptom:** Alerts `MongoDB CPU Usage` or `MongoDB Memory Usage`.

**Resolution:**
- Find slow operations with `db.currentOp({"secs_running": {$gte: 3}})`.
- Inspect `system.profile` for slow queries.
- Optimize indexes/queries.
- Raise CPU or memory limits if the cluster is genuinely under-provisioned.

See the detailed sections in `references/troubleshooting.md`:
- *Prometheus Alerts Troubleshooting - MongoDB CPU Usage*
- *Prometheus Alerts Troubleshooting - MongoDB Memory Usage*

#### Too Many Open Files

**Symptom:** `Error accepting new connection on local endpoint` or `Too many open files` in mongod log.

**Cause:** Container's `LimitNOFILE` is too low.

**Resolution:** Raise `LimitNOFILE` in the containerd systemd drop-in on affected worker nodes to at least `64000`, restart containerd, then restart MongoDB pods.

See the detailed section *Common Issues - Connections Refused With "Too Many Open Files" Error* in `references/troubleshooting.md`.

### 5. Backup Issues

#### Backup Job Failed

**Symptom:** Alert `MongoDB last backup failed`.

**Diagnostic Steps:**
1. Locate the failed backup folder from the alert/error (e.g., `20180717T020000`).
2. Read the backup daemon console log:
   ```bash
   kubectl exec -n <namespace> <backup-daemon-pod> -- \
     cat /backups/<failed_backup_folder>/.console
   ```
3. Search for the line starting with `Backup failed`.

**Common Causes:**
- Insufficient backup daemon PVC space/inodes.
- S3/credentials misconfiguration.
- I/O saturation during large backups.

**Resolution:**
- Increase backup PVC size or limit ingress bandwidth with `BACKUP_INGRESS_BANDWIDTH` deploy parameter.
- Verify backup credentials and storage settings.

See the detailed sections in `references/troubleshooting.md`:
- *Common Issues - Backup Failed for Some Reason*
- *Common Issues - Random Pods Restart During MongoDB Backup Process*

#### No Backup Information in Monitoring

**Symptom:** Grafana shows no backup metrics; `mongodb-monitoring-agent` logs `AttributeError: 'NoneType' object has no attribute 'get'`.

**Cause:** `monitoring-agent` service account lacks `view` role in the MongoDB namespace.

**Resolution:**
```bash
oc policy add-role-to-user view system:serviceaccount:<namespace>:monitoring-agent -n <namespace>
```

See the detailed section *Common Issues - No Backup Information in Monitoring* in `references/troubleshooting.md`.

### 6. DR Issues

#### Switchover / Failover Failed

**Symptom:** DR switchover/failover operation does not complete or status stays unexpected.

**Resolution:**
1. On the failing side, edit the `last-applied-configuration-info` ConfigMap and delete the `summary-spec` value.
2. Restart the operator pod on that side.
3. Retry the switchover/failover.

See the detailed section *DR Issues - Swithover/Failover failed* in `references/troubleshooting.md`.

#### Clean Install Fails on Only One Side of a DR Pair

**Symptom:** One side of an active/standby pair installs successfully, the other fails.

**Common Causes:**
- Standby was not deployed first.
- Domain names were not swapped between sides.

**Resolution:**
1. Deploy the **standby** side first (`disasterRecovery.mode: standby`), wait for success.
2. Then deploy the **active** side (`disasterRecovery.mode: active`).
3. Ensure `schemaSettings.thisDomainName` and `schemaSettings.otherDomainName` are swapped between the two sides.

See the detailed section *DR Issues - Clean Install Fails on Only One Side of a DR Pair* in `references/troubleshooting.md`.

### 7. Authentication & Connectivity

#### Cannot Log In to MongoDB

**Symptom:** `UserNotFound: Could not find user root@admin` or mongos log `can't choose primary for datars<N>`.

**Cause:** Replica set was not initiated; root/monitoring users were not created.

**Resolution:**
1. Initialize the replica set with `rs.initiate(...)`.
2. Create the root user and monitoring user on the primary.
3. Restart mongos to pick up the new users.

See the detailed section *Common Issues - Cannot Login to MongoDB* in `references/troubleshooting.md`.

#### Intra-Cluster Authentication Fails During Rolling Update

**Symptom:** SCRAM-SHA-256 / `SASL step` failures for user `__system` between members during a rolling update.

**Cause:** Keyfile rotation — pods on the new `mongo-secret` cannot authenticate to pods still on the old keyfile.

**Resolution:**
1. Verify keyfile consistency across all replica-set pods.
2. Restart any mongod pods that have not picked up the new keyfile.
3. Restart mongos routers to refresh cached auth state.

See the detailed section *Common Issues - Intra-Cluster Authentication Fails During Rolling Update (Keyfile Rotation)* in `references/troubleshooting.md`.

### 8. Pod Startup Issues

#### Pods Do Not Start With `context deadline exceeded`

**Symptom:** Pod Events show `context deadline exceeded`.

**Cause:** Docker/container runtime issue on the node.

**Resolution:** Restart Docker/containerd on the affected worker node (operations/IT task).

See the detailed section *Common Issues - Pods Do Not Start With Error in Events* in `references/troubleshooting.md`.

#### Pods Stuck in Pending State

**Symptom:** Pod stays in `Pending` on older installations.

**Cause:** Exited init-container removed from the node by a cleanup cronjob.

**Resolution:** Delete the pending pods one-by-one, waiting for each to reach `Running 1/1` before deleting the next. Update to the newest operator version for a permanent fix.

See the detailed section *Common Issues - Pods Stuck in Pending State* in `references/troubleshooting.md`.

#### Index Build Stuck

**Symptom:** Log line `Index build: waiting for next action before completing final phase`.

**Cause:** One replica is not in `PRIMARY`/`SECONDARY` state and the commit quorum cannot be reached.

**Resolution:** Fix the unhealthy replica (usually by clearing its data and resyncing), then retry the index build.

See the detailed section *Common Issues - Index build stuck* in `references/troubleshooting.md`.

### 9. DBaaS Adapter / Backup Registration Issues

#### Adapter Reports backupRestore: false

**Symptom:** DBaaS backup request fails with "some adapters do not supported backup operation"; registered physical database shows `"backupRestore": false`; dbaas-mongo-adapter log shows `Backup credentials are not set. Backup API is not working.`; secret `mongodb-backup-api-credentials.v1` exists but the adapter Deployment has no backup credential env vars.

**Cause:** Enabling backup or upgrading the operator without changing a `spec.dbaas` field may not trigger reconciliation of the dbaas-mongo-adapter Deployment, so the `BACKUP_DAEMON_API_CREDENTIALS_USERNAME` / `PASSWORD` env vars are never added.

**Resolution:**
1. **Recommended workaround:** Touch any `spec.dbaas` field (for example, `dbaasResources.limits.cpu`) and redeploy to force the operator to recreate the adapter Deployment with the backup credentials.
2. **Direct patch:** Patch the running `dbaas-mongo-adapter` Deployment to add env vars from `mongodb-backup-api-credentials.v1`.
3. **Verify:** Confirm the Deployment has the env vars, re-register if needed, and retry the backup request.

See the detailed section *Common Issues - DBaaS Adapter Does Not Support Backup Operation* in `references/troubleshooting.md`.

## Diagnostic Workflow

### Initial Triage

1. **Identify the failing component:**
   - Operator logs: `kubectl logs -n <namespace> deployment/mongodb-operator`
   - Mongod logs: `kubectl logs -n <namespace> <datarsN-pod> -c mongod`
   - Mongos logs: `kubectl logs -n <namespace> <mongos-pod>`
   - Backup daemon logs: `kubectl logs -n <namespace> <mongodb-backup-daemon-pod>`
   - Monitoring agent logs: `kubectl logs -n <namespace> <mongodb-monitoring-agent-pod>`

2. **Check resource state:**
   - CRD status: `kubectl get mongodbdeployment,mongodbsupplservice -n <namespace> -o yaml`
   - Pod status: `kubectl get pods -n <namespace> -o wide`
   - PVC status: `kubectl get pvc -n <namespace>`
   - Replica-set state: `kubectl exec <pod> -c mongod -- mongosh admin -u root -p <pass> --eval='rs.status()'`

3. **Review events:**
   ```bash
   kubectl get events -n <namespace> --sort-by='.lastTimestamp'
   ```

### Narrowing Down Issues

Use the symptom-to-section mapping below to jump to the relevant high-level failure mode above, then follow the reference into `references/troubleshooting.md` for step-by-step remediation.

| Symptom / Alert | Failure Mode / Detailed Section |
|---|---|
| `spec.mongodb.customDataRSParameters: Invalid value: "null"` | [1. Deployment Failures - CRD Validation Failures](#1-deployment-failures) |
| `Reconciliation exception: PVC ... 'Bound' status waiting failed` | [1. Deployment Failures - PVC Binding Timeout](#1-deployment-failures) |
| `Assertion: 28595:2: No such file or directory src/mongo/db/storage/wiredtiger/wiredtiger_kv_engine.cpp 267` | [2. Replica Set / Replication Issues](#2-replica-set--replication-issues) |
| Every pod of a replica set down after WiredTiger assertion | [2. Replica Set / Replication Issues](#2-replica-set--replication-issues) |
| `replication: Data too stale, halting replication` | [2. Replica Set / Replication Issues](#2-replica-set--replication-issues) |
| `Could not find host matching read preference { mode: "primary" } for set datarsX` | [2. Replica Set / Replication Issues](#2-replica-set--replication-issues) |
| `UserNotFound: Could not find user root@admin` / `can't choose primary for datars<N>` | [7. Authentication & Connectivity - Cannot Log In](#7-authentication--connectivity) |
| SCRAM-SHA-256 / `SASL step` failures for `__system` during rolling update | [7. Authentication & Connectivity - Intra-Cluster Authentication Fails](#7-authentication--connectivity) |
| `Error accepting new connection on local endpoint` / `Too many open files` | [4. Performance / Resource Issues - Too Many Open Files](#4-performance--resource-issues) |
| `WT_CACHE_FULL` / `MongoWiredTigerCachePressure` alert | [4. Performance / Resource Issues - WiredTiger Cache Pressure](#4-performance--resource-issues) |
| `MongoDB last backup failed` alert | [5. Backup Issues - Backup Job Failed](#5-backup-issues) |
| No backup metrics; `AttributeError: 'NoneType' object has no attribute 'get'` | [5. Backup Issues - No Backup Information in Monitoring](#5-backup-issues) |
| DBaaS backup request: "some adapters do not supported backup operation" | [9. DBaaS Adapter / Backup Registration Issues - Adapter Reports backupRestore: false](#9-dbaas-adapter--backup-registration-issues) |
| Physical DB `"backupRestore": false` | [9. DBaaS Adapter / Backup Registration Issues - Adapter Reports backupRestore: false](#9-dbaas-adapter--backup-registration-issues) |
| Adapter log: `Backup credentials are not set. Backup API is not working.` | [9. DBaaS Adapter / Backup Registration Issues - Adapter Reports backupRestore: false](#9-dbaas-adapter--backup-registration-issues) |
| `MongoDB replication lag` alert | [2. Replica Set / Replication Issues - Replication Lag](#2-replica-set--replication-issues) |
| `MongoDB CPU Usage` alert | [4. Performance / Resource Issues - High CPU / Memory Usage](#4-performance--resource-issues) |
| `MongoDB Memory Usage` alert | [4. Performance / Resource Issues - High CPU / Memory Usage](#4-performance--resource-issues) |
| `MongoDB cursors timeouts` alert | *Prometheus Alerts Troubleshooting - MongoDB Cursors Timeouts* in `references/troubleshooting.md` |
| Replica `stateStr` = `RECOVERING` | [2. Replica Set / Replication Issues](#2-replica-set--replication-issues) |
| Replica `stateStr` = `UNKNOWN` | [2. Replica Set / Replication Issues](#2-replica-set--replication-issues) |
| Replica `stateStr` = `DOWN` | [2. Replica Set / Replication Issues](#2-replica-set--replication-issues) |
| Replica `stateStr` = `ROLLBACK` | [2. Replica Set / Replication Issues](#2-replica-set--replication-issues) |
| Replica `stateStr` = `REMOVED` | [2. Replica Set / Replication Issues](#2-replica-set--replication-issues) |
| Two primaries in one replica set | [2. Replica Set / Replication Issues](#2-replica-set--replication-issues) |
| Pods restart during backup window | [5. Backup Issues - Backup Job Failed](#5-backup-issues) |
| Primary flips repeatedly | [4. Performance / Resource Issues - High CPU / Memory Usage](#4-performance--resource-issues) |
| DR switchover/failover fails | [6. DR Issues - Switchover / Failover Failed](#6-dr-issues) |
| DR clean install fails on one side only | [6. DR Issues - Clean Install Fails on Only One Side](#6-dr-issues) |
| `context deadline exceeded` in Events | [8. Pod Startup Issues - Pods Do Not Start](#8-pod-startup-issues) |
| Pod phase `Pending` (older installations) | [8. Pod Startup Issues - Pods Stuck in Pending](#8-pod-startup-issues) |
| `Index build: waiting for next action before completing final phase` | [8. Pod Startup Issues - Index Build Stuck](#8-pod-startup-issues) |

## Using the Full Troubleshooting Guide

For step-by-step remediation, load only the relevant section from `references/troubleshooting.md` (or `docs/public/troubleshooting.md`) — **do NOT read the full file**.

**Steps:**

1. Grep section headers to find line numbers:
   ```bash
   grep -n "^## " references/troubleshooting.md
   ```
2. Match the symptom/alert to a section header using the mapping above.
3. Read only that section using `offset` and `limit`:
   - `offset` = section start line
   - `limit` = lines until the next `## ` header

**Example:**
If grep shows:
```
30:## Common Issues - No Free Space Left on MongoDB PV
60:## Common Issues - MongoDB Disk Space Usage Not Changed After Data Deleted
```

To read the "No Free Space" section:
```
Read(file_path="references/troubleshooting.md", offset=30, limit=30)
```

## Key Helm Chart Parameters

### Critical Configuration

**Topology:**
```yaml
schemaSettings:
  sharded: true
  schemaType: "ha"
  cnfReplicaSize: 3
  dataReplicaSize: 3
  mongosReplicas: 1
  shardCount: 3
  thisDomainName: "cluster.local"
  # otherDomainName: "peer.cluster.local"  # required for DR; swap with peer side
```

**Disaster Recovery:**
```yaml
disasterRecovery:
  mode: "active"  # active | standby | disable
  siteManagerEnabled: true
```

**Custom mongod parameters:**
```yaml
mongodb:
  customDataRSParameters: '["logLevel=0", "diagnosticDataCollectionEnabled=false"]'
```

**WiredTiger cache and resources:**
```yaml
mongodb:
  dataWiredTigerCacheGb: "1.0"   # data shards
  cnfWiredTigerCacheGb: "0.5"    # config servers
  singleWiredTigerCacheGb: "0.25"  # standalone deployments
  dataResources:
    limits:
      memory: 1Gi
      cpu: 500m
  cnfResources:
    limits:
      memory: 1Gi
      cpu: 500m
  mongosResources:
    limits:
      memory: 512Mi
      cpu: 500m
```

**Storage:**
```yaml
mongodb:
  storage:
    waitPvcBound: false
    size:
      - "5Gi"
```

**Backup PVC wait:**
```yaml
backup:
  storage:
    waitPvcBound: false
```

**Recycler for clean installs:**
```yaml
recycler:
  install: false
```

**TLS:**
```yaml
tls:
  mode: disabled  # disabled | allowTLS | preferTLS | requireTLS
```

**Alert thresholds (`mongodb-services` chart):**
```yaml
prometheusExporter:
  alerts:
    common:
      replicationLagSeconds: 10
      cursorTimeoutThreshold: 100
      cpuUsageThreshold: 95
      memoryUsageThreshold: 95
      cachePressureThreshold: 85
```

## Log Analysis Tips

### Operator Logs

**Key patterns:**
```bash
# Reconciliation errors
kubectl logs deployment/mongodb-operator | grep -iE "error|failed|unable|exception"

# CRD updates
kubectl logs deployment/mongodb-operator | grep -iE "MongodbDeployment|MongodbSupplService"

# DR / switchover
kubectl logs deployment/mongodb-operator | grep -iE "switchover|failover|disasterRecovery"
```

### Mongod Logs

**Key patterns:**
```bash
# Replication issues
kubectl logs <datarsN-pod> -c mongod | grep -iE "replication|recover|rollback|stale|election"

# WiredTiger / storage
kubectl logs <datarsN-pod> -c mongod | grep -iE "wiredtiger|wt_cache_full|assertion|corrupt"

# Authentication
kubectl logs <datarsN-pod> -c mongod | grep -iE "scr|sasl|authentication|userNotFound"

# Startup / shutdown / crash
kubectl logs <datarsN-pod> -c mongod | grep -iE "starting|shutdown|crash|fatal"
```

### Mongos Logs

**Key patterns:**
```bash
# Routing errors
kubectl logs <mongos-pod> | grep -iE "can't choose primary|read preference|config server"

# Connection limits
kubectl logs <mongos-pod> | grep -iE "too many open files|connection refused"
```

### Backup Daemon Logs

**Key patterns:**
```bash
# Backup operations
kubectl logs <mongodb-backup-daemon-pod> | grep -iE "backup|restore|snapshot"

# Storage / credentials
kubectl logs <mongodb-backup-daemon-pod> | grep -iE "s3|credentials|permission denied|disk full"
```

## Integration Points

### Monitoring Exporter

- Exposes MongoDB metrics for Prometheus via `ServiceMonitor`.
- Alert thresholds are configured in `mongodb-services` chart values (`prometheusExporter.alerts.common.*`).
- Health check:
  ```bash
  kubectl port-forward -n <namespace> <mongodb-monitoring-agent-pod> 9216:9216
  curl http://localhost:9216/metrics
  ```

### DBaaS Adapter

- Provides MongoDB lifecycle API for DBaaS Aggregator.
- Logs are in the `dbaas-mongo-adapter` container.
- Common symptom: bulk-drop timeouts caused by prepared/frozen transactions (MongoDB equivalent: long-running or stuck DDL).

### Backup Storage

- Backup daemon writes to `/backups/<timestamp>` inside its PVC or to configured S3-compatible storage.
- Verify credentials in the `mongodb-backup-credentials` Secret.
- Check backup daemon PVC space and inode usage when backups fail or alerts fire.

## Prevention Best Practices

1. **Always set `mongodb.customDataRSParameters` correctly** (JSON-encoded `"key=value"` list); redeploy replaces the whole list, so include existing entries when adding new ones.
2. **Use the typed `mongodb.*WiredTigerCacheGb` values** instead of injecting cache settings via `customDataRSParameters`.
3. **Raise cache and memory limits together** — cache must stay below the pod memory limit to avoid OOM kills.
4. **Always deploy the DR standby side first**, then the active side, with swapped `thisDomainName`/`otherDomainName` values.
5. **Use `recycler.install: true` for clean reinstalls** on reused PVs, or wipe PVs manually.
6. **Monitor disk usage and oplog window** to avoid "data too stale" resyncs.
7. **Set appropriate PVC storage class and size**; expand before the cluster runs out of space.
8. **Test backup/restore regularly** in non-production.
9. **Enable TLS in production** (`tls.mode: preferTLS` or `requireTLS`).
10. **Review slow queries and indexes** before raising CPU/memory limits.
11. **Keep containerd `LimitNOFILE` at 64000 or higher** on MongoDB worker nodes.
12. **Regularly test failover and DR switchover procedures**.

## Related Documentation

- [Installation Guide](docs/public/installation_guide.md) — Prerequisites, Helm parameters, platform-specific setup
- [Configuration Guide](docs/public/configuration.md) — CRD fields and value reference
- [Architecture](docs/public/architecture.md) — Cluster topology and component architecture
- [Operations Guide](docs/public/operations_guide.md) — Day-2 operations
- [Troubleshooting Guide](docs/public/troubleshooting.md) — Detailed issue-by-issue remediation
- [DR Operations](docs/public/dr_operations.md) — Active-standby DR procedures
- [Backup Daemon](docs/public/backup-daemon.md) — Backup/restore configuration
- [Grafana Dashboard](docs/public/grafana_dashboard.md) — Monitoring dashboards

## When to Escalate

Escalate to component maintainers when:

- Operator crashes with a Go panic (attach stack trace).
- Data corruption suspected (`Assertion: 28595` across multiple members, checksum errors, unrecoverable WiredTiger files).
- Split-brain with multiple primaries and data divergence.
- All members of a replica set are down and no backup is available.
- Backup restore fails repeatedly with cryptic or storage-side errors.
- DR switchover/failover cannot be recovered by the documented retry steps.
- Upgrade or rolling update leaves the cluster in an inconsistent state.
- Performance degradation unexplained by query patterns, resource limits, or load.

Include in escalation:
- Full operator logs (`kubectl logs deployment/mongodb-operator`)
- Mongod logs from all pods in the affected replica set(s)
- CRD YAML (`kubectl get mongodbdeployment,mongodbsupplservice -o yaml`)
- Replica-set status (`rs.status()` output from each set)
- PVC and node events (`kubectl get events -n <namespace>`)
- Backup daemon `.console` log for the failing backup folder
