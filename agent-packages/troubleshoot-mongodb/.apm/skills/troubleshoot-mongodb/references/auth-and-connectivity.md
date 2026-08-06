# Authentication and Connectivity Issues

## Cannot Login to MongoDB — Unauthorized / `UserNotFound: root@admin`

**Description:** Symptoms include mongos logs showing `can't choose primary for datars${number}` and all queries failing with the same error. Logging in with `mongosh admin` (no credentials) succeeds but no commands can be run — this means MongoDB has never been initialized and no root user exists yet.

**How to solve:**
1. **If the replica set isn't initialized**, initiate it from any member:
   ```
   rs.initiate({_id:'datars${RS_NUMBER}', members: [
     {_id:0, host:'datars${RS_NUMBER}0-0.datars${RS_NUMBER}.${PROJECT}.svc.cluster.local:27017'},
     {_id:1, host:'datars${RS_NUMBER}1-0.datars${RS_NUMBER}.${PROJECT}.svc.cluster.local:27017'},
     {_id:2, host:'datars${RS_NUMBER}2-0.datars${RS_NUMBER}.${PROJECT}.svc.cluster.local:27017'}
   ]});
   ```
2. Log in to the new PRIMARY without credentials (`mongosh admin`) and create the root user:
   ```
   use admin;
   db.createUser({user: 'root', pwd: '${MONGO_PASSWORD}', roles: ['root']});
   ```
3. Log back in with credentials (`mongosh -u root -p ${MONGO_PASSWORD} admin`) and create the monitoring user:
   ```
   db.createUser({user: 'monitoring', pwd: '${MONGO_MONITORING_PASSWORD}', roles: [{role:'readWrite', db:'test'}, 'clusterMonitor']});
   ```
4. Delete the mongos pod to force it to re-resolve shard hosts. Check mongos logs for `can't resolve host <pod>`.
5. **If mongos can't resolve a shard host by short name** (`sh.status()` shows an entry like `datars3/datars31-0:27017` instead of the full DNS name — a known symptom when a shard was originally registered with an incomplete host list): ping the short name from the mongos pod; if it fails but the FQDN resolves, add a temporary `/etc/hosts` entry mapping the short name to that pod's IP. Mongos will re-fetch the correct full config and self-correct — the `/etc/hosts` entry is only needed to unblock that one re-resolution, not permanently.

## Connections Refused: `Error accepting new connection on local endpoint` / `Too many open files`

**Description:** MongoDB logs report `Error accepting new connection on local endpoint`, with the underlying OS error `Too many open files`. This is a file-descriptor exhaustion issue at the container/node level, not a MongoDB config problem — the pod's `ulimit -n` (`LimitNOFILE`) is too low for MongoDB's connection and WiredTiger file-handle usage under load. This matches the pattern behind the EMFILE-triggered WiredTiger fatal assertion outage — left unaddressed, FD exhaustion here can escalate to that.

**How to solve:**
1. Increase the open-file limit for MongoDB pods on all affected worker nodes. Two mechanisms, depending on how the node/runtime is configured:
   - **containerd (most common in this environment):** raise `LimitNOFILE` in the containerd service config (typically a systemd drop-in for `containerd.service`, or `[plugins."io.containerd.runtime.v1.linux"]` / OCI runtime spec settings) on each affected worker node, then restart containerd.
   - **Pod-level fallback:** if node-level access isn't available, set the equivalent limit on the pod spec directly (e.g. via `securityContext` ulimits on supporting runtimes, or an init container that raises it before the MongoDB container starts).
2. MongoDB's own documentation recommends a `ulimit` of 64000 as the minimum for this setting.
3. After raising the limit, restart the affected MongoDB pods (and containerd, if that's the layer changed) so they pick up the new limit.

## General mongos Connectivity Checks

- `sh.status()` from a mongos pod shows the registered host list per shard — compare it against actual pod DNS names to catch stale or incomplete shard registration.
- A shard registered with only one member's short hostname (rather than the full replica set host list) is the specific pattern behind the login/connectivity failure above — it's worth checking even if the symptom looks like a pure auth issue at first.
