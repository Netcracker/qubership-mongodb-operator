# Deploy and Pod Startup Issues

## Deploy Fails: `Reconciliation exception: PVC ... 'Bound' status waiting failed`

**Description:** The storage provisioner is configured to create/bind the PV only when its first consumer is created, and the operator's reconciler is waiting on a bind that won't happen until then.

**How to solve:**
1. In the deployment parameter configuration (CMDB), add `waitPvcBound: false` under both the `mongodb.storage` and `backup.storage` sections.
2. Retry the deployment job.

## Deploy Fails: `customDataRSParameters` / `otherDomainName` Invalid Value `"null"`

**Description:** Known issue where these two parameters default to `null` instead of a valid empty value.
```
spec.mongodb.customDataRSParameters: Invalid value: "null": ... must be of type array: "null"
spec.schemaSettings.otherDomainName: Invalid value: "null": ... must be of type string: "null"
```

**How to solve:** Set explicit values:
```
schemaSettings.otherDomainName: "cluster.local"
mongodb.customDataRSParameters: '["logLevel=0"]'
```

## Pods Not Starting: `context deadline exceeded` in Events

**Description:** This is a Docker-level issue on the node, not a MongoDB or operator issue.

**How to solve:** Restart the Docker daemon on the affected node (typically requires Ops/IT access).

## Pods Stuck in `Pending` State

**Description:** Seen on older, un-updated installations that use init-containers to manage security keys. If an exited init-container gets pruned from the node (e.g. by a cron cleanup task) before the pod fully settles, the whole pod can get stuck `Pending`.

**How to solve:**
- **Workaround:** delete the `Pending` pods one at a time, waiting for each replacement to reach `Running 1/1` before deleting the next.
- **Real fix:** upgrade to a newer release of the MongoDB component that no longer relies on this init-container pattern.

## Clean Deploy Fails: `command replSetInitiate requires authentication`

**Description:** The PVs specified in the deploy parameters already contain data from a previous installation, so `replSetInitiate` is running against a cluster that thinks it's already initialized and secured.

**How to solve:** Set `recycler.install: true` to clean the PVs before installation, or clean the PVs manually before retrying the deploy.
