---
name: troubleshoot-mongodb
description: Diagnose failures in operator-managed MongoDB clusters (sharded topology, HA/DR pairs, backups) by matching an exact alert name, log/error string, or replica-set state (RECOVERING, UNKNOWN, DOWN, ROLLBACK, REMOVED) to a section of references/troubleshooting.md. Covers the alerts defined in prometheus-alerts.yaml of the mongodb-services chart, WiredTiger assertion and WT_CACHE_FULL errors, PVC/disk problems, SCRAM/keyfile auth errors, "Too many open files", mongos routing errors, failed deploys, failed backups, DR switchover and DR clean-install failures.
---

## Finding the section

Run from the skill directory:

1. `grep -n "^## " references/troubleshooting.md` — lists all 32 issue headers with line numbers. Every issue title is a
   `##` line; the only `###` lines are Description / Alerts / Stack trace(s) / How to solve / Recommendations.
2. Look the alert name, error string, or state up in the tables below. Tables give the exact header text; use it to
   pick the line number from step 1.
3. To search by literal string instead, run `grep -n -F "<exact string>" references/troubleshooting.md` (for
   example `grep -n -F "Data too stale" references/troubleshooting.md`). Hits inside `### Alerts` or
   `### Stack trace(s)` blocks identify the section; read from that section's `##` line to the next `##` line.
4. Read only that line range with offset/limit. Do not load the whole file.

If the wording is not an exact match, search again with one distinctive term from it (for example `stale`, `Bound`,
`replSetInitiate`, `SASL`, `endpoint`) using `grep -n -i -F "<term>" references/troubleshooting.md`. Use a hit only
if the term appears in a section's `### Stack trace(s)`, `### Alerts`, or `## ` header line. Then:

- Present the section as an **inferred match**. Quote the string from that section's Stack trace(s) block and say the
  fix applies only if the reported logs contain that string.
- Do not infer from a generic term (`error`, `failed`, `pod`, `timeout`, `restart`, `down`) or from a term that only
  appears in a Description or How to solve block.

If no distinctive term hits, the issue is undocumented: say so, and do not apply the nearest-looking section.

## Alert name → section

Alert names are the `alert:` values in `services/service/charts/helm/mongodb-services/templates/prometheus-alerts.yaml`
(`PrometheusRule` `prometheus-mongodb-rules.v1`). Match the name exactly.

| Alert name | Header in references/troubleshooting.md |
|---|---|
| `MongoDB replication lag` | Prometheus Alerts Troubleshooting - MongoDB Replication Lag |
| `MongoDB replication Status 3 (RECOVERING)` | Prometheus Alerts Troubleshooting - MongoDB Replication Status 3 |
| `MongoDBRecoveringTooLong` | Prometheus Alerts Troubleshooting - MongoDB Replication Status 3 |
| `MongoDB replication Status 6 (UNKNOWN)` | Prometheus Alerts Troubleshooting - MongoDB Replication Status 6 |
| `MongoDB replication Status 8 (DOWN)` | Prometheus Alerts Troubleshooting - MongoDB Replication Status 8 |
| `MongoDB replication Status 9 (ROLLBACK)` | Prometheus Alerts Troubleshooting - MongoDB Replication Status 9 |
| `MongoDB replication Status 10 (REMOVED)` | Prometheus Alerts Troubleshooting - MongoDB Replication Status 10 |
| `MongoDB cursors timeouts` | Prometheus Alerts Troubleshooting - MongoDB Cursors Timeouts |
| `MongoDB CPU Usage` | Prometheus Alerts Troubleshooting - MongoDB CPU Usage |
| `MongoDB Memory Usage` | Prometheus Alerts Troubleshooting - MongoDB Memory Usage |
| `MongoWiredTigerCachePressure` | Common Issues - MongoDB Read Latency Spikes / WT_CACHE_FULL (Not Corruption) |
| `MongoDB last backup failed` | Common Issues - Backup Failed for Some Reason |

