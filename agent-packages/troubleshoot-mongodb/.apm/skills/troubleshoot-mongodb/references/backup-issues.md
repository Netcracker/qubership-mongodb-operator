# Backup Issues

## Backup Failed for Some Reason

**How to solve:**
1. Go to `rc/mongodb-backup-daemon` or `dc/mongodb-backup-daemon` for the failing installation.
2. Open `/backups/<failed_backup_folder>/.console` — the folder name looks like `20180717T020000` and is given in the error.
3. Search that file for a line starting with `Backup failed` — the actual error is near it.

## No Backup Information in Monitoring

**Description:** The `monitoring-agent` OpenShift service account lacks sufficient rights to view resources in the Mongo project. Logs show:
```
AttributeError: 'NoneType' object has no attribute 'get'
```
in `mongodb-monitoring-agent`, originating from a failed configmap list call.

**How to solve:**
```
service_account="monitoring-agent"
oc policy add-role-to-user view system:serviceaccount:${NAMESPACE}:${service_account} -n ${NAMESPACE}
```

## Random Pod Restarts During Backup

**Description:** The backup process can stress the filesystem enough to cause pod restarts — this shows up particularly when the backup pod uses GlusterFS and dumps 5GB+ of data at once, leading to I/O waits.

**How to solve:**
- Limit ingress traffic to the backup pod so the load is reduced (the backup itself will take longer as a trade-off): set the `BACKUP_INGRESS_BANDWIDTH` parameter to the desired value in the deploy job parameters.
