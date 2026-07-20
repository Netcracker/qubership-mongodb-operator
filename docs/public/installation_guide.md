The following topics are discussed in the chapter:

[[_TOC_]]

# Prerequisites

Following are the prerequisites that have to be satisfied before installing MongoDB.

## Common

* The deployer user (SA) must have the following Role bound:

 ```
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: nc-role
rules:
  - apiGroups:
      - netcracker.com
    resources:
      - '*'
    verbs:
      - create
      - get
      - list
      - patch
      - update
      - watch
      - delete
```

An example of role binding is as follows:

```
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
  name: nc-role-binding
  namespace: mongodb
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: nc-role
subjects:
- kind: ServiceAccount
  name: service-deployer
```

* The project or namespace should be created.
* The Custom Resource Definition (CRD) should be created by the cloud administrator.
* If Dynamic Volume Provisioning is not available, all the persistent volumes should be created manually.
* If pre-created PVs are used in OpenShift, the project must be annotated with the same UID that is used for the PV.
* If the deployment to OpenShift is done with restricted SCC, the project supplemental group annotation must have the same UID as in the **podSecurityContext.runAsUser** and **podSecurityContext.fsGroup** parameters.
* If the Pod Security Policy is enabled on the Kubernetes (K8s) cluster, it is mandatory to set the **podSecurityContext.fsGroup** and **podSecurityContext.runAsUser** parameters. For more information, refer to [https://kubernetes.io/docs/concepts/policy/pod-security-policy/](https://kubernetes.io/docs/concepts/policy/pod-security-policy/).
* In case of Prometheus Monitoring stack deployment, you should have the rights to create the **integreatly.org/v1alpha1** and **monitoring.coreos.com/v1** objects.

The following is an example of such a role:

```
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: generic-monitoring-role
rules:
  - apiGroups:
      - "monitoring.coreos.com"
    resources:
      - servicemonitors
      - prometheusrules
    verbs:
      - get
      - list
      - create
      - update
      - delete
      - watch
      - patch
  - apiGroups:
      - "integreatly.org"
    resources:
      - grafanadashboards
    verbs:
      - get
      - list
      - create
      - update
      - delete
      - watch
      - patch

```

An example of role binding is as follows:

```
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
  name: monitoring-role-binding
  namespace: mongodb
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: generic-monitoring-role
subjects:
- kind: ServiceAccount
  name: service-deployer
```

### CRD Upgrade

The upgrade of CRD happens automatically through pre-deploy scripts.

**Note**: An automatic CRD upgrade requires the corresponding role for the deploy user:

```yaml
- apiGroups: ["apiextensions.k8s.io"]
  resources: ["customresourcedefinitions"]
  verbs: ["get", "create", "patch"]
```

To disable this feature, add `DISABLE_CRD: true`.

### Apply New Custom Resource Definition Version

According to Helm restrictions, you have to apply CRD manually in the following two cases:

* If you need to add a new version of the `MongoService` custom resource.
* If you are upgrading between two minor or major versions. If letter x or y are changed in monitoring-operator version x.y.z. For example, if you are upgrading from 0.9.0 to 0.10.0.

You can find multiple CRDs in the `charts/helm/mongodb-operator/` directory:

* `legacy_crd/crd.yaml` - CRD for Kubernetes version below 1.22 and OpenShift below 4.9.
* `crds/k8s_1.22_crd.yaml` - CRD for Kubernetes version 1.22+ and OpenShift 4.9+.

If the CRD is created manually, apply the new version of CRD using the following command:

`kubectl apply -f charts/helm/mongodb-operator/crds/<crd_name>.yaml`

If the CRD is created by Helm, apply the new version of CRD using the following command:

`kubectl replace -f charts/helm/mongodb-operator/crds/<crd_name>.yaml`

Specify --skip-crds in the ADDITIONAL_OPTIONS parameter of the DP Deployer job.

Specify DISABLE_CRD=true; in the CUSTOM_PARAMS parameter of the App Deployer job.

# Best Practices and Recommendations

Specify here all the known recommendations applicable to generic deployments and deployments of NC products.
For example, information about the recommended PG max_connections, formulas for resources' calculation, and so on.

## HWE

MongoDB services resources can be selected during deployment using parameter `global.profile` that takes values: `small`, `medium`, `large`.

### Small

Following are the recommendations for development purposes, PoC, and demos.


| Module                   | CPU Requests | RAM Requests | CPU Limits | RAM Limits | Storage, Gb |
| ------------------------ | ------------ | ------------ | ---------- | ---------- | ----------- |
| operator                 | 50m          | 32Mi         | 100m       | 128Mi      | -           |
| mongodb.cnfResources     | 100m         | 128Mi        | 500m       | 512Mi      | -           |
| mongodb.dataResources    | 100m         | 256Mi        | 500m       | 512Mi      | 5Gi         |
| mongodb.arbiterResources | 100m         | 256Mi        | 200m       | 256Mi      | -           |
| mongodb.mongosResources  | 100m         | 256Mi        | 500m       | 512Mi      | -           |
| backup                   | 100m         | 256Mi        | 500m       | 512Mi      | 5Gi         |
| dbaas                    | 20m          | 32Mi         | 20m        | 64Mi       | -           |
| prometheusExporter       | 200m         | 128Mi        | 300m       | 256Mi      | -           |
| robotTests               | 200m         | 128Mi        | 200m       | 256Mi      | -           |
| Total                    | 1870m        | 2Gi          | 8          | 8Gi        | 20Gi        |

### Medium

Following are the recommendations for deployments with an average load.

| Module                   | CPU Requests | RAM Requests | CPU Limits | RAM Limits | Storage, Gb |
| ------------------------ | ------------ | ------------ | ---------- | ---------- | ----------- |
| operator                 | 50m          | 32Mi         | 100m       | 128Mi      | -           |
| mongodb.cnfResources     | 300m         | 128Mi        | 1          | 1Gi        | -           |
| mongodb.dataResources    | 500m         | 256Mi        | 2          | 2Gi        | 50Gi        |
| mongodb.arbiterResources | 100m         | 256Mi        | 200m       | 256Mi      | -           |
| mongodb.mongosResources  | 500m         | 256Mi        | 2          | 2Gi        | -           |
| backup                   | 500m         | 256Mi        | 1          | 1Gi        | 50Gi        |
| dbaas                    | 20m          | 32Mi         | 20m        | 64Mi       | -           |
| prometheusExporter       | 200m         | 128Mi        | 300m       | 256Mi      | -           |
| robotTests               | 200m         | 128Mi        | 200m       | 256Mi      | -           |
| Total                    | 7            | 2Gi          | 25         | 25Gi       | 200Gi       |

### Large

Following are the recommendations for deployments with a high workload and a large amount of data.

| Module                   | CPU Requests | RAM Requests | CPU Limits | RAM Limits | Storage, Gb |
| ------------------------ | ------------ | ------------ | ---------- | ---------- | ----------- |
| operator                 | 50m          | 32Mi         | 100m       | 128Mi      | -           |
| mongodb.cnfResources     | 300m         | 512Mi        | 2          | 2Gi        | -           |
| mongodb.dataResources    | 1            | 1Gi          | 4          | 4Gi        | 100Gi       |
| mongodb.arbiterResources | 100m         | 256Mi        | 200m       | 256Mi      | -           |
| mongodb.mongosResources  | 1            | 1Gi          | 4          | 4Gi        | -           |
| backup                   | 500m         | 512Mi        | 2          | 2Gi        | 100Gi       |
| dbaas                    | 20m          | 32Mi         | 20m        | 64Mi       | -           |
| prometheusExporter       | 200m         | 128Mi        | 300m       | 256Mi      | -           |
| robotTests               | 200m         | 128Mi        | 200m       | 256Mi      | -           |
| Total                    | 12.5         | 12.5         | 58         | 58Gi       | 400Gi       |

# Parameters

The [values.yaml](charts/helm/mongodb-operator/values.yaml) file contains the description and example of each parameter and its default value. Some parameters are self-explanatory.

The following sections provide the list of parameters.

## Operator Parameters

The list of Operator parameters is as follows:

| Parameter                                         | Mandatory | Type    | Default                           | Description                                                                                                                                                                                                                                                                                  |
| ------------------------------------------------- |-----------|---------|-----------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `operator.operatorName`                           | false     | string  | `operator-service`                | The name of the operator.                                                                                                                                                                                                                                                                    |
| `operator.image`                                  | false     | string  |                                   | The Docker image of the operator.                                                                                                                                                                                                                                                            |
| `operator.resources`                              | false     | object  | `{}`                              | The pod resource requests and limits.                                                                                                                                                                                                                                                        |
| `schemaSettings.schemaType`                       | false     | string  | `ha`                              | The type of schema to deploy. The possible values are `ha`, `single`, `dr`, and `arbiter`.                                                                                                                                                                                                   |
| `schemaSettings.cnfReplicaSize`                   | false     | int     | 3                                 | The number of CNFRS replicas.                                                                                                                                                                                                                                                                |
| `schemaSettings.dataReplicaSize`                  | false     | int     | 3                                 | The number of DATARS replicas.                                                                                                                                                                                                                                                               |
| `schemaSettings.mongosReplicas`                   | false     | int     | 1                                 | The number of Mongos replicas. Setting more than 1 replica can lead to potential problems refer to [Monos HA Schema](#mongos-ha-schema)                                                                                                                                                      |
| `schemaSettings.shardCount`                       | false     | int     | 3                                 | The number of shards. This parameter is ignored if `schemaSettings.sharded` is set to `false`.                                                                                                                                                                                               |
| `schemaSettings.arbiterIndex`                     | false     | int     | -1                                | The index of replica to be an arbiter. If not set, the arbiter index is the middle replica (node).                                                                                                                                                                                           |
| `schemaSettings.sharded`                          | false     | bool    | true                              | MongoDB needs to be deployed in sharded schema.                                                                                                                                                                                                                                              |
| `recycler.resources`                              | false     | object  | `{}`                              | The pod resource requests and limits.                                                                                                                                                                                                                                                        |
| `podSecurityContext.fsGroup`                      | false     | int     | 1001                              | The fsGroup of pods.                                                                                                                                                                                                                                                                         |
| `podSecurityContext.runAsUser`                    | false     | int     | 1001                              | The runAsUser of pods.                                                                                                                                                                                                                                                                       |
| `podSecurityContext.supplementalGroups`           | false     | string  | ""                                | The supplemental groups of pods.                                                                                                                                                                                                                                                             |
| `podSecurityContext.seLinuxOptions`                    | false     | object     | `{}`                              | The SELinux context labels to apply to the pod.                                                                                                                                                                                                                                                                       |
| `policies.tolerations[$idx].key`                  | false     | string  | ""                                | The taint key the toleration applies to.                                                                                                                                                                                                                                                     |
| `policies.tolerations[$idx].operator`             | false     | string  | ""                                | The key relationship to the value.                                                                                                                                                                                                                                                           |
| `policies.tolerations[$idx].value`                | false     | string  | ""                                | The taint value the toleration matches to.                                                                                                                                                                                                                                                   |
| `policies.tolerations[$idx].effect`               | false     | string  | ""                                | The taint effect to match.                                                                                                                                                                                                                                                                   |
| `policies.tolerations[$idx].tolerationSeconds`    | false     | int     | -                                 | The period the toleration (which must be of effect NoExecute, otherwise this field is ignored) tolerates the taint.                                                                                                                                                                          |
| `deletePVConUninstall`                            | false     | bool    | false       	                     | If PVCs needs to be deleted when `helm uninstall` executed.                                                                                                                                                                                                                                  |
| `waitSeconds`                                     | false     | int     | 300                               | The timeout of the main operation during the deployment.                                                                                                                                                                                                                                     |
| `drTimeoutSeconds`                                | false     | int     | 360                               | The timeout of DR operations.                                                                                                                                                                                                                                                                |
| `imagePullPolicy`                                 | false     | string  | `IfNotPresent`                    | If the image should be pulled prior to starting the container. Values: `Always` - always pull the image; `IfNotPresent` - only pull the image if it does not already exist on the node; `Never` - never pull the image.                                                                      |
| `debugLog`                                        | false     | bool    | false                             | The debug level of the operator log.                                                                                                                                                                                                                                                         |
| `mongodbOperatorWaitDelaySeconds`                 | false     | int     | 30                                | The delay in seconds before the services operator checks if the MongoDB operator CR is ready. Provides a time buffer for the MongoDB operator pod to start and update the CR status when both operators are deployed simultaneously. |
| `authDb`                                          | false     | bool    | `admin`                           | The mongo auth DB.                                                                                                                                                                                                                                                                           |
| `ipV6`                                            | false     | bool    | false                             | If ipV6 should be used.                                                                                                                                                                                                                                                                      |
| `role.create`                                     | false     | bool    | `yes`                             | If an operator role should be created.                                                                                                                                                                                                                                                       |
| `roleBinding.create`                              | false     | bool    | `yes`                             | If the operator role binding should be created.                                                                                                                                                                                                                                              |
| `serviceAccount.create`                           | false     | bool    | `yes`                             | If the operator service account should be created.                                                                                                                                                                                                                                           |
| `disasterRecovery.image`                          | false     | string  |                                   | The Disaster Recovery Mongo Service operator container image.                                                                                                                                                                                                                                |
| `disasterRecovery.httpAuth.enabled`               | false     | bool    | `false`                           | Authentication should be enabled or not.                                                                                                                                                                                                                                                     |
| `disasterRecovery.httpAuth.smNamespace`           | false     | string  | `site-manager`                    | The name of Kubernetes Namespace from which the site manager API calls are done.                                                                                                                                                                                                             |
| `disasterRecovery.httpAuth.smSecureAuth`          | false     | boolean | false                             | Whether the `smSecureAuth` mode is enabled for Site Manager or not.                                                                                                                                                                                                                          |
| `disasterRecovery.httpAuth.smServiceAccountName`  | false     | string  | `sm-auth-sa` or `site-manager-sa` | The name of Kubernetes Service Account under which the site manager API calls are done. Default values depend on the `smSecureAuth` mode.                                                                                                                                                    |
| `disasterRecovery.httpAuth.customAudience`        | false     | string  | `sm-services`                     | The name of custom audience for rest api token, that is used to connect with services. It is necessary if Site Manager installed with `smSecureAuth=true` and has applied custom audience (`sm-services` by default). It is considered if `disasterRecovery.httpAuth.smSecureAuth` parameter is set to `true` |
| `disasterRecovery.httpAuth.restrictedEnvironment` | false     | bool    | `false`                           | If the parameter is `true`, the `system:auth-delegator` cluster role is bound to the Mongo Service operator service account. The cluster role is not bound if the disaster recovery mode is disabled, or the disaster recovery server authentication is disabled.                            |
| `disasterRecovery.mode`                           | Mandatory | Type    | Default                           | The current side is active during service installation in the Disaster Recovery mode. If you do not specify this parameter, the service is deployed in a regular mode, and not in the Disaster Recovery mode. For more information, see [Disaster Recovery Modes](#disaster-recovery-modes). |
| `disasterRecovery.afterServices`                  | false     | array   | `[]`                              | The list of `SiteManager` names for services after which the Mongo service switchover is to be run.                                                                                                                                                                                          |

### Disaster Recovery Modes

| DR Mode    | Description                                                                                                                |
| ---------- | -------------------------------------------------------------------------------------------------------------------------- |
| `active`   | The mode in which MongoDB accepts external requests from clients.                                                           |
| `standby`  | The mode in which MongoDB does not accept external requests from clients and the replication from `active` MongoDB is enabled. |
| `disabled` | The mode in which MongoDB does not accept external requests from clients and the replication from `active` MongoDB is disabled. |

**Important**: If MongoDB replication is not required, the `disasterRecovery` section must be empty.

**Note**: You need to set this parameter during the primary initialization through `clean install` or `reinstall`. Do not change it with the `upgrade` process. To change the mode, use the `SiteManager` functionality or the MongoDB disaster recovery REST server API.

## MongoDB Parameters

The list of MongoDB parameters is as follows:

| Parameter                                                                    | Mandatory | Type           | Default              | Description                                                                                                                                                                                                                   |
| ---------------------------------------------------------------------------- | --------- | -------------- | -------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `mongodb.install`                                                            | true      | bool           | `true`               | If MongoDB needs to be installed.                                                                                                                                                                                              |
| `mongodb.clusterAuthMode`                                                            | false      | string           | `keyfile`               | Specifies the authentication mode used between MongoDB cluster members. Supported values: `keyfile`, `x509` (described in the [Cluster Auth Mode](#cluster-auth-modes) section).                                                                                                                                                                                              |
| `mongodb.flavor`                              | false     | string            | small   | The flavor of Mongodb deployment resources. Possible values are  `small`, `medium`, `large`.           |
| `mongodb.dockerImage`                                                        | false     | string         | ``                   | The Docker image of MongoDB.                                                                                                                                                                                                   |
| `mongodb.additionalNodeLabels`                                               | false     | object         | `{}`                 | The additional node labels for each mongo replica.                                                                                                                                                                             |
| `mongodb.rootUser`                                                           | false     | string         | `root`               | The root user of MongoDB.                                                                                                                                                                                                      |
| `mongodb.rootPassword`                                                       | false     | string         | `root`               | The root password of MongoDB.                                                                                                                                                                                                  |
| `mongodb.rootUserRole`                                                       | false     | string         | `'root', '__system'` | The root user role in MongoDB.                                                                                                                                                                                                 |
| `mongodb.cnfWiredTigerCacheGb`                                               | false     | string         | `0.12`               | The MongoDB Wired Tiger cache in GB for CNFRS replica.                                                                                                                                                                         |
| `mongodb.dataWiredTigerCacheGb`                                              | false     | string         | `0.25`               | The MongoDB Wired Tiger cache in GB for DATARS replica.                                                                                                                                                                        |
| `mongodb.cnfOpLogSizeMb`                                                     | false      |       int | [Default Size](https://www.mongodb.com/docs/manual/core/replica-set-oplog/#oplog-size) | The MongoDB OpLog size in MB for CNFRS replica. |
| `mongodb.dataOpLogSizeMb`                                                    | false |        int | [Default Size](https://www.mongodb.com/docs/manual/core/replica-set-oplog/#oplog-size) | The MongoDB OpLog size in MB for DATARS replica. |
| `mongodb.customDataRSParameters`                                             | false     | list of string | `[]`                 | The custom parameters for DATARS replica.                                                                                                                                                                                      |
| `mongodb.cnfResources.limits.memory`                                         | true      | Quantity       | `512Mi`              | The memory limit of CNFRS replica.                                                                                                                                                                                             |
| `mongodb.cnfResources.limits.cpu`                                            | true      | Quantity       | `300m`               | The CPU limit of CNFRS replica.                                                                                                                                                                                                |
| `mongodb.cnfResources.requests.memory`                                       | true      | Quantity       | `128Mi`              | The minimum amount of memory for CNFRS replica.                                                                                                                                                                                |
| `mongodb.cnfResources.requests.cpu`                                          | true      | Quantity       | `100m`               | The minimum number of CPUs for CNFRS replica.                                                                                                                                                                                  |
| `mongodb.dataResources.limits.memory`                                        | true      | Quantity       | `512Mi`              | The maximum amount of memory for DATARS replica.                                                                                                                                                                               |
| `mongodb.dataResources.limits.cpu`                                           | true      | Quantity       | `500m`               | The maximum number of CPUs for DATARS replica.                                                                                                                                                                                 |
| `mongodb.dataResources.requests.memory`                                      | true      | Quantity       | `256Mi`              | The memory request of DATARS replica.                                                                                                                                                                                          |
| `mongodb.dataResources.requests.cpu`                                         | true      | Quantity       | `100m`               | The minimum number of CPUs for DATARS replica.                                                                                                                                                                                 |
| `mongodb.arbiterResources.limits.memory`                                     | false     | Quantity       | `256Mi`              | The maximum amount of memory for Arbiter replica.                                                                                                                                                                              |
| `mongodb.arbiterResources.limits.cpu`                                        | false     | Quantity       | `200m`               | The maximum number of CPUs for Arbiter replica.                                                                                                                                                                                |
| `mongodb.arbiterResources.requests.memory`                                   | false     | Quantity       | `128Mi`              | The memory request of Arbiter replica.                                                                                                                                                                                         |
| `mongodb.arbiterResources.requests.cpu`                                      | false     | Quantity       | `100m`               | The minimum number of CPUs for Arbiter replica.                                                                                                                                                                                |
| `mongodb.mongosResources.limits.memory`                                      | false     | Quantity       | `512Mi`              | The memory limit of mongos replica.                                                                                                                                                                                           |
| `mongodb.mongosResources.limits.cpu`                                         | false     | Quantity       | `500m`               | The CPU limit of mongos replica.                                                                                                                                                                                              |
| `mongodb.mongosResources.requests.memory`                                    | false     | Quantity       | `256Mi`              | The maximum amount of memory for mongos replica.                                                                                                                                                                              |
| `mongodb.mongosResources.requests.cpu`                                       | false     | Quantity       | `100m`               | The minimum number of CPUs for mongos replica.                                                                                                                                                                                |
| `mongodb.containerTimeoutSeconds`                                            | false     | int            | 10                    | The parameter that allows to override the default `timeoutSeconds` value of the readiness probe.                                                                                                                                                     |
| `mongodb.containerPeriodSeconds`                                             | false     | int            | 10                   | The parameter that allows to override the default `periodSeconds` value of the readiness probe.                                                                                                                                            |
| `mongodb.singleWiredTigerCacheGb`                                            | false     | float          | 0.12                 | The MongoDB Wired Tiger cache in GB for mongos replica.                                                                                                                                                                       |                                                                                                                                                                                                                   |
| `mongodb.storage.waitPVCBound`                                               | true      | bool           | false                | The parameter specifies if the operator needs to wait for all PVCs to bind before creating pods.                                                                                                                                                                                                                             |
| `mongodb.storage.size`                                                       | false     | Quantity       | `5Gi`                | The size of storage for MongoDB DATARS replica. It can be a list of sizes that should be aligned with `mongodb.storage.volumes` or a single value as a string or as one item of a list to make all volumes of the same size. |
| `mongodb.storage.storageClasses`                                             | true      | list of string | `[]`                 | The list of storage classes for MongoDB.|
| `mongodb.storage.nodeLabels`                                                 | false     |                |                      | The labels to map the pods to the nodes. To set the node name, use the `kubernetes.io/hostname` label.                                                                                                                                |
| `mongodb.storage.volumes`                                                    | false     | list           |                      | The list of PV for MongoDB. For auto-provision, leave the value blank for this parameter.                                                                                                                                     |
| `mongodb.storage.storageClasses`                                             | true      | list of string | `- ""`               | The list of storage classes for MongoDB.                                                                                                                                                                                      |
| `mongodb.storage.matchLabelSelectors`                                        | false     | `{}`           | {}                   | The `key:value` pair of PVs to bind to MongoDB. PVCs.                                                                                                                                                                          |

## TLS Parameters

The list of TLS Parameters is as follows:

| Parameter                                                        | Mandatory | Type   | Default            | Description                                                                                                          |
| ---------------------------------------------------------------- | --------- | ------ | ------------------ | -------------------------------------------------------------------------------------------------------------------- |
| `tls.mode`                                                       | false     | string | `disabled`         | The parameter that enables TLS used for all network connections (all modes are described in the [TLS Encryption](#tls-encryption) section). |
| `tls.certificateSecretName`                                      | false     | string | `root-ca`          | The name of the secret where the certificates for the connection are stored.                                         |
| `tls.rootCAFileName`                                             | false     | string | `ca.crt`           | The name of the CA root certificate file.                                                                                |
| `tls.signedCRTFileName`                                          | false     | string | `tls.crt`          | The name of the file that contains the TLS certificate and key.                                                              |
| `tls.privateKeyFileName`                                         | false     | Type   | `tls.key`          | The name of the file that contains the TLS private key.                                                                       |
| `tls.combinedKeyAndCRTFileName`                                  | false     | Type   | `tls-combined.pem` | The name of the file that contains the key and certificate.                                                                   |
| `tls.generateCerts.enabled`                                      | false     | bool   | `false`            | The parameter that specifies if certificates must be obtained from cert-manager, otherwise self-signed certificates are generated.         |
| `tls.generateCerts.clusterIssuerName`                            | false     | string | ""                 | The name of ClusterIssuer to integrate with Cert Manager. It must be set if the integration with Cert Manager is enabled.  |
| `tls.generateCerts.subjectAlternativeName.additionalDnsNames`    | false     | array  | `[]`               | The list of additional DNS names.                                                                                    |
| `tls.generateCerts.subjectAlternativeName.additionalIpAddresses` | false     | array  | `[]`               | The list of additional IP addresses.                                                                                 |

## Backup Daemon Parameters

The list of Backup daemon parameters are as follows:

| Parameter                                                                   | Mandatory | Type           | Default                      | Description                                                                                                                                                                      |
| --------------------------------------------------------------------------- | --------- | -------------- | ---------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `backup.install`                                                            | false     | bool           | `true`                       | If backup daemon needs to be installed.                                                                                                                                          |
| `backup.storageDirectory`                                                   | false     | string         | `/backup-storage`            | The directory inside of Mongo Backup Daemon to mount to PV. Override this parameter value to set subfolder if S3 storage is used. For example, `/mongo_subfolder/backup-storage`. |
| `backup.s3.enabled`                                                         | false     | bool           | `false`                      | If S3 storage should be used.                                                                                                                                                    |
| `backup.s3.bucketName`                                                      | false     | string         | ""                           | The S3 bucket to store backups.                                                                                                                                                      |
| `backup.s3.endpointUrl`                                                     | false     | string         | ""                           | The S3 URL.                                                                                                                                                                          |
| `backup.s3.accessKeyId`                                                     | false     | string         | ""                           | The S3 access key ID.                                                                                                                                                                |
| `backup.s3.accessKeySecret`                                                 | false     | string         | ""                           | The S3 access key secret.                                                                                                                                                            |
| `backup.dockerImage`                                                        | false     | string         | ``                           | The Docker image of MongoDB Backup Daemon.                                                                                                                                       |
| `backup.additionalNodeLabels`                                               | false     | object         | `{}`                         | The additional node labels for backup-daemon replica.                                                                                                                            |
| `backup.backupSchedule`                                                     | false     | string         | `0 * * * *`                  | The schedule for automatic periodic full backups in cron syntax.                                                                                                                 |
| `backup.incBackupSchedule`                                                  | false     | string         | ""                           | The schedule for automatic periodic incremental backups.                                                                                                                         |
| `backup.evictionPolicy`                                                     | false     | string         | `0/1h,3d/7d,1m/1m,1y/delete` | The eviction policy for full backups.                                                                                                                                            |
| `backup.incEvicitionPolicy`                                                 | false     | string         | ""                           | The eviction policy for incremental backups.                                                                                                                                     |
| `backup.mongoDatabasePrefixPattern`                                         | false     | string         | `.*`                         | The DB prefix for incremental backup.                                                                                                                                            |
| `backup.backupUser`                                                         | false     | string         | `backup`                     | The MongoDB user with 'backup' privileges.                                                                                                                                       |
| `backup.backupPassword`                                                     | false     | string         | `backup`                     | The password for MongoDB user with 'backup' privileges.                                                                                                                          |
| `backup.backupUserRole`                                                     | false     | string         | `"'backup'"`                 | The role of user with 'backup' privileges.                                                                                                                                       |
| `backup.restoreUser`                                                        | false     | string         | `restore`                    | The MongoDB user with 'create\drop database' privileges.                                                                                                                         |
| `backup.restorePassword`                                                    | false     | string         | `restore`                    | The password for MongoDB user with 'create\drop database' privileges.                                                                                                            |
| `backup.restoreUserRole`                                                    | false     | string         | `"'restore'"`                | The role of user with 'create\drop database' privileges.                                                                                                                         |
| `backup.apiUser`                                                            | false     | string         | `backup`                     | The user for REST API of backup daemon. To make API available without auth, set empty value.                                                                                     |
| `backup.apiPassword`                                                        | false     | string         | `backup`                     | The password of user for REST API of backup daemon. To make API available without auth, set empty value.                                                                         |
| `backup.storage.nodeLabels`                                                 | false     | object         | `{}`                         | The labels to map the pod to the node. To set the node name, use the `kubernetes.io/hostname` label.                                                                                     |
| `backup.storage.size`                                                       | true      | Quantity       | `5Gi`                        | The size of storage for the backup daemon DATARS replica.                                                                                                                            |
| `backup.storage.volumes`                                                    | false     | Type           |                              | The list of PV for backup daemon. For auto-provision, leave the value blank for this parameter.                                                                                  |
| `backup.storage.storageClasses`                                             | true      | string         | ""                           | The storage class for backup daemon. For hostpath PVs, leave the value blank for this parameter.                                                                                 |
| `backup.storage.matchLabelSelectors`                                        | false     | Type           |                              | The `key:value` pair of PVs to bind to backup daemon PVCs.                                                                                                                       |
| `backup.backupResources.limits.memory`                                      | false     | Quantity       | `512Mi`                      | The maximum amount of memory for backup-daemon replica.                                                                                                                          |
| `backup.backupResources.limits.cpu`                                         | false     | Quantity       | `500m`                       | The maximum number of CPUs for backup-daemon replica.                                                                                                                            |
| `backup.backupResources.requests.memory`                                    | false     | Quantity       | `256Mi`                      | The minimum amount of memory for backup-daemon replica.                                                                                                                          |
| `backup.backupResources.requests.cpu`                                       | false     | Quantity       | `100m`                       | The minimum number of CPUs for backup-daemon replica.                                                                                                                            |
| `backup.numParallelConnections`                                             | false     | int            | 4                            | The number of parallel connections during mongo restore.                                                                                                                         |
| `backup.granularNumParallelConnections`                                     | false     | int            | 4                            | The number of parallel connections during granular mongo restore.                                                                                                                |
| `backup.mongoBackupDB`                                                      | false     | string         | `mongos`                     | The host of backup dump.                                                                                                                                                         |
| `backup.mongoSourceDB`                                                      | false     | string         | `mongos`                     | The host of restore.                                                                                                                                                             |
| `backup.enableFullRestore`                                                  | false     | bool           | `false`                      | If full restore should be enabled.                                                                                                                                               |
| `backup.configCollections`                                                  | false     | list of string | `[]`                         | The collections from the config DB that should be dumped.                                                                                                                        |
| `backup.tls.backupDaemonCASecretName`                                       | false     | string         | `backup-daemon-certificate`  | The name of the secret where the certificate is stored.                                                                                                                          |
| `backup.tls.duration`                                                       | false     | int            | 365                          | The certificate validity period.|
| `backup.tls.subjectAlternativeName.additionalDnsNames`                      | false     | list of string | [ ]                          | The list of additional DNS names.                                                                                                                                                |
| `backup.tls.subjectAlternativeName.additionalIpAddresses`                   | false     | list of string | [ ]                          | The list of additional IP addresses.                                                                                                                                             |

## DBaaS Adapter Parameters

The list of DBaaS Adapter parameters is as follows:

| Parameter                                                | Mandatory | Type           | Default                                                                                        | Description                                                           |
| -------------------------------------------------------- | --------- | -------------- | ---------------------------------------------------------------------------------------------- | --------------------------------------------------------------------- |
| `dbaas.dockerImage`                                      | false     | string         | ``                                                                                             | The Docker image of the DBaaS adapter.                                    |
| `dbaas.install`                                          | false     | bool           | `true`                                                                                         | If the DBaaS adapter needs to be installed.                               |
| `dbaas.apiVersion`                                       | false     | string         | `v1`                                                                                           | The DBaaS adapter REST API version.                                       |
| `dbaas.multiUsers`                                       | false     | bool           | `false`                                                                                        | If the DBaaS adapter needs to create multiple users on Create DB request. |
| `dbaas.additionalNodeLabels`                             | false     | object         |                                                                                                | The additional node labels for the DBaaS adapter replica.                 |
| `dbaas.dbaasUser`                                        | false     | string         | `dbaas`                                                                                        | The MongoDB user for DBaaS.                                           |
| `dbaas.dbaasPassword`                                    | false     | string         | `dbaas`                                                                                        | The password of MongoDB user for DBaaS.                               |
| `dbaas.dbaasUserRole`                                    | false     | string         | `'userAdminAnyDatabase', 'readAnyDatabase', 'clusterMonitor'`                                  | The role of a MongoDB user for DBaaS.                                   |
| `dbaas.dbaasAggregatorUser`                              | false     | string         | `dbaas-aggregator`                                                                             | The REST API user of DBaaS adapter.                                   |
| `dbaas.dbaasAggregatorPassword`                          | false     | string         | `dbaas-aggregator`                                                                             | The password of REST API user of DBaaS adapter.                       |
| `dbaas.dbaasAggregatorRegistrationAddress`               | false     | string         | `http://dbaas-aggregator.dbaas:8080`                                                           | The address of DBaaS aggregator.                                      |
| `dbaas.dbaasAggregatorRegistrationUser`                  | false     | string         | `cluster-dba`                                                                                  | The user for registration in the DBaaS aggregator.                        |
| `dbaas.dbaasAggregatorRegistrationPassword`              | false     | string         | `Bnmq5567_PO`                                                                                  | The user password for registration in the DBaaS aggregator.               |
| `dbaas.dbaasAggregatorRegistrationFixedDelayMS`          | false     | int            | 150000                                                                                         | The delay for registration in the DBaaS aggregator.                       |
| `dbaas.dbaasAggregatorRegistrationRetryDelayMS`          | false     | int            | 5000                                                                                           | The retry delay for registration in the DBaaS aggregator.                 |
| `dbaas.dbaasAggregatorRegistrationRetryTimeMS`           | false     | int            | 60000                                                                                          | The number of retries for registration in the DBaaS aggregator.           |
| `dbaas.dbaasAggregatorPhysicalDatabaseIdentifier`        | false     | string         | ""                                                                                             | The database identifier in the DBaaS aggregator.                          |
| `dbaas.dbaasPhysicalDatabasesCustomLabels`               | false     | object         | `{}`                                                                                           | The custom labels for registration in the DBaaS aggregator.               |
| `dbaas.dbaasStreamingRoleName`                           | false     | string         | `streaming`                                                                                    | The streaming user role name. |
| `dbaas.dbaasStreamingRoles`                              | false     | string         | `{ role: 'read', db: 'admin' }, { role: 'read', db: 'config' }, { role: 'read', db: 'local' }` | The streaming user roles.|
| `dbaas.dbaasStreamingPrivileges`                         | false     | string         | `{ resource: { db: '', collection: '' }, actions: ['find', 'changeStream'] }`                  | The streaming user priviliges.|
| `dbaas.dbaasResources.limits.memory`                     | false     | Quantity       | `64Mi`                                                                                         | The memory limit of the DBaaS agent replica.|
| `dbaas.dbaasResources.limits.cpu`                        | false     | Quantity       | `20m`                                                                                          | The CPU limit of the DBaaS agent replica.|
| `dbaas.dbaasResources.requests.memory`                   | false     | Quantity       | `32Mi`                                                                                         | The memory request of the DBaaS agent replica.|
| `dbaas.dbaasResources.requests.cpu`                      | false     | Quantity       | `20m`                                                                                          | The CPU request of the DBaaS agent replica.|
| `dbaas.tls.dbaasAdapterCASecretName`                     | false     | string         | `dbaas-adapter-certificate`                                                                    | The name of the secret where the certificate is stored.|
| `dbaas.tls.duration`                                     | false     | int            | 365 | The certificate validity period.|
| `dbaas.tls.subjectAlternativeName.additionalDnsNames`    | false     | list of string | [ ]                                                                                            | The list of additional DNS names.|
| `dbaas.tls.subjectAlternativeName.additionalIpAddresses` | false     | list of string | [ ]                                                                                            | The list of additional IP addresses.|
| `dbaas.gatewayAPI.gatewayName`                        | false     | string | default-external-gateway           | Chart default for Gateway name.                   |
| `dbaas.gatewayAPI.gatewayNamespace`                   | false     | string | gateway-system                     | Chart default for Gateway namespace.            |

## Prometheus Exporter Parameters

The list of Prometheus Exporter parameters is as follows:

| Parameter                                              | Mandatory | Type     | Default                                           | Description                                                   |
| ------------------------------------------------------ | --------- | -------- | ------------------------------------------------- | ------------------------------------------------------------- |
| `prometheusExporter.install`                           | false     | bool     | `true`                                            | If the Prometheus exporter needs to be installed.             |
| `prometheusExporter.dockerImage`                       | false     | string   | ``                                                | The Docker image of the Prometheus exporter.                      |
| `prometheusExporter.additionalNodeLabels`              | false     | object   | `{}`                                              | The additional node labels for the Prometheus exporter replica.   |
| `prometheusExporter.monitoringUser`                    | false     | string   | `monitoring`                                      | The MongoDB user for the Prometheus exporter.                     |
| `prometheusExporter.monitoringPassword`                | false     | string   | `monitoring`                                      | The password of MongoDB user for the Prometheus exporter.         |
| `prometheusExporter.monitoringUserRole`                | false     | string   | `{role:'readWrite', db:'test'}, 'clusterMonitor'` | The role of MongoDB user for the Prometheus exporter.             |
| `prometheusExporter.exporterUser`                      | false     | string   | `admin`                                           | The API user of the Prometheus exporter.                          |
| `prometheusExporter.exporterPassword`                  | false     | string   | `admin`                                           | The API user password of the Prometheus exporter.             |
| `prometheusExporter.collectionInterval`                | false     | int      | 30                                                | The interval at which metrics should be collected.            |
| `prometheusExporter.collectionScrapeTimeout`           | false     | int      | 30                                                | The timeout in seconds after which the scrape is ended.           |
| `prometheusExporter.exporterResources.limits.memory`   | false     | Quantity | `256Mi`                                           | The maximum amount of memory for the Prometheus exporter replica. |
| `prometheusExporter.exporterResources.limits.cpu`      | false     | Quantity | `300m`                                            | The maximum number of CPUs for the Prometheus exporter replica.   |
| `prometheusExporter.exporterResources.requests.memory` | false     | Quantity | `128Mi`                                           | The minimum amount of memory for the Prometheus exporter replica. |
| `prometheusExporter.exporterResources.requests.cpu`    | false     | Quantity | `200m`                                            | The minimum number of CPUs for the Prometheus exporter replica.   |
| `prometheusExporter.alerts.common.podRestartCount`     | false     | int      | `0`                                              | Threshold for pod restart count.                               |
| `prometheusExporter.alerts.common.podRestartPeriod`    | false     | string   | `6m`                                             | Period to monitor pod restarts.                                |
| `prometheusExporter.alerts.common.replicationLagSeconds` | false   | int      | `10`                                             | Threshold for replication lag in seconds.                     |
| `prometheusExporter.alerts.common.cursorTimeoutThreshold` | false  | int      | `100`                                            | Threshold for number of cursors timing out in 10 minutes.     |
| `prometheusExporter.alerts.common.cpuUsageThreshold`   | false     | int      | `95`                                             | CPU usage threshold in percent.                                |
| `prometheusExporter.alerts.common.memoryUsageThreshold` | false    | int      | `95`                                             | Memory usage threshold in percent.                             |
| `prometheusExporter.alerts.common.balancerDisabledFor` | false     | string   | `2m`                                             | Duration to wait before firing balancer disabled alert.       |
| `prometheusExporter.alerts.common.chunkMigrationThreshold` | false  | int      | `0`                                              | Threshold for failed chunk migrations.                         |
| `prometheusExporter.alerts.common.chunkUnevenThreshold` | false    | int      | `30`                                             | Stddev/avg % for uneven chunk distribution.                   |
| `prometheusExporter.alerts.common.chunkUnevenFor`     | false     | string   | `10m`                                            | Duration to wait before firing uneven chunk alert.            |
| `prometheusExporter.alerts.common.connectionsThreshold` | false    | int      | `90`                                             | Threshold % of connections used.                               |
| `prometheusExporter.alerts.common.connectionsFor`     | false     | string   | `2m`                                             | Duration to wait before firing too many connections alert.    |
| `prometheusExporter.alerts.common.cachePressureThreshold` | false  | int      | `85`                                             | WiredTiger cache pressure threshold %.                         |
| `prometheusExporter.alerts.common.cachePressureFor`   | false     | string   | `3m`                                             | Duration to wait before firing WiredTiger cache pressure alert. |
| `prometheusExporter.alerts.backup.usedSpaceThreshold` | false     | int      | `80`                                             | Backup used space threshold %.                                 |
| `prometheusExporter.alerts.backup.usedInodesThreshold` | false    | int      | `80`                                             | Backup used inodes threshold %.                                |

## Robot Tests Parameters

The list of Vault Registration parameters is as follows:

| Parameter                              | Mandatory | Type     | Default | Description                                                                                                              |
| -------------------------------------- | --------- | -------- | ------- | ------------------------------------------------------------------------------------------------------------------------ |
| `robotTests.install`                   | false     | bool     | false   | If Robot tests needs to be installed.                                                                                    |
| `robotTests.tags`                      | false     | string   | `smoke` | The tags of Robot tests. The possible values are: `smoke`, `ha`, `backup`, `dbaas`, `dr`.                                    |
| `robotTests.externalBackupPath`        | false     | string   | ""      | The path to the external NFS storage folder.                                                                                     |
| `robotTests.nodeLabels`                | false     | object   | `{}`    | The additional node labels for Robot tests replica.                                                                      |
| `robotTests.dockerImage`               | false     | string   | ``      | The Docker image of Robot tests.                                                                                         |
| `robotTests.resources.requests.cpu`    | false     | Quantity | `200m`  | The CPU request of the Robot tests replica.                                                                                  |
| `robotTests.resources.requests.memory` | false     | Quantity | `128Mi` | The memory request of the Robot tests replica.                                                                               |
| `robotTests.resources.limits.cpu`      | false     | Quantity | `200m`  | The CPU limit of the Robot tests replica.                                                                                    |
| `robotTests.resources.limits.memory`   | false     | Quantity | `256Mi` | The memory limit of the Robot tests replica.                                                                                 |
| `robotTests.mainSide`                  | false     | string   | `left`  | The datacenter where operating instances of MongoDB are located for DR test cases. The possible values are: `left`, `right`. |
| `robotTests.leftNodesPattern`          | false     | string   | `left`  | The nodes' name pattern that is used to find nodes on the left side for DR test cases.                                       |
| `robotTests.rightNodesPattern`         | false     | string   | `right` | The nodes' name pattern that is used to find nodes on the right side for DR test cases. The default values is `right`.       |

## Parameters' Examples

The parameter' examples for the different scenarios of deployment are given below.


### Different Sizes for MongoDB Volumes

An example of different sizes for MongoDB volumes is as follows:

```
mongodb:
  install: true

  storage:
    size:
      - 5Gi
      - 6Gi
      - 4Gi
    nodeLabels:
      - "kubernetes.io/hostname": node-1
      - "kubernetes.io/hostname": node-2
      - "kubernetes.io/hostname": node-3
    volumes:
      - mongo-1
      - mongo-2
      - mongo-3
```

### Arbiter with Unique Resources

By default, if `mongodb.arbiterResources` is not defined, then the resources for its node are the same as for `mongodb.dataResources`. This is useful to save environment resources as the Arbiter node does not require much CPU and memory.
The recommended parameters are as follows:

```
schemaSettings:
  schemaType: "arbiter"

mongodb:
  arbiterResources:
    limits:
      memory: 128Mi
      cpu: 150m
    requests:
      memory: 128Mi
      cpu: 100m

```

### PVs Selected by Labels

An example of PVs selected by labels is as follows:

```


mongodb:
  install: true

  storage:
    size: 5Gi
    nodeLabels:
      - "kubernetes.io/hostname": node-1
      - "kubernetes.io/hostname": node-2
      - "kubernetes.io/hostname": node-3
      - "kubernetes.io/hostname": node-4
      - "kubernetes.io/hostname": node-5
    matchLabelSelectors:
      - first-selector: first-value
        first-additional-selector: first-additional-value
      - second-selector: second-value
      - third-selector: third-value
      - fourth-selector: fourth-value
      - fifth-selector: fifth-value
```

### Mongo Backup Daemon API Without Auth

An example of Mongo Backup Daemon API without auth is as follows:

```
backup:
  install: true
  apiUser:
  apiPassword:
  storage:
    size: 1Gi
    storageClasses:
      - local-path
```

## On-Prem

### Simple (Non-Sharded)

It is required to set the `schemaSettings.sharded` parameter value to `false`.

Example of deployment parameters:

```
mongodb:
  schemaSettings:
    sharded: false
  install: true
  storage:
    size: 5Gi
    storageClasses:
      - storage-class

backup:
  install: true
  storage:
    size: 1Gi
    storageClasses:
      - storage-class
```

### HA Schema

```
mongodb:
  install: true
  storage:
    size: 5Gi
    nodeLabels:
      - "kubernetes.io/hostname": node-1
      - "kubernetes.io/hostname": node-2
      - "kubernetes.io/hostname": node-3
    volumes:
      - mongo-1
      - mongo-2
      - mongo-3

backup:
  install: true
  storage:
    size: 1Gi
    storageClasses:
      - storage-class
```

### Mongos HA Schema

Note: more that one mongos replca can lead to the following problems:
- If mutiple pods write and read simultaneously the same database they can access not latest data
- Session affinity parameter on Mongos service has timeout of 3 hours, which means that application can be redirected to another mongos if this timeout is reached.
```
mongodb:
  install: true
  storage:
    size: 5Gi
    nodeLabels:
      - "kubernetes.io/hostname": node-1
      - "kubernetes.io/hostname": node-2
      - "kubernetes.io/hostname": node-3
    volumes:
      - mongo-1
      - mongo-2
      - mongo-3

backup:
  install: true
  storage:
    size: 1Gi
    storageClasses:
      - storage-class
schemaSettings:
  mongosReplicas: 2
```

### Non-HA Schema

**Note**: For development purposes only.

An example of Non-HA schema is as follows:

```
schemaSettings:
  schemaType: "single"
mongodb:
  install: true
  storage:
    size: 5Gi
    storageClasses:
      - storage-class
```

### DR Schema

This section describes the deployment of MongoDB on two separate Kubernetes instances with a Pod-to-Pod IP connectivity between them.

Key notes:

* BGP and Calico CNI plugin must be configured to provide a Pod-to-Pod connectivity between Kubernetes clusters.
* Separate deployment job must be run consecutively.

To deploy MongoDB pods with Calico:

1. Run the deployment job.

   * The `schemaSettings.schemaType` parameter must be set to `dr`.
   * The `schemaSettings.schemaType.thisDomainName` parameter must be set to the DNS name of the STANDBY Kubernetes instance.
   * The `schemaSettings.schemaType.otherDomainName` parameter must be set to the DNS name of the ACTIVE Kubernetes instance.
   * The `schemaSettings.mode` parameter must be set to `standby`.

1. Wait for the job to finish successfully.

1. Run the deployment job:

   * The `schemaSettings.schemaType` parameter must be set to `dr`.
   * The `schemaSettings.schemaType.thisDomainName` parameter must be set to the DNS name of the ACTIVE Kubernetes instance.
   * The `schemaSettings.schemaType.otherDomainName` parameter must be set to the DNS name of the STANDBY Kubernetes instance.
   * The `schemaSettings.mode` parameter must be set to `standby`.

**Note**: The `schemaSettings.schemaType.thisDomainName` and `schemaSettings.schemaType.otherDomainName` parameters change between installations.

### Backup Daemon With S3 Storage

MongoDB Backup Daemon can be configured to save backups to S3 storage.

It is required to set the `backup.s3.enabled` parameter value to `true` and set the `backup.s3.bucketName`, `backup.s3.endpointUrl`, `backup.s3.endpointUrl`, `backup.s3.accessKeyId`, and `backup.s3.accessKeySecret` parameters.

MongoDB Backup Daemon with S3 still requires a local storage as a clipboard. It can be hostPath PV or a dynamic storage. The storage is configured in the `backup.storage` parameter.

An example of MongoDB Backup Daemon with S3 is as follows:

```
backup:
  install: true
  storage:
    size: 1Gi
    storageClasses:
      - local-path
  storageDirectory: /mongo/backup-storage
  s3:
    enabled: true
    secretName: mongo-backup-s3-credentials
    bucketName: backup
    accessKeyId: minio
    accessKeySecret: *****
    endpointUrl:
```

### TLS Encryption

Secure communication between a client machine and a database cluster can be provided using TLS encryption.
By default, TLS encryption is disabled. To enable it, set one of the modes.

TLS modes:

**disabled** - The server does not use TLS.

**allowTLS** - Connections between servers do not use TLS. For incoming connections, the server accepts both TLS and non-TLS.

**preferTLS** - Connections between servers use TLS. For incoming connections, the server accepts both TLS and non-TLS.

**requireTLS** - The server uses and accepts only TLS encrypted connections.

To enable automatic certificate generation with Cert Manager, set the `tls.generateCerts.enabled` parameter to "true" and specify the `ClusterIssuer` name in `tls.generateCerts.clusterIssuerName`.

#### Set Certificates Manually

To pass pre-generated certificates as deploy parameters use the following parameters:

`tls.certificates.ca_crt` - a base 64 encoded CA certificate

`tls.certificates.tls_crt` - a base 64 encoded certificate for MongoDB, DNS name is `mongos.<namespace>.svc`

`tls.certificates.tls_key` - a base 64 encoded key for MongoDB

`tls.backup.certificates.tls_crt` - a base 64 encoded certificate for Backup Daemon, DNS name is `mongodb-backup-daemon.<namespace>.svc`

`tls.backup.certificates.tls_key` - a base 64 encoded key for Backup Daemon

`tls.dbaas.certificates.tls_crt` - a base 64 encoded certificate for Dbaas Adapter, DNS name is `dbaas-mongo-adapter.<namespace>.svc`

`tls.dbaas.certificates.tls_key` - a base 64 encoded key for Dbaas Adapter

`tls.disasterRecovery.certificates.tls_crt` - a base 64 encoded certificate for DR Site Manager, DNS name is `mongodb-disaster-recovery.<namespace>.svc`

`tls.disasterRecovery.certificates.tls_key` - a base 64 encoded key for Dbaas Adapter


### Cluster Auth Modes

#### `keyFile` (default)

Uses a shared key file for internal authentication between MongoDB cluster members.  
All replica set members and sharded cluster components use the same key to authenticate and establish trusted inter-node communication.

#### `x509`

Uses X.509 certificates for internal authentication between MongoDB cluster members.  
TLS must be enabled when using `x509` authentication.

Separate certificates are generated for:
- Each shard replica set
- Config server replica set
- `mongos` router

Cluster members authenticate each other using their certificates during inter-node communication.

# Upgrade

The following sections explain the upgrade process.

## Upgrade version path from MongoDB 4.4 to MongoDB 7.0 

- Current version of deployed MongoDB must be 4.4.X

- Check that driver is compatible with MongoDB 7 and update if required:    
  * Java - https://www.mongodb.com/docs/drivers/java/sync/current/compatibility/#compatibility-table-legend    
  * Go - https://www.mongodb.com/docs/drivers/go/current/compatibility/#std-label-golang-compatibility   
  * Other - https://www.mongodb.com/docs/drivers/

- Resolve the incompatibilities for current applications:
  * https://www.mongodb.com/docs/manual/release-notes/5.0-compatibility/
  * https://www.mongodb.com/docs/manual/release-notes/6.0-compatibility
  * https://www.mongodb.com/docs/rapid/release-notes/7.0-compatibility/

- Upgrade applications that use Mongodb first
- **IMPORTANT!!!** Upgrade MongoDB to 5.0, 6.0 and 7.0 sequentially using App Deployer. Skipping upgrade of an intermediate version is not allowed.

### Breaking changes 

- the `mongo` shell cli is no longer supported in MongoDB 7, the new cli is `mongosh`. It has the same arguments as `mongo`, but requires more CPU, so limits of mongo pods must be at least 500m

## Upgrade Operator Installation

This section provides information about the upgrade procedure from one operator version to another version.
The Upgrade procedure is identical to a clean installation, the only difference is that **DEPLOY_MODE** needs to be set to **Rolling Update**. If needed, change the deployment parameters.

### Prerequisites

Ensure you use the same type of deployer for the previous and current installation.
For example, if App Deployer was used for the previous installation, it should be used for the current installation as well. A DP Deployer cannot be used.

### Upgrade Procedure

The upgrade procedure is as follows:

1. Log in to the cloud, `oc login <OPENSHIFT_URL> -u <USENAME> -p <PASSWORD>`.
2. Switch to the MongoDB namespace, `oc project <MONGODB_NAMESPACE>`.
3. Label the current MongoDB PVC, `oc label pvc pvc-mongo-cluster-data-0 pvc-mongo-cluster-data-1 pvc-mongo-cluster-data-2 microservice=mongo-cluster`.
4. If the platform mongodb-backup-daemon exists in the namespace, label the backup PVC, `oc label pvc mongodb-backup-storage name=mongodb-backup-daemon`. 
5. Check the current user ID in docker containers, `oc rsh rc/mongos bash -c 'id -u $(whoami)'`.

Example of the command output:

```
~$ oc rsh rc/mongos bash -c 'id -u $(whoami)'
whoami: cannot find name for user ID 100600
100600
```

Where, `100600` is the user ID.

6. If you have Deployment Config in the namespace that are not part of Platform MongoDB Installation and should not be deleted, delete the Deployment Configs with the name specified as follows:

  * `oc delete dc dbaas-mongo-adapter`
  * `oc delete dc mongodb-backup-daemon`
  * `oc delete dc mongodb-monitoring-agent`
  
  Otherwise, delete all Deployment Configs, `oc delete dc --all`. 
  
7. Run the App Deployer in the **Rolling Update** mode.
   Set the operator deployment parameters according to the current parameters of MongoDB installation.
   Pay special attention to the parameters `backup.storage.size`, `backup.storage.volumes`, `backup.storage.nodeLabels`, `mongodb.storage.size`, `mongodb.storage.nodeLabels`, and `mongodb.storage.volumes`. They must correspond to the parameters of the current MongoDB installation.
   Set the user ID from step 5 to the `podSecurityContext.fsGroup` and `podSecurityContext.runAsUser` parameters.

To check the installation result, see [Deployment Validation](#deployment-validation).

## Upgrade from Sharded to Non-sharded Schema

The upgrade process for sharded to non-sharded schema is given below.

1. Connect to mongos pod.

2. To ensure the successful data transfer from all shards into one, it is necessary to first check the list of all databases in the system - `show dbs`. Note down the output somewhere for further use.

3. Check balancer status - must be true. `sh.getBalancerState()`. If false, enable the balancer - `sh.startBalancer()`. For more information, follow the [Sharded Cluster Balancer](https://www.mongodb.com/docs/manual/tutorial/manage-sharded-cluster-balancer/#enable-the-balancer).

4. Get the list of shards and select shards for removing - `db.adminCommand( { listShards: 1 } )` or `sh.status()`.

![Status Output Example](/docs/public/images/status-output-example.png)

5. Remove all shards except `datars1` - `db.adminCommand({removeShard: "<SHARD_NAME>"})`. **Please note that this operation can take from a few minutes to several days to complete.**

6. After execution, removeShard command output will be shown. Example:
```JSON
[direct: mongos] test> db.adminCommand( { removeShard: "datars3" } );
{
  msg: 'draining ongoing',
  state: 'ongoing',
  remaining: { chunks: Long('0'), dbs: Long('1'), jumboChunks: Long('0') },
  note: 'you need to drop or movePrimary these databases',
  dbsToMove: [ 'public' ],
  ok: 1,
  '$clusterTime': {
    clusterTime: Timestamp({ t: 1715008176, i: 1 }),
    signature: {
      hash: Binary.createFromBase64('AAAAAAAAAAAAAAAAAAAAAAAAAAA=', 0),
      keyId: Long('0')
    }
  },
  operationTime: Timestamp({ t: 1715008176, i: 1 })
}
```
Here, we can see the field `dbsToMove`. If it is empty, do not need to do anything, but if it is not, it means that shard is a primary shard for some databases and need to appoint `datars1` as primary for all database from `dbsToMove` field.
The shard will not be removed until we reassign the primary shard for all databases
reassign primary shard - `db.adminCommand( { movePrimary: "<DB_NAME>>", to: "datars1" })` - reassign the primary shard for 'public' databases.
 
7. To check the progress of the migration at any stage in the process, run `removeShard` from the admin database again. 

8. Check status again - `sh.status()`. Have to see only one shard in the system.

9. Delete all stateful sets for all cnfrs, unnecessary datars (all of it except datars1x), and the mongos replication controller.

10. Set the `sharded` parameter to `false` in the deploy parameters in CMDB.
```yml
schemaSettings:
  dataReplicaSize: 3
  shardCount: 1
  sharded: false
```

11. Set **DEPLOY_MODE** to **Rolling Update** and run deploy.

12. After successful completion of the upgrade job, check the list of databases again (as in step 2). Make sure that all databases were migrated.

13. Check that the remaining shard is operating in normal mode, i.e. has 1 primary and 2 secondary replicas - execute `rs.status()` on any datars1 replica.
