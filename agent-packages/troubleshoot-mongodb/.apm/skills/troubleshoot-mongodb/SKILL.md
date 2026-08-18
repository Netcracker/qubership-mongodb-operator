---
name: troubleshoot-mongodb
description: Diagnose and resolve failures in operator-managed MongoDB clusters — sharded topology, HA/DR pairs, and backups. Use whenever the user reports a pod crash or stuck replica state (RECOVERING/ROLLBACK/UNKNOWN/DOWN/REMOVED), a failed deploy or reconciliation error, WiredTiger corruption or cache-pressure symptoms, disk/PVC problems, an auth error (including SCRAM/keyfile failures between members during a rolling update), a "too many open files" error, mongos routing errors, a failed/missing backup, a Prometheus alert (replication lag, cursor timeouts, CPU, memory, replication status), or a customDataRSParameters/Helm config value question. Always load when a clean install succeeds on one DR side but fails on the other — an install-ordering issue, not an image/version bug; check the DR section before recommending an upgrade. Matches the reported symptom to documented fixes; falls back to a general checklist, framed for the user to check, when nothing matches.
---

## Reading the reference file

1. Grep issue headers with line numbers: `grep -n "^## " references/troubleshooting.md`. This is
   level-anchored, not content-filtered — a filter that drops lines matching words like "Alerts" or "Description"
   will also drop real issue titles that happen to contain those words (this guide has several, e.g. "Prometheus
   Alert - ..."). Level-anchoring instead relies on `validate_guide_headers.py` enforcing at merge time that only
   issue titles use `##` and only the five fixed sub-headers (Description/Alerts/Stack trace(s)/How to
   solve/Recommendations) use `###` — see that script for the enforcement.
2. If nothing in step 1's output looks like a plausible match, also run `grep -n "^##[^# ]"` before concluding the
   issue is undocumented — this catches a header missing the space after `##` (e.g. `##No Free Space...`), which
   won't show up in step 1 but is almost always a real section that got mistyped, not a genuinely missing one.
3. Match the symptom at hand against the jump table below (or the remaining raw headers if the table is stale) to
   pick the issue.
4. Read only that section: offset at its line number, limit through to the next issue header's line number. Never
   load the whole file for one lookup.

## Symptom → reference section

| Symptom | Header in references/troubleshooting.md |
|---|---|
| Switchover/failover between active and standby DR sites fails or times out | DR Issues - Swithover/Failover failed |
| Clean install succeeds on one side of a DR pair but fails on the other | DR Issues - Clean Install Fails on Only One Side of a DR Pair |
| MongoDB pods rebooting; PV/disk out of space | Common Issues - No Free Space Left on MongoDB PV |
| Deleted data from a collection but disk usage on the PV didn't drop | Common Issues - MongoDB Disk Space Usage Not Changed After Data Deleted |
| Index build stuck at "waiting for next action before completing final phase" | Common Issues - Index build stuck |
| `Could not find host matching read preference { mode: "primary" } for set datarsX` | Common Issues - Could not find host matching read preference for set datarsX |
| Operator logs: `Reconciliation exception: PVC ... 'Bound' status waiting failed` | Common Issues - Deploy Failed With Error in Operator Logs |
| StatefulSet won't start; `Assertion: 28595:2: No such file or directory .../wiredtiger_kv_engine.cpp 267` | Common Issues - Mongodb Statefulset is Not Starting With Error |
| Deploy fails; `customDataRSParameters`/`otherDomainName` invalid value `"null"` | Common Issues - Deploy Fails |
| Same WiredTiger error as above, but no pod in the replica set is alive to resync from | Common Issues - Mongodb Statefulset is Not Starting With Error and You Do Not Have Enough Alive Pods to Sync From |
| Pods not starting; `context deadline exceeded` in Events | Common Issues - Pods Do Not Start With Error in Events |
| Pods stuck in `Pending` | Common Issues - Pods Stuck in Pending State |
| Login fails; `UserNotFound: root@admin` or mongos logs `can't choose primary for datarsN` | Common Issues - Cannot Login to MongoDB |
| No backup data shown in monitoring | Common Issues - No Backup Information in Monitoring |
| A scheduled backup failed | Common Issues - Backup Failed for Some Reason |
| Pod down for a while, now won't start: `replication: Data too stale, halting replication` | Common Issues - Mongo Pod Was Down for Some Time and Now Cannot Start |
| Same query returns different results across runs (possible split-brain) | Common Issues - Same Queries Return Different Results |
| Random pod restarts during backup | Common Issues - Random Pods Restart During MongoDB Backup Process |
| Clean deploy fails; `command replSetInitiate requires authentication` | Common Issues - Clean deploy failed with error message in operator logs |
| Prometheus alert: replication lag | Prometheus Alerts Troubleshooting - MongoDB Replication Lag |
| Prometheus alert / cursor timeout errors | Prometheus Alerts Troubleshooting - MongoDB Cursors Timeouts |
| Prometheus CPU alert; slow queries | Prometheus Alerts Troubleshooting - MongoDB CPU Usage |
| Prometheus memory alert | Prometheus Alerts Troubleshooting - MongoDB Memory Usage |
| Replica stuck in RECOVERING | Prometheus Alerts Troubleshooting - MongoDB Replication Status 3 |
| Replica in UNKNOWN state | Prometheus Alerts Troubleshooting - MongoDB Replication Status 6 |
| Replica shows DOWN | Prometheus Alerts Troubleshooting - MongoDB Replication Status 8 |
| Replica actively in ROLLBACK | Prometheus Alerts Troubleshooting - MongoDB Replication Status 9 |
| Replica REMOVED from the set | Prometheus Alerts Troubleshooting - MongoDB Replication Status 10 |
| Read latency spikes, `WT_CACHE_FULL` log lines, no crash or corruption | Common Issues - MongoDB Read Latency Spikes / WT_CACHE_FULL (Not Corruption) |
| Primary keeps stepping down and re-electing (stepdown loops) | Common Issues - Replica Set Primary Flips Repeatedly (Stepdown Loops) |
| SCRAM-SHA-256 / `SASL step` auth failures between members mid-rolling-update | Common Issues - Intra-Cluster Authentication Fails During Rolling Update (Keyfile Rotation) |
| `Error accepting new connection on local endpoint` / `Too many open files` | Common Issues - Connections Refused With "Too Many Open Files" Error |

Start every diagnosis by getting the exact error text, alert name, or replica `stateStr`, and the affected component
(mongos, `cnfrsN`, `datarsN`, or the backup daemon) — most sections match on the literal error string or state, not a
general description.

Check context/attachment for already-provided error logs, operator output, or replica-set status before asking for it.

If the symptom plausibly matches more than one row (e.g. "stuck in RECOVERING" alone matches three different rows
above for three different reasons), ask which one applies rather than guessing — a fix aimed at the wrong root cause
is worse than a clarifying question.

## Guardrails

The operator reconciles continuously — engine-level changes made outside its normal config path get reverted on the
next reconcile loop, and some manual commands actively fight the operator or are destructive. Don't recommend:

- `db.adminCommand({setParameter: ...})` / `mongo --eval` for persistent settings — reverted on pod restart. Use the
  proper config value path (e.g. `customDataRSParameters` or a typed value — see "Config value conventions" below)
  instead.
- `rs.reconfig(...)` against an operator-managed replica set — the operator rewrites the configuration on its next
  loop, so this doesn't stick and can fight an in-progress reconcile.
- `kubectl exec ... mongo ...` for any DDL/DML — destructive; not something to suggest running ad hoc.

## Config value conventions

- There's no single `mongod.conf` umbrella value. The closest is `mongodb.customDataRSParameters` — a list of
  `"key=value"` strings (equals sign, **not** colon — a common trip wire), itself wrapped as a JSON-encoded string in
  YAML, e.g. `mongodb.customDataRSParameters: '["logLevel=0", "diagnosticDataCollectionEnabled=false"]'`. The value
  is replaced wholesale on redeploy, not merged — a recommended change must preserve the existing entries in the
  list, not just add the new one.
- Several engine settings have dedicated typed values instead — prefer these over `customDataRSParameters` when one
  exists: `mongodb.dataWiredTigerCacheGb` (data shards), `mongodb.cnfWiredTigerCacheGb` (config servers),
  `mongodb.singleWiredTigerCacheGb` (standalone).
- For DR schema values, `schemaSettings.thisDomainName` / `schemaSettings.otherDomainName` swap between the two
  sides — a common misconfiguration is copying the same pair of values to both sides instead of swapping them. See
  the DR install-order section in `references/troubleshooting.md`.
- Phrase every recommended change as "set `<value-path>: <value>` in the deploy configuration and redeploy through
  the normal delivery channel," not as a live in-cluster command — see Guardrails above for why.
- For replica-set topology changes (member count, voting members, arbiters), point to the operator's documented
  procedure — there's no config value for these; they're operational steps, not settings.

## Cluster conventions

- Namespace referred to as `mongo-cluster` below — substitute the actual namespace for the deployment in question.
- Shard replica sets: `datarsN` → pods `datarsN0-0`, `datarsN1-0`, `datarsN2-0`, ...
- Config server replica set: `cnfrsN` → pods `cnfrsN-0`, ...
- Check replica state from any member:
  `mongosh admin -u root -p <pass> --eval='rs.status().members.forEach(m=>{print(m.name);print("\t"+m.stateStr)})'`
- Check shard-to-primary mapping from a mongos pod:
  `mongosh admin -u root -p <pass> --eval="sh.status()"`
- **Safety rule that applies across nearly every section:** at least one member of a replica set must remain PRIMARY
  or SECONDARY before attempting any data-clearing recovery step. If an entire replica set is down, the data on it
  is not recoverable except from backup — don't recommend resync/clear-data steps in that case, and say so.
