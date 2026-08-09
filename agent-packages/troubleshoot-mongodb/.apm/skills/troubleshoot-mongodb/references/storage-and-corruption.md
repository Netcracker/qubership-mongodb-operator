# Storage, Disk Space, and WiredTiger Corruption

## No Free Space Left on Mongodb PV

**Description:** All MongoDB pods can start rebooting if the underlying PV runs out of free space.

**How to solve:**
1. In the `mongo-cluster` namespace, scale down the affected pods (e.g. `datars10-0`, `datars21-0`, `datars32-0`).
2. Connect to the nodes/NFS backing the PVs.
3. Remove the contents of the corresponding PV folders (e.g. `/data/datars10`, `/data/datars21`, `/data/datars32`) — this lets the other pods in those replica sets start.
4. Remove unnecessary data from MongoDB collections.
5. Run `compact` on the affected collections — see "Disk Usage Not Changed After Data Deleted" below.
6. Scale the pods back up.

## Disk Usage Not Changed After Data Deleted

**Description:** Deleting documents from a collection doesn't reclaim disk space on its own — a `compact` must be run.

**How to solve:**
1. From a mongos pod, find the primary shard for the database: `mongosh admin -u root -p <pass> --eval="sh.status()"`.
2. On the first pod of that shard's replica set, check member states: `mongosh admin -u root -p <pass> --eval='rs.status().members.forEach(m=>{print(m.name);print("\t"+m.stateStr)})'`.
3. On each SECONDARY:
   ```
   mongosh admin -u root -p <pass>
   use <db_name>
   db.runCommand({compact: "<collection_name>"})
   ```
4. On the PRIMARY, step down first, then compact:
   ```
   mongosh admin -u root -p <pass>
   rs.stepDown()
   use <db_name>
   db.runCommand({compact: "<collection_name>"})
   ```

## StatefulSet Not Starting: `Assertion: 28595:2: No such file or directory .../wiredtiger_kv_engine.cpp 267`

**Description:** The backend filesystem failed and the data for that instance is corrupted. This instance's data cannot be recovered directly — it must be resynced from other members.

**Before doing anything destructive:** confirm at least one pod is alive in every replica set that needs to be resynced (e.g. one of `datars10-0`/`11-0`/`12-0`, one of `datars2*`, one of `cnfrs*`, etc.). If an entire replica set is down, its data is lost and cannot be recovered this way — see the "not enough alive pods" case below instead.

**How to solve (with filesystem access):**
1. Scale the affected StatefulSet to 0 replicas.
2. Clean the PV folder for that StatefulSet (e.g. `<pv_root>/<statefulset_name>`).
3. Scale back up and let it resync from the surviving members.

**How to solve (without filesystem access, via OpenShift `oc`):**
1. `oc get statefulset <name> -o yaml > /tmp/<name>.yaml` and back it up: `cp /tmp/<name>.yaml /tmp/<name>.yaml_back`.
2. In the `_back` copy, change the `spec.template.spec.args` entry that runs the keyfile-copy command to `sleep 5000`, and set `spec.replicas: 1`.
3. `oc delete statefulset <name>`, then `oc create -f /tmp/<name>.yaml_back`.
4. `oc rsh <name>-0` and inspect `/data` to confirm the corrupted data files (WiredTiger.*, collection-*.wt, index-*.wt, etc.).
5. `rm -f /data/*` to clear the corrupted data.
6. `oc delete statefulset <name>`, then recreate with the **original** config: `oc create -f /tmp/<name>.yaml`.
7. `oc scale <name> --replicas=1` and wait for it to reach Running 1/1 while it resyncs.

## StatefulSet Not Starting AND Not Enough Alive Pods to Sync From

**Description:** An entire replica set is down — there is no easy recovery path except restoring from backup.

**How to solve (if an empty-but-working database is acceptable):**
1. Clear data in all `datars*` and `cnfrs*` pods as described above.
2. Clone the repository to a machine with cluster access, log in via `oc`.
3. Set the required environment variables at the top of `./scripts/re-init.sh`.
4. Run `./scripts/re-init.sh` and confirm the run/smoke tests pass.

If data must be preserved, restore from backup instead of running this script.

## Pod Down for a While, Then Won't Start: `replication: Data too stale, halting replication`

**Description:** The pod fell too far behind the oplog window while it was down and can no longer catch up via normal replication.

**How to solve:** Resync the pod using the same procedure as the WiredTiger corruption case above (clear its data directory and let it resync from surviving members).
