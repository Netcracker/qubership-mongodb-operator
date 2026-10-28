---
name: troubleshoot-mongodb
description: "Diagnose and resolve MongoDB operator issues: sharded/replica-set failures, backup/restore, DBaaS adapter, deployment, DR switchover/HA-to-DR, WiredTiger, SCRAM/keyfile auth, connection limits, resource exhaustion. Covers MongodbDeployment/MongodbSupplService CRDs, mongod/mongos, backup daemon, monitoring/prometheus exporter, DBaaS adapter."
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
- **Disaster Recovery** (active-standby switchover/failover, clean-install ordering, HA-to-DR live migration, domain name swap, prometheus-exporter EXPORT_MONGOS misconfiguration after switchover)
- **Authentication** (root/monitoring users, SCRAM-SHA failures, keyfile rotation)
- **Connection problems** (`Too many open files`, mongos routing errors)
- **DBaaS adapter** integration (backup registration, bulk operations, missing metrics)

## Component Overview

**qubership-mongodb-operator** manages MongoDB-as-a-Service via two CRDs:
- **`MongodbDeployment`** (`mongodb-operator` chart) — topology, resources, storage, DR, TLS, WiredTiger cache.
- **`MongodbSupplService`** (`mongodb-services` chart) — backup daemon, monitoring exporter, DBaaS adapter.

**Key pods/containers:** `datarsN*` / `cnfrsN*` mongods, `mongos`, `mongodb-backup-daemon`, `mongodb-monitoring-agent`, `mongodb-prometheus-exporter`, `dbaas-mongo-adapter`, optional recycler init job.

**Key fields:** `schemaSettings.schemaType`/`thisDomainName`/`otherDomainName`; `disasterRecovery.mode`; `mongodb.customDataRSParameters`; `mongodb.*WiredTigerCacheGb`; `mongodb.*Resources.limits`; `mongodb.storage.waitPvcBound`; `backup.storage.waitPvcBound`; `recycler.install`; `tls.mode`; `prometheusExporter.alerts.common.*`.

### Forbidden Actions

Never recommend without explicit confirmation:
- `rs.reconfig(...)` on an operator-managed replica set.
- `db.adminCommand({setParameter: ...})` for persistent settings (use `mongodb.customDataRSParameters`).
- `rm -rf /data/...` unless another replica-set member is `PRIMARY`/`SECONDARY`.
- `kubectl delete pvc` for MongoDB data volumes.
- Manual edits to `last-applied-configuration-info` outside documented switchover retry steps.

## Common Failure Modes

For each failure mode below, detailed step-by-step remediation is in the corresponding section of `references/troubleshooting.md`.

### 1. Deployment Failures

- **CRD Validation Failures:** `spec.mongodb.customDataRSParameters: Invalid value: "null"` or `spec.schemaSettings.otherDomainName: Invalid value: "null"`. Set valid defaults in the `mongodb-operator` chart values.
- **PVC Binding Timeout:** `Reconciliation exception: PVC ... 'Bound' status waiting failed`. Set `mongodb.storage.waitPvcBound: false` and/or `backup.storage.waitPvcBound: false`.

### 2. Replica Set / Replication Issues

- **Replica Stuck in Recovering/Unhealthy:** Alerts `RECOVERING`, `MongoDBRecoveringTooLong`, `UNKNOWN`, `DOWN`, `ROLLBACK`, `REMOVED`; log `replication: Data too stale, halting replication`; WiredTiger assertion `28595`; or split-brain. Verify another member is `PRIMARY`/`SECONDARY` before clearing data.
- **Replication Lag:** Alert `MongoDB replication lag`. Check disk IOPS, network latency, large writes, resource usage.

### 3. Storage / Disk Issues

- **No Free Space Left on PV:** Pods rebooting at 100% disk. Scale down a member, free data, run `compact`, scale back up.
- **Disk Space Not Reclaimed After Delete:** Run `db.runCommand({compact: "<collection>"})` on each member, stepping down the primary first.

### 4. Performance / Resource Issues

