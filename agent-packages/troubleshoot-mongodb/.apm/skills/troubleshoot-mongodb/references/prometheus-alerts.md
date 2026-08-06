# Prometheus Alerts

## Replication Lag

**Possible reasons:** secondaries can't keep up with the primary's write rate. Common culprits: network latency/packet loss, slow disk throughput on the secondary, large/long-running writes locking the system and blocking replication, or high write volume outpacing the secondary's oplog read rate.

**How to solve:**
1. Investigate disk I/O and network issues on the environment (see "Random Pod Restarts During Backup" in `backup-issues.md` for one known I/O-pressure pattern).
2. Check for unoptimized queries or missing/incorrect indexes driving load — see CPU Usage and Memory Usage below.

## Cursor Timeouts

**Possible reasons:** unoptimized queries or a large volume of data being pulled in one batch. Default cursor timeout is 10 minutes.

**Mitigation options (each with a trade-off):**
1. Read all needed data at once and process it in memory — simple, but doesn't scale to large result sets.
2. Reduce the default batch size returned per cursor fetch, so each batch takes well under 10 minutes — reduces timeout risk but increases the number of round trips/connections.
3. Set `noCursorTimeout: true` — removes the timeout entirely, but if the client process dies unexpectedly the cursor is never closed and keeps holding server resources until MongoDB is restarted. Use cautiously.

## CPU Usage

**Possible reasons:** incorrect indexes or unoptimized queries causing high CPU load and slow queries.

**How to solve:**
1. Find in-progress slow operations: `db.currentOp({"secs_running": {$gte: 3}})` (add a filter if CPU is pegged at 100%).
2. If the Database Profiler is enabled, inspect `system.profile`. Enable/set the threshold with e.g. `db.setProfilingLevel(1, 1000)` (profile level 1, 1000ms threshold), then query:
   ```
   db.system.profile.find().pretty()
   db.system.profile.find({op: {$eq: 'query'}}, {millis:1, ns:1, ts:1, query:1}).sort({ts:-1}).pretty()
   ```
3. Use `explain('executionStats')` on the suspect query to see `totalKeysExamined` / `totalDocsExamined` versus `nReturned`. A `docsExamined` far higher than `nReturned` (ideally it should be close to zero) indicates a missing or poorly-matched index — the fix is adding/adjusting the index, not a cluster-level change.

## Memory Usage

**Possible reasons:** too much data in the database, or non-optimal index usage.

**How to solve:**
1. Run `db.stats()` and look at `indexSize` and `dataSize` (in bytes) — up to ~50% of this combined size can be cached in RAM under normal working-set behavior.
2. If that's the driver, either increase RAM limits on the `datars` StatefulSets or review the index policy to shrink the working set.

## Replication Status Alerts (per-state)

These map to `rs.status()` member `stateStr` values:

| Status | Meaning | How to solve |
|---|---|---|
| 3 — RECOVERING | Member is performing startup self-checks, or transitioning out of a rollback/resync. | Check pod logs for the specific error driving the state before acting; this can be transient. |
| 6 — UNKNOWN | This member's state, as seen by another member, isn't known yet. | Check pod logs for the specific error driving the state. |
| 8 — DOWN | Member is unreachable from another member's perspective — typically network issues. | Check pod logs and node/network health. |
| 9 — ROLLBACK | Member is actively rolling back to match the primary's oplog. | No direct action — let it complete; investigate only if it doesn't resolve. |
| 10 — REMOVED | Member was once part of the replica set but has since been removed from the config. | No direct action unless it should be re-added. |

For any of these, if a member is stuck outside PRIMARY/SECONDARY beyond a normal resync window, treat it using the same recovery pattern as `replication-and-consistency.md` (clear its data directory and let it resync) — but only after confirming from logs that it isn't mid-rollback/resync already, and only if at least one other member of the set is PRIMARY or SECONDARY.
