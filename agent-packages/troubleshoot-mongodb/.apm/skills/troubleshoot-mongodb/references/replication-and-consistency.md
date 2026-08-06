# Replication State and Data Consistency Issues

## Index Build Stuck: "waiting for next action before completing final phase"

**Description:** Since MongoDB 4.4, index builds run simultaneously across all data-bearing replica set members and need a commit quorum of members to finish before the index is marked ready. If a member isn't in PRIMARY/SECONDARY state, the build stalls.

**How to solve:**
1. From a mongos pod: `mongosh admin -u root -p <pass> --eval="sh.status()"` to find the primary shard of the affected database.
2. On that shard's first pod, check member states: `mongosh admin -u root -p <pass> --eval='rs.status().members.forEach(m=>{print(m.name);print("\t"+m.stateStr)})'`.
3. For the replica stuck outside PRIMARY/SECONDARY (e.g. RECOVERING), clear its data directory (`rm -rf /data/<name>/*`) and reboot the pod.
4. Wait for it to rejoin as SECONDARY.
5. Retry the index build.

## `Could not find host matching read preference { mode: "primary" } for set datarsX`

**Description:** One or more replicas in `datarsX` are in an unexpected state.

**How to solve:**
1. On any pod of `datarsX`: `mongosh admin -u root -p <pass> --eval='rs.status().members.forEach(m=>{print(m.name);print("\t"+m.stateStr)})'`.
2. **Important:** at least one member must be PRIMARY or SECONDARY. If none are, stop — do not continue with data-clearing steps; that data isn't recoverable this way.
3. For each replica with unexpected status (not PRIMARY/SECONDARY):
   ```
   kubectl exec <pod> -- sh -c 'rm -rf /data/<name>/*'
   kubectl delete pod <pod>
   ```
4. Wait for each to rejoin as SECONDARY.

## Same Query Returns Different Results Across Runs

**Description:** Two likely causes:
- **Split-brain** — more than one member reports `PRIMARY`.
- One server is stale and `ReadFromSecondary` is enabled.

**How to solve:**
1. Run `rs.status()` on each replica set and look for more than one member with `"stateStr" : "PRIMARY"` — that confirms split-brain.
2. If split-brain is confirmed, you must sacrifice one of the primaries and delete all its data (resync it as if it were a failed member — see `storage-and-corruption.md`).

**Caveat:** this can harm the application, since some documents may be unique to whichever primary is discarded — flag this to the user before proceeding, it's not a purely mechanical fix.

## Replica Stuck in RECOVERING / ROLLBACK / UNKNOWN / DOWN / REMOVED

These map to the Prometheus "MongoDB Replication Status" alerts (3/6/8/9/10) — see `prometheus-alerts.md` for the per-state detail. In all cases, start by checking pod logs for the specific error driving the state before taking any recovery action; the underlying causes range from a normal resync-in-progress to network partition to permanent removal from the set.