These alerts have **no section** in the guide — state that no documented fix exists; do not map them to a nearby
section: `MongoDB DBaaS Adapter is down`, `MongoDB Backup Daemon is down`,
`MongoDB Backup Daemon uses more than <N> % of disk space`, `MongoDB Backup Daemon uses more than <N> % of available inodes`,
`MongoBalancerDisabled`, `MongoChunkMigrationFailures`, `MongoUnevenChunkDistribution`, `MongoTooManyConnections`.

An alert named `Pod is Down` does not exist in this repo's alert rules. If a report uses that name, treat it as a
pod crash/restart symptom and match by the pod's log or event string using the table below, not by the alert name.

## Error string / state → section

| Exact string or state | Header in references/troubleshooting.md |
|---|---|
| `Assertion: 28595:2: No such file or directory` with `wiredtiger_kv_engine.cpp 267` | Common Issues - Mongodb Statefulset is Not Starting With Error |
| Same assertion, and every pod of that replica set (`datarsN*` or `cnfrsN*`) is down | Common Issues - Mongodb Statefulset is Not Starting With Error and You Do Not Have Enough Alive Pods to Sync From |
| `replication: Data too stale, halting replication` | Common Issues - Mongo Pod Was Down for Some Time and Now Cannot Start |
| `Could not find host matching read preference { mode: "primary" } for set datarsX` | Common Issues - Could not find host matching read preference for set datarsX |
| `Index build: waiting for next action before completing final phase` | Common Issues - Index build stuck |
| `Reconciliation exception: PVC <name> 'Bound' status waiting failed` | Common Issues - Deploy Failed With Error in Operator Logs |
| `spec.mongodb.customDataRSParameters: Invalid value: "null"` or `spec.schemaSettings.otherDomainName: Invalid value: "null"` | Common Issues - Deploy Fails |
| `command replSetInitiate requires authentication` in operator logs | Common Issues - Clean deploy failed with error message in operator logs |
| `context deadline exceeded` in pod Events | Common Issues - Pods Do Not Start With Error in Events |
| Pod phase `Pending` | Common Issues - Pods Stuck in Pending State |
| `UserNotFound: Could not find user root@admin` or mongos log `can't choose primary for datars<N>` | Common Issues - Cannot Login to MongoDB |
| `SCRAM-SHA-256` / `SASL step` failures for user `__system` between members during a rolling update | Common Issues - Intra-Cluster Authentication Fails During Rolling Update (Keyfile Rotation) |
| `Error accepting new connection on local endpoint` or `Too many open files` | Common Issues - Connections Refused With "Too Many Open Files" Error |
| `WT_CACHE_FULL` in mongod log, no crash | Common Issues - MongoDB Read Latency Spikes / WT_CACHE_FULL (Not Corruption) |
| mongodb-monitoring-agent log `AttributeError: 'NoneType' object has no attribute 'get'` | Common Issues - No Backup Information in Monitoring |
| Backup log line starting `Backup failed` in `/backups/<folder>/.console` | Common Issues - Backup Failed for Some Reason |
| Replica `stateStr` = `RECOVERING` | Prometheus Alerts Troubleshooting - MongoDB Replication Status 3 |
| Replica `stateStr` = `UNKNOWN` | Prometheus Alerts Troubleshooting - MongoDB Replication Status 6 |
| Replica `stateStr` = `DOWN` | Prometheus Alerts Troubleshooting - MongoDB Replication Status 8 |
| Replica `stateStr` = `ROLLBACK` | Prometheus Alerts Troubleshooting - MongoDB Replication Status 9 |
| Replica `stateStr` = `REMOVED` | Prometheus Alerts Troubleshooting - MongoDB Replication Status 10 |
| Two members of one replica set both report `stateStr` = `PRIMARY` | Common Issues - Same Queries Return Different Results |

## Symptom (no exact string) → section

