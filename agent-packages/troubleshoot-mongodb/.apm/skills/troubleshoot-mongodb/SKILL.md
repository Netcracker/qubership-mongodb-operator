---
name: troubleshoot-mongodb
description: Diagnose and resolve failures in operator-managed MongoDB clusters — sharded topology (mongos, config servers, shard replica sets), HA/DR site pairs, and backups. Use whenever the user reports a MongoDB pod crash or stuck replica state (RECOVERING/ROLLBACK/UNKNOWN/DOWN/REMOVED), a failed deploy or reconciliation error, a failed switchover/failover, a WiredTiger corruption or "no such file or directory" assertion, disk space or PVC binding problems, an authentication or "can't choose primary" error, connections being refused with a "too many open files" error, a failed or missing backup, or a Prometheus alert about MongoDB (replication lag, cursor timeouts, CPU, memory, replication status). Matches the reported symptom or error string against known, documented failure modes and their resolution steps; falls back to a general diagnostic checklist when nothing matches rather than improvising.
---

# Troubleshoot: MongoDB (Operator-Managed Clusters)

## Scope

Covers MongoDB clusters deployed and reconciled by the custom Kubernetes operator: sharded topology (mongos routers, config server replica set `cnfrsN`, shard replica sets `datarsN`), active/standby DR site pairs, and the backup daemon. This skill is derived from the component's own git-hosted guide at `operator/docs/public/troubleshooting.md` — that file is the source of truth; if it's updated, regenerate `references/` from it rather than editing them independently.

## How to use this skill

1. **Get the symptom.** Pull the exact error string, stack trace, Prometheus alert name, or replica state (from `rs.status()` or pod Events) — don't guess from a vague description if you can get the real text.
2. **Match it against the symptom index below** and open the corresponding file under `references/`.
3. **Follow the "How to solve" steps in that file.** Many steps are destructive — deleting pod data, scaling a StatefulSet to 0, clearing a PV folder. Confirm with the user before instructing them to run anything destructive, especially against a production or DR site.
4. **If nothing matches**, say so plainly and use the general diagnostic checklist at the end of this file. Don't stretch a documented fix to cover an issue it wasn't written for.
5. **If a genuinely new failure mode surfaces**, suggest the user capture it in `operator/docs/public/troubleshooting.md` so it's documented for next time.

## Symptom index

| Symptom / error / alert | Reference file |
|---|---|
| Switchover/failover fails or times out | `references/dr-and-failover.md` |
| MongoDB pods rebooting, no free space left on PV | `references/storage-and-corruption.md` |
| Disk usage unchanged after deleting collection data | `references/storage-and-corruption.md` |
| `Assertion: 28595:2: No such file or directory .../wiredtiger_kv_engine.cpp 267` | `references/storage-and-corruption.md` |
| StatefulSet won't start — not enough alive pods left to sync from (data loss) | `references/storage-and-corruption.md` |
| Pod down for a while, then `replication: Data too stale, halting replication` | `references/storage-and-corruption.md` |
| `Reconciliation exception: PVC ... 'Bound' status waiting failed` in operator logs | `references/deploy-and-startup-issues.md` |
| Deploy fails: `customDataRSParameters` / `otherDomainName` invalid value `"null"` | `references/deploy-and-startup-issues.md` |
| Pods stuck in `Pending` state | `references/deploy-and-startup-issues.md` |
| Pods not starting, `context deadline exceeded` in Events | `references/deploy-and-startup-issues.md` |
| Clean deploy fails: `command replSetInitiate requires authentication` | `references/deploy-and-startup-issues.md` |
| Index build stuck: "waiting for next action before completing final phase" | `references/replication-and-consistency.md` |
| `Could not find host matching read preference { mode: "primary" } for set datarsX` | `references/replication-and-consistency.md` |
| Same query returns different results across runs (possible split-brain) | `references/replication-and-consistency.md` |
| Replica stuck in RECOVERING / ROLLBACK / UNKNOWN / DOWN / REMOVED | `references/replication-and-consistency.md` |
| Login fails: `SCRAM-SHA-1 authentication failed`, `UserNotFound: root@admin`, or mongos logs `can't choose primary for datarsN` | `references/auth-and-connectivity.md` |
| `Error accepting new connection on local endpoint` / `Too many open files` | `references/auth-and-connectivity.md` |
| Backup fails (`Backup failed` in the backup daemon's `.console` log) | `references/backup-issues.md` |
| No backup data in monitoring / `AttributeError: 'NoneType' object has no attribute 'get'` in monitoring-agent | `references/backup-issues.md` |
| Random pod restarts during backup (I/O wait, e.g. on GlusterFS) | `references/backup-issues.md` |
| Prometheus alert: Replication Lag | `references/prometheus-alerts.md` |
| Prometheus alert: Cursor Timeouts | `references/prometheus-alerts.md` |
| Prometheus alert: CPU Usage / slow queries | `references/prometheus-alerts.md` |
| Prometheus alert: Memory Usage | `references/prometheus-alerts.md` |
| Prometheus alert: Replication Status 3 / 6 / 8 / 9 / 10 | `references/prometheus-alerts.md` |

## Cluster conventions used throughout the references

- Namespace referred to as `mongo-cluster` below — substitute the actual namespace for the deployment in question.
- Shard replica sets: `datarsN` → pods `datarsN0-0`, `datarsN1-0`, `datarsN2-0`, ...
- Config server replica set: `cnfrsN` → pods `cnfrsN-0`, ...
- Check replica state from any member:
  `mongosh admin -u root -p <pass> --eval='rs.status().members.forEach(m=>{print(m.name);print("\t"+m.stateStr)})'`
- Check shard-to-primary mapping from a mongos pod:
  `mongosh admin -u root -p <pass> --eval="sh.status()"`
- **Safety rule that applies across nearly every reference file:** at least one member of a replica set must remain PRIMARY or SECONDARY before attempting any data-clearing recovery step. If an entire replica set is down, the data on it is not recoverable except from backup — don't run resync/clear-data steps in that case, and say so.

## When nothing in the index matches

1. Pull the operator pod logs and the specific StatefulSet/pod Events for the affected component.
2. Run `rs.status()` on the affected replica set(s) and `sh.status()` from a mongos pod to compare actual member states against expected (PRIMARY/SECONDARY only).
3. Check node-level resources — disk space, memory, and file-descriptor limits are the root cause behind several documented issues here (e.g. WiredTiger corruption, pod restarts under backup load).
4. If it's genuinely undocumented, say so directly rather than adapting an unrelated fix, and suggest the user add it to `operator/docs/public/troubleshooting.md` once resolved.