- **WiredTiger Cache Pressure (`WT_CACHE_FULL`):** Read latency spikes / `MongoWiredTigerCachePressure`. Raise `mongodb.*WiredTigerCacheGb` and matching memory limits.
- **High CPU / Memory Usage:** Alerts `MongoDB CPU Usage` / `MongoDB Memory Usage`. Check `db.currentOp` / `system.profile`; optimize queries/indexes or raise limits.
- **Too Many Open Files:** `Error accepting new connection...` / `Too many open files`; may show `WiredTiger error (24)` or fatal assertion 50882/50853. Raise containerd `LimitNOFILE` >= 64000.

### 5. Backup Issues

- **Backup Job Failed:** Alert `MongoDB last backup failed`. Check backup daemon `.console` log for `Backup failed`; PVC/inodes, S3 credentials, or I/O saturation.
- **No Backup Information in Monitoring:** Grafana empty; `mongodb-monitoring-agent` logs `AttributeError: 'NoneType' object has no attribute 'get'`. Grant `view` role to the `monitoring-agent` service account.

### 6. DR Issues

- **Switchover / Failover Failed:** Operation does not complete. Delete `summary-spec` in `last-applied-configuration-info`, restart operator, retry.
- **Clean Install Fails on One Side of a DR Pair:** Standby was not deployed first or domain names not swapped. Deploy standby first, then active, with swapped `thisDomainName`/`otherDomainName`.
- **Transitioning from HA to DR Fails:** Operator logs `Reconciliation exception: Microservice execution exception: cannot get major and minor version from mongo version` or `schemaType` changed `ha`→`dr` without standby members added. Use operator **2.16.6+** and follow the [HA to DR Transition](https://github.com/Netcracker/qubership-mongodb-operator/blob/main/docs/public/installation_guide.md#ha-to-dr-transition) procedure.
- **Prometheus Exporter Authentication Errors After Switchover:** Exporter logs `AuthenticationFailed` / `sasl conversation error` / `ReplicaSetNoPrimary` after switchover. Set `EXPORT_MONGOS=false` on the new standby side; upgrade to operator **2.16.10+** / mongodb-services **master-supplementary-20260921.033143-345+**.

### 7. Authentication & Connectivity

- **Cannot Log In to MongoDB:** `UserNotFound: Could not find user root@admin` or mongos `can't choose primary for datars<N>`. Initiate replica set, create root/monitoring users, restart mongos.
- **Intra-Cluster Authentication Fails During Rolling Update:** SCRAM-SHA-256 / `SASL step` failures for `__system`. Keyfile rotation mismatch; verify keyfile consistency and restart affected mongod/mongos pods.
- **Mongos Did Not Restart After Keyfile Change:** After rolling update, mongos keeps old keyfile while datars/cnfrs have new one; `AuthenticationFailed: SCRAM authentication failed, storedKey mismatch` for `__system`. Restart mongos (`kubectl rollout restart deployment/mongos`).

### 8. Pod Startup Issues

- **Pods Do Not Start With `context deadline exceeded`:** Container runtime issue. Restart Docker/containerd on the node.
- **Pods Stuck in Pending State:** Exited init-container cleaned by node cronjob. Delete pods one-by-one and update operator.
- **Index Build Stuck:** Log `Index build: waiting for next action before completing final phase`. Fix unhealthy replica (resync), retry index build.

### 9. DBaaS Adapter / Backup Registration Issues

- **Adapter Reports backupRestore: false:** `some adapters do not supported backup operation`, `"backupRestore": false`, `Backup credentials are not set. Backup API is not working.`, missing `BACKUP_DAEMON_API_CREDENTIALS_USERNAME` / `PASSWORD` env vars. Touch any `spec.dbaas` field and redeploy, or patch the Deployment directly.

## Diagnostic Workflow

### Initial Triage

1. **Logs:** operator, `datarsN`/cnfrs mongod, mongos, backup daemon, monitoring agent, prometheus exporter.
2. **State:** `kubectl get mongodbdeployment,mongodbsupplservice -o yaml`; pod, PVC, and event listings; `rs.status()` from a mongod pod.
3. **Use the symptom-to-section table below** to find the failure mode, then read only the matching section in `references/troubleshooting.md`.

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
| Mongos did not restart after keyfile rotation; `AuthenticationFailed: SCRAM authentication failed, storedKey mismatch` for `__system` | [7. Authentication & Connectivity - Mongos Did Not Restart After Keyfile Change](#7-authentication--connectivity) |
| `Error accepting new connection on local endpoint` / `Too many open files` | [4. Performance / Resource Issues - Too Many Open Files](#4-performance--resource-issues) |
| `WiredTiger error (24)` or fatal assertion 50882/50853 immediately after `Too many open files` | [4. Performance / Resource Issues - Too Many Open Files](#4-performance--resource-issues) |
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
| `cannot get major and minor version from mongo version` while converting HA to DR | [6. DR Issues - HA to DR Transition](#6-dr-issues) |
| `schemaSettings.schemaType` changed from `ha` to `dr` but standby members are never added | [6. DR Issues - HA to DR Transition](#6-dr-issues) |
| Transitioning existing HA deployment to active-standby DR fails | [6. DR Issues - HA to DR Transition](#6-dr-issues) |
| Exporter `AuthenticationFailed` / `sasl conversation error` after DR switchover | [6. DR Issues - Prometheus Exporter Authentication Errors After Switchover](#6-dr-issues) |
| Exporter `ReplicaSetNoPrimary` for mongos endpoints after switchover | [6. DR Issues - Prometheus Exporter Authentication Errors After Switchover](#6-dr-issues) |
| `EXPORT_MONGOS` not `false` on standby-side prometheus-exporter | [6. DR Issues - Prometheus Exporter Authentication Errors After Switchover](#6-dr-issues) |
| `context deadline exceeded` in Events | [8. Pod Startup Issues - Pods Do Not Start](#8-pod-startup-issues) |
| Pod phase `Pending` (older installations) | [8. Pod Startup Issues - Pods Stuck in Pending](#8-pod-startup-issues) |
| `Index build: waiting for next action before completing final phase` | [8. Pod Startup Issues - Index Build Stuck](#8-pod-startup-issues) |

## Using the Full Troubleshooting Guide

For step-by-step remediation, read only the matching section from `references/troubleshooting.md` (or `docs/public/troubleshooting.md`) — **do NOT read the full file**.

1. Find the section line: `grep -n "^## " references/troubleshooting.md`
2. Match the symptom/alert to a section header using the table above.
3. Read only that section with `offset = section start line` and `limit = lines until the next "## " header`.

## Key Helm Chart Parameters

```yaml
schemaSettings:
  sharded: true
  schemaType: "ha"          # ha | dr
  thisDomainName: "cluster.local"
  otherDomainName: "peer.cluster.local"  # swap with peer side in DR

disasterRecovery:
  mode: "active"

mongodb:
  customDataRSParameters: '["logLevel=0"]'
  dataWiredTigerCacheGb: "1.0"
  cnfWiredTigerCacheGb: "0.5"
  singleWiredTigerCacheGb: "0.25"
  dataResources: { limits: { memory: 1Gi, cpu: 500m } }
  storage:
    waitPvcBound: false
    size: ["5Gi"]

backup:
  storage:
    waitPvcBound: false

recycler:
  install: false

tls:
  mode: disabled
```

## Log Analysis Tips

```bash
# Operator: reconciliation, CRD, DR / switchover / HA-to-DR
kubectl logs deployment/mongodb-operator | grep -iE "error|failed|unable|exception|MongodbDeployment|MongodbSupplService|switchover|failover|disasterRecovery|schemaType|major and minor version"

# Mongod: replication, WiredTiger, EMFILE, auth, crashes
kubectl logs <datarsN-pod> -c mongod | grep -iE "replication|recover|rollback|stale|election|wiredtiger|wt_cache_full|assertion|corrupt|too many open files|emfile|error \(24\)|fatal assertion 50882|fatal assertion 50853|sasl|authentication|userNotFound|starting|shutdown|crash|fatal"

# Mongos: routing, connection limits
kubectl logs <mongos-pod> | grep -iE "can't choose primary|read preference|config server|too many open files|connection refused"

# Prometheus exporter: post-switchover auth / mongos scrape errors
kubectl logs <mongodb-prometheus-exporter-pod> | grep -iE "cannot connect to mongodb|sasl conversation error|replicasetnoprimary|export_mongos|error|cannot connect"

# Backup daemon: backup/credentials/storage
kubectl logs <mongodb-backup-daemon-pod> | grep -iE "backup|restore|snapshot|s3|credentials|permission denied|disk full"
```

## Integration Points

- **Monitoring exporter:** Prometheus `ServiceMonitor`; health-check via `kubectl port-forward <mongodb-monitoring-agent-pod> 9216:9216` and `curl localhost:9216/metrics`; thresholds in `prometheusExporter.alerts.common.*`.
- **DBaaS adapter:** `dbaas-mongo-adapter` container; common symptom is bulk-drop timeouts / stuck DDL.
- **Backup storage:** `/backups/<timestamp>` in PVC or S3; verify `mongodb-backup-credentials` Secret and PVC space/inodes when backups fail.

## Prevention Best Practices

1. Set `mongodb.customDataRSParameters` as a JSON-encoded `"key=value"` list; redeploy replaces the whole list, so include existing entries.
2. Use typed `mongodb.*WiredTigerCacheGb` and raise cache + memory limits together (cache < pod memory limit).
3. Deploy DR standby first with swapped `thisDomainName`/`otherDomainName`; if standby isn't ready, deploy HA first and use HA-to-DR migration (operator 2.16.6+).
4. Use `recycler.install: true` or wipe PVs for clean reinstalls.
5. Monitor disk usage/oplog window and expand PVCs before they fill.
6. Test backup/restore and DR failover/switchover regularly in non-production.
7. Enable TLS in production (`preferTLS`/`requireTLS`).
8. Keep containerd `LimitNOFILE` >= 64000 on MongoDB worker nodes.

## Related Documentation

- [Installation Guide](https://github.com/Netcracker/qubership-mongodb-operator/blob/main/docs/public/installation_guide.md) — Prerequisites, Helm parameters, platform-specific setup
- [Configuration Guide](https://github.com/Netcracker/qubership-mongodb-operator/blob/main/docs/public/configuration.md) — CRD fields and value reference
- [Architecture](https://github.com/Netcracker/qubership-mongodb-operator/blob/main/docs/public/architecture.md) — Cluster topology and component architecture
- [Operations Guide](https://github.com/Netcracker/qubership-mongodb-operator/blob/main/docs/public/operations_guide.md) — Day-2 operations
- [Troubleshooting Guide](https://github.com/Netcracker/qubership-mongodb-operator/blob/main/docs/public/troubleshooting.md) — Detailed issue-by-issue remediation
- [DR Operations](https://github.com/Netcracker/qubership-mongodb-operator/blob/main/docs/public/dr_operations.md) — Active-standby DR procedures
- [Backup Daemon](https://github.com/Netcracker/qubership-mongodb-operator/blob/main/docs/public/backup-daemon.md) — Backup/restore configuration
- [Grafana Dashboard](https://github.com/Netcracker/qubership-mongodb-operator/blob/main/docs/public/grafana_dashboard.md) — Monitoring dashboards

## When to Escalate

Escalate to maintainers for: operator Go panic; suspected data corruption (`Assertion: 28595`, checksum errors, unrecoverable WiredTiger files); split-brain with divergent primaries; all replica-set members down without backup; repeated backup-restore failures; unrecoverable DR switchover/failover; inconsistent state after upgrade/rolling update; unexplained performance degradation.

**Include:** operator logs, mongod logs from affected replica sets, CRD YAML (`kubectl get mongodbdeployment,mongodbsupplservice -o yaml`), `rs.status()` from each set, `kubectl get events`, and backup daemon `.console` log for failing backups.