| Symptom | Header in references/troubleshooting.md |
|---|---|
| DR switchover/failover fails or times out | DR Issues - Swithover/Failover failed |
| Clean install succeeds on one DR side and fails on the other | DR Issues - Clean Install Fails on Only One Side of a DR Pair |
| All MongoDB pods restart repeatedly and the PV is full | Common Issues - No Free Space Left on MongoDB PV |
| Collection data deleted, PV usage did not drop | Common Issues - MongoDB Disk Space Usage Not Changed After Data Deleted |
| Same query returns different results between runs | Common Issues - Same Queries Return Different Results |
| Pods restart during the backup window | Common Issues - Random Pods Restart During MongoDB Backup Process |
| Primary changes repeatedly within a short period | Common Issues - Replica Set Primary Flips Repeatedly (Stepdown Loops) |
| Backup metrics missing in monitoring | Common Issues - No Backup Information in Monitoring |

Section header spelling is exact, including `Swithover`.

## Guardrails

The operator rewrites replica-set configuration and mongod parameters on every reconcile loop. Never recommend:

- `db.adminCommand({setParameter: ...})` for a persistent setting — lost on pod restart. Change
  `mongodb.customDataRSParameters` or the dedicated typed value (see "Config value conventions") instead.
- `rs.reconfig(...)` on an operator-managed replica set — the operator overwrites it on the next reconcile.
- Any data-clearing step (`rm -rf /data/...`, `rm -f /data/*`) unless at least one member of the same replica set is
  `PRIMARY` or `SECONDARY` at that moment. If every member of a replica set is down, the data is recoverable only
  from backup; say so and do not recommend a resync.

## Config value conventions

Where each parameter lives:

| Parameter | Resource it lives on |
|---|---|
| `mongodb.customDataRSParameters` | `mongodb-operator` Helm chart values → `MongodbDeployment` CR `spec.mongodb.customDataRSParameters` |
| `mongodb.dataWiredTigerCacheGb`, `mongodb.cnfWiredTigerCacheGb`, `mongodb.singleWiredTigerCacheGb` | same chart/CR (`spec.mongodb.*`); data shards / config servers / standalone respectively |
| `mongodb.dataResources.limits.memory` / `.cpu`, `mongodb.cnfResources.limits.*`, `mongodb.mongosResources.limits.*` | same chart/CR; limits of `datarsN`, `cnfrsN`, and mongos pods respectively |
| `mongodb.storage.waitPvcBound` | `mongodb-operator` chart / `MongodbDeployment` CR — MongoDB PVCs |
| `backup.storage.waitPvcBound` | `mongodb-services` chart / `MongodbSupplService` CR — backup daemon PVC |
| `recycler.install` | `mongodb-operator` chart / `MongodbDeployment` CR |
| `disasterRecovery.mode` (`active` / `standby` / `disable`) | `mongodb-operator` chart / `MongodbDeployment` CR `spec.disasterRecovery.mode` |
| `schemaSettings.thisDomainName`, `schemaSettings.otherDomainName` | `mongodb-operator` chart / `MongodbDeployment` CR `spec.schemaSettings.*` |
| `prometheusExporter.alerts.common.*` (`replicationLagSeconds`, `cursorTimeoutThreshold`, `cpuUsageThreshold`, `memoryUsageThreshold`, `cachePressureThreshold`) | `mongodb-services` chart values (alert thresholds) |

Rules:

- `mongodb.customDataRSParameters` is a list of `"key=value"` strings (`=`, not `:`), written in values as a
  JSON-encoded string, for example `'["logLevel=0", "diagnosticDataCollectionEnabled=false"]'`. Redeploy replaces the
  list, it does not merge: the recommended value must contain every existing entry plus the new one.
- For WiredTiger cache size, use the typed `mongodb.*WiredTigerCacheGb` value, not `customDataRSParameters`.
- DR domain names must be swapped between sides: side A `thisDomainName` = side B `otherDomainName`, and
  side A `otherDomainName` = side B `thisDomainName`.
- Present every config change as "set `<value-path>: <value>` in the deploy parameters of `<chart>` and redeploy".
- Replica counts are `schemaSettings.cnfReplicaSize`, `schemaSettings.dataReplicaSize` and `schemaSettings.arbiterIndex`
  on the `MongodbDeployment` CR. Never change membership with `rs.reconfig`.
