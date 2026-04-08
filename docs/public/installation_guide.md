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

### Store Credentials in Vault

There are several manual steps required for the Vault logic to work correctly in the MongoDB Operator.
Before you start, ensure that the Vault is working correctly.
For more information, refer to the _Official Vault Documentation_ at [https://www.vaultproject.io/docs](https://www.vaultproject.io/docs).

To store the credentials:

1. Create a new v1 key or value storage in the Vault terminal if it does not exist.

   `vault login <VAULT_TOKEN>`

   `vault secrets enable -path="secret" -version=1 kv`

1. Create the Cluster Role Binding.

```
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  annotations:
    cloudops.netcracker.com/tag: KEYMANAGER
  name: <MONGODB_NAMESPACE>-key-manager-auth-delegator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:auth-delegator
subjects:
  - kind: ServiceAccount
    name: vault-integration
    namespace: <MONGODB_NAMESPACE>
```

Where, `<MONGODB_NAMESPACE>` is the namespace where MongoDB Service is deployed through MongoDB Operator.

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
| `mongodb.flavor`                              | false     | string            | small   | The flavor of Mongodb deployment resources. Possible values are  `small`, `medium`, `large`.           |
| `mongodb.dockerImage`                                                        | false     | string         | ``                   | The Docker image of MongoDB.                                                                                                                                                                                                   |
| `mongodb.additionalNodeLabels`                                               | false     | object         | `{}`                 | The additional node labels for each mongo replica.                                                                                                                                                                             |
| `mongodb.rootUser`                                                           | false     | string         | `root`               | The root user of MongoDB.                                                                                                                                                                                                      |
| `mongodb.rootPassword`                                                       | false     | string         | `root`               | The root password of MongoDB.                                                                                                                                                                                                  |
| `mongodb.rootUserRole`                                                       | false     | string         | `'root', '__system'` | The root user role in MongoDB.                                                                                                                                                                                                 |
| `mongodb.cnfWiredTigerCacheGb`                                               | false     | string         | `0.12`               | The MongoDB Wired Tiger cache in GB for CNFRS replica.                                                                                                                                                                         |
| `mongodb.dataWiredTigerCacheGb`                                              | false     | string         | `0.25`               | The MongoDB Wired Tiger cache in GB for DATARS replica.                                                                                                                                                                        |
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

## Vault Registration Parameters

The list of Vault Registration parameters is as follows:

| Parameter                          | Mandatory | Type   | Default                   | Description                                                                                                                        |
| ---------------------------------- | --------- | ------ | ------------------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| `vaultRegistration.enabled`        | false     | bool   | `false`                   | If the Vault storing credentials are enabled or not.                                                                                |
| `vaultRegistration.url`            | false     | string | ""                        | The URL address to the Vault service.                                                                                                  |
| `vaultRegistration.rotationPeriod` | false     | int    | 8640                      | The amount of time Vault should wait before rotating the DB passwords.                                                             |
| `vaultDBEngine.enabled`            | false     | bool   | `false`                   | If the Vault MongoDB DB plugin is enabled or not. If the DBaaS adapter needs to store credentials in Vault, set this parameter to "true". |
| `vaultDBEngine.name`               | false     | string | `mongodb-db-engine`       | The Vault DB engine name. |
| `vaultDBEngine.pluginName`         | false     | string | `mongodb-database-plugin` | The Vault DB engine plugin name.|
| `vaultDBEngine.allowedRoles`       | false     | string | `nc*`                     | The Vault DB engine allowed roles mask. |

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

### Vault Registration

```  
vaultRegistration:
  enabled: true
  url: http://vault-service.vault-test:8200
vaultDBEngine:
  enabled: true

```

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

# Installation

The following sections describe the installation procedure of MongoDB Operator.

**Note!** Clean Install must be successful! Successful Rolling Update on top of failed clean install will lead to broken mongo cluster.

## Before You Begin

Before the installation is started, it is necessary to implement the following procedures.

### App Deployer Preparation

You need to have the artifacts ready for App Deployer.

1. Select the artifact version for the App Deployer (version as a string) from the related location. <!-- #GFCFilterMarkerStart# -->
[https://github.com/Netcracker/qubership-mongodb-operator/-/releases](https://github.com/Netcracker/qubership-mongodb-operator/-/releases). <!-- #GFCFilterMarkerEnd# -->
You need to use this location link as the value for the `ARTIFACT_DESCRIPTOR_VERSION` parameter.
2. Navigate to groovy.deploy.v3 at [https://cloud-deployer.netcracker.com/job/INFRA/job/groovy.deploy.v3/](https://cloud-deployer.netcracker.com/job/INFRA/job/groovy.deploy.v3/).
3. Specify the values for the following parameters:
   * `PROJECT` - The target namespace for the installation procedure.
   * `OPENSHIFT_CREDENTIALS` - The credentials of the user on behalf of whom the deployment process needs to run.
   * `DEPLOY_MODE` - The mode of deployment, either `Rolling Update` or `Clean Install`. The `Clean Install` mode deletes everything from the project before the deployment.
   * `ARTIFACT_DESCRIPTOR_VERSION` - The version as a string from the related location. <!-- #GFCFilterMarkerStart# --> For example, [https://github.com/Netcracker/qubership-mongodb-operator/-/releases](https://github.com/Netcracker/qubership-mongodb-operator/-/releases). <!-- #GFCFilterMarkerEnd# -->
   * `CUSTOM_PARAMS` - The custom parameters that overwrite values from **values.yaml**. For more information, see [Deployment Parameters](#deployment-parameters) and [Parameters' Examples](#parameters-examples).

**Note**: `DEPLOY_W_HELM: true` **must** be set in `CUSTOM_PARAMS` or in `CMDB`.

4. Click **Build**.
5. Сheck the installation result. For more information, see [Deployment Validation](#deployment-validation).

### Helm

Before you start with the manual deployment of MongoDB service using Helm, ensure that you have Helm 3 release.

Alternatively, you can install the operator using the Helm chart from the `charts/helm/mongodb-operator` folder.
Install the Helm CLI on your machine. For more information about Helm v3.0.0, refer to [https://github.com/helm/helm/releases/tag/v3.0.0](https://github.com/helm/helm/releases/tag/v3.0.0).

To use a CRD-based configuration, specify the **values.yaml** file as shown in the following example.

Also, you have to specify the proper microservice images in the **values.yaml** file. 
<!-- #GFCFilterMarkerStart# --> The list of microservices can be found in Microservice versions section of each release on the Releases page at https://github.com/Netcracker/qubership-mongodb-operator/-/releases.
<!-- #GFCFilterMarkerEnd# --> 
Then follow each microservice tag link and find the Artifacts section with the Docker image name.

1. Clone the project to a local machine using the following command: 
 
   ```git clone git@github.com/Netcracker/qubership-mongodb-operator.git```

1. Navigate to the Mongo operator directory using the following command: 

   ```cd mongodb-operator/charts/helm/mongodb-operator```

1. Login to OpenShift using the following command:

   ```oc login https://openshift_url:8443```

4. Deploy the operator using Helm `helm install mongodb-operator`.

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

1. Run the deployment job ([DP Deployer](#deployment-using-microservicedeployer-helm) or [App Deployer](#deployment-using-app-deployer)) on the STANDBY Kubernetes instance. For more information, see [calico_deploy_standby.yaml](examples/calico_deploy_standby.yaml):

   * The `schemaSettings.schemaType` parameter must be set to `dr`.
   * The `schemaSettings.schemaType.thisDomainName` parameter must be set to the DNS name of the STANDBY Kubernetes instance.
   * The `schemaSettings.schemaType.otherDomainName` parameter must be set to the DNS name of the ACTIVE Kubernetes instance.
   * The `schemaSettings.mode` parameter must be set to `standby`.

1. Wait for the job to finish successfully.

1. Run the deployment job ([DP Deployer](#deployment-using-microservicedeployer-helmr) or [App Deployer](#deployment-using-app-deployer))  on the ACTIVE Kubernetes instance. For more information, see [calico_deploy_active.yaml](examples/calico_deploy_active.yaml):

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
    endpointUrl: http://minio-tenant-4.paas-miniha-kubernetes.openshift.sdntest.netcracker.com
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

Example: 

```
backup:
  tls:
    certificates:
      tls_key: "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb3dJQkFBS0NBUUVBd2h0bTBubjBpV3lwd0wvUk11eHc3TytZcVF4b1I3Z2FUNlVZYU1yNWh3MjYxdWozClBSd1dkbThkUzZWZm1sS3gzY3RPM1h1TEo4VUVoUys1QS90Q044MEFobUtqVlFobDZrbmpPeUErUjlkL3QyNEEKdXJjRWNEaTlXT2lMV2RJT0dvRDFGd1VRY2ZQZk5qc2pWdWFLaldTSCtLbHhtbm9uV2prQ1NKakJZZWFwM3ZjZAp0QTlrUjF4MWs3aFp4R3RNeEdOSXRJQ1NUT3NtY1d6OXFFeDhNOUdDTFZwN1VVSytUMFJzQTVtdmlFY0llSlhxCkZZYlNoRklBejIvSGU5OWxLR3lZRUpMWjMzYkt1ZzBVbTcvbXMxNWdLK0JXOUxNNi9wL1VpbjV5dDU1MThlcm0KOXFKS203WVhiNWtzb0V5cEdiTVNCK3d4T3VDd3F5OE5BbDVPSFFJREFRQUJBb0lCQUdscHhpcFJ6c3ArOTZWVQp4b0NZUlM5M1l2bVRZbUpvaWVsczZGZW91MkJyeFdjRzk1WDVWZjJWbEZ4TGdDTG4rKzVPaGhMa0VBdFdCSUZzCkRGY3NNYWJxTHZuTVFaVmhUUyt5VnJQNmE3aEtRUExWeTVHYTZNOGxFVGRpZXFNWjMwem5jYkxCcmsra09EbFUKWG5uSUU4QjdzeGdJdFVoR1JHN0wvUUI5N0srRVNJMFExT2t6UlRCT3hubDNOd1hmditVQ3hScGtESVprcS9GZwpaRnp2Sk1jcHdLYXBVRk1mZ2laYjZBTnZDSFk1STRiQVo2YnZXbjBnalNTOU5QSnVQcjRqcCtIcExaQW1CV2FtClgvRThYU1B3Y0JYcFplMFExb1VXTktyOVV6R2FYNlNNaWkrM3dnNzBoSmM3Q3pybWh6WnhVOWR0TGs5MFk2NkkKSDVabDlBRUNnWUVBeUk0UTRMOWhnZzNuTUpOanhXb2hDNHllelFXR3c3YVpyZU82ZWxhb0hJYWdRN2l5TmwxTwpkRWlwRWsxZTd5bGw3dXZGMzVsZ2gvSlJ6VDlsVWQ2dXBsd0pPVFg0OHhZL2FOR2ZwU1Z5UmRYblN1TmQrWkxTCnQ5ZUp3ZWp0dUtQVU5oc2hxdGI5OEIvUXNqS2p1VlFOdU5HZURnd1VtalVGSnNtdTRHNm9qUUVDZ1lFQTk4VCsKL3ZtVlhFd0h2UnI5UDNLcE4xWXNqRlZ3T1IvTEd6OXVBcVRmdFo4WE1XQkxibjBIZWtpUWp2cS9nTWkvQ1U4Mgp1UTlIWTcrN2hiZVI4L2tiTCtqNkQyWFl3M0FtaUd1eXY1bXBOa1BqREIzM3BTQ2g3alJBUmw0anlrYXMyNUpQCkhCVzdEajMyTDlHYndub1daaVM3am1sYVVrNGlIdkE5ZlJGMVZSMENnWUF3cU8rRmFFbmpRVFpQdmVNZU9mTE0KbDVETUU4UXY1alVCVU5pazZET2Z3RFpRV0JhOVJBUk9DSGNsSHFxakFudGQ3Y3l6eE1YOEZob3MzMjNZNEZ1bAp0M3p4YVp2K2R1NXBvenJGMmdFUTJxWmtzQ2ZUN3dDN1pFdGpSZjJ2cCtoTVBHYjl5VzRSZmRhbjljdHRvdXcxClpINmh6K0tMeThOMU5zZjhZano1QVFLQmdDS2hYaUsxTDdNZXpWWVpGNXh1b2tnaHUwaENDTlZ6SkNoQ3pWV0IKUmVOVXdTRWRuRzFzL0VhVExlRk9Hc1lkU05ZOFJDSEppT2pnTzQyTkF0RmUxL1h5VWtFa3N3OWQ5WVRMeU1nTwo2aCt6aldCOEw4amNyc1ZrZURkZG9STDhuZHh5cnF2MlBaYllBampjeXpCN2IvWUczRkFqV1lSM2R6MXJ4cXhjCmJGSGhBb0dCQU1hczVrTXNGWWZWVFl3SlZsQ2FVbUNnSVVuUkNwNlhnaFYwNE5FR0VRM08zU2dWUGtEbGZPQ0oKeGhvbXFWZ2Vhdi9RY3RWWDl3bWZ6Yk83cExHdTRrTTg4L1p6Mld4cGJ2a3JKZU1ZWUorbXJNVHp3S3NMTzFFUAp1b2h1ZUUwYzBPbWJlWlZJbU83WDNYSFphNnZZZ0J1MzBJOHZLZ2lyd2hKZDMyNlAwejBGCi0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0tCg=="
      tls_crt: "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUV4ekNDQXErZ0F3SUJBZ0lSQUkvMkp3ckpBcmp6R2FOOVFqZnRUZDh3RFFZSktvWklodmNOQVFFTEJRQXcKT3pFTE1Ba0dBMVVFQmhNQ1FWVXhFekFSQmdOVkJBZ01DbE52YldVdFUzUmhkR1V4RnpBVkJnTlZCQW9NRGs1bApkR055WVdOclpYSWdURlJFTUI0WERUSTBNREl4TlRBNU16YzFPVm9YRFRJMU1ESXhOREE1TXpjMU9Wb3dKekVsCk1DTUdBMVVFQXhNY1ltRmphM1Z3TFdSaFpXMXZiaTFqWlhKMGFXWnBZMkYwWlMxamJqQ0NBU0l3RFFZSktvWkkKaHZjTkFRRUJCUUFEZ2dFUEFEQ0NBUW9DZ2dFQkFNSWJadEo1OUlsc3FjQy8wVExzY096dm1La01hRWU0R2srbApHR2pLK1ljTnV0Ym85ejBjRm5adkhVdWxYNXBTc2QzTFR0MTdpeWZGQklVdnVRUDdRamZOQUlaaW8xVUlaZXBKCjR6c2dQa2ZYZjdkdUFMcTNCSEE0dlZqb2kxblNEaHFBOVJjRkVISHozelk3STFibWlvMWtoL2lwY1pwNkoxbzUKQWtpWXdXSG1xZDczSGJRUFpFZGNkWk80V2NSclRNUmpTTFNBa2t6ckpuRnMvYWhNZkRQUmdpMWFlMUZDdms5RQpiQU9acjRoSENIaVY2aFdHMG9SU0FNOXZ4M3ZmWlNoc21CQ1MyZDkyeXJvTkZKdS81ck5lWUN2Z1Z2U3pPdjZmCjFJcCtjcmVlZGZIcTV2YWlTcHUyRjIrWkxLQk1xUm16RWdmc01UcmdzS3N2RFFKZVRoMENBd0VBQWFPQjJUQ0IKMWpBT0JnTlZIUThCQWY4RUJBTUNBcVF3RHdZRFZSMFRBUUgvQkFVd0F3RUIvekFkQmdOVkhRNEVGZ1FVODg2Wgo3ZCtrbWFwd01tYzN4MGQ1L2hvdWxZWXdId1lEVlIwakJCZ3dGb0FVL1J4ZGpJS0NkZ2t2bGtQbzJlNXhRb2NvCkVSc3djd1lEVlIwUkJHd3dhb0lKYkc5allXeG9iM04wZ2hWdGIyNW5iMlJpTFdKaFkydDFjQzFrWVdWdGIyNkMKSFcxdmJtZHZaR0l0WW1GamEzVndMV1JoWlcxdmJpNXRiMjVuYjJSaWdpRnRiMjVuYjJSaUxXSmhZMnQxY0MxawpZV1Z0YjI0dWJXOXVaMjlrWWk1emRtT0hCSDhBQUFFd0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dJQkFJSGZtUnlvCkJTT2lvaHJlRjNJclJCeWNTVkxBdHF6a2NSMHg0QzZOQ0llc1VQSDVHdFpzNWQ1bExHQTlPd2tjR2J0NEtNWk0KbDNBTzQ4V2NYMkk2bkRIc1hrN2J3WE1hbWlTN1RueU5vZ1d6U3dERUtxNnhBZkMzYXA3OEdqRmo1ajU5T2FpeAovRW40dGM1dkVkbkJuODU4WkVCVW9XR1RXOVNXR2RkM3lSaENLcXdlNUc0YkFDWW5ZeWZjSWR6L1NGQ1ArcS9qCk9KSVdORXRVdnpLSzUwWDZXS21vV1piRXAzallHMGVUcmkxYXY1Qmk4eDY5N0FtbHQ2ZTZ5bWZKaHpZMHhYSm4KRzJXL01LQy90VUk1eEVjM1hybUd1K2pzNjZUNWJBOC9rcXBGZitHSDZXbjZzTFJDMkUvNTJuWFhpalVBSTFKTQpLQjJIdjB1NXFrR0dzK3JXSjBocFVhME5Ednc4RmtyWUFENjJEeHVyNjVMeExMZC84U0pmWkVrRDZDaEVpelJ4CnNONm9vaUR6bnRoRDdDeWp5bm1zVzRQdzZhbFFoaXBGNFpQYzJucGlFWGEzUDA1OXV3MXV5clN3akNRTzl5aGsKNDl2ZTZBWTdRLyt5VnpqODNmRFlPY3NCU29xZ0ZLOHZlVnM5dUo2NU9hUEdLelVydmJiYlozQnJmcnZCVFBpcwpvMmoxbHFIbFdjNm5vNFVYM0dzRkJIQ1p5UDRqZitwZXpuTUdlSFFyODVZNk10VXBkaTMzbFJLNURWWGhNbis4ClVGK1N1S2E0VHhJMG1BRE9ya1NZTTlHbnNHRWJXeVB2ZDU1cFhWM0ZtTWhWU0lHLzMxUTd4VnVLamFFWlh4RWwKV0toWFh4eU1QRUJqZ0NqRVVTM2ZDMlVIMlo2TGZQaHRKZGRjCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K"
dbaas:
  tls:
    certificates:
      tls_key: "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcGdJQkFBS0NBUUVBeHRrSzBrcEhUdDdHaHQwRzlEbU14SEtLQUdOcHJpSTFZK09hL0ZqTFBpRGF6RDk5CjBZcGVWYUFVSkJTc2pPMVlaSnVFMlh6dDVISHdpM3cxWC9IS2lwL0xwSDRCRHVURzZQRHBEQzFRTTJZWjhSL0YKZms5eEFoTFhtUGJHa29OWVF2L3p1YTByVWxmSzhqRnFKNHN4M1lBelhVRlNTZU9kZENhVDlxRC9LeFR5MU9xcwptMEx0Q2p4aEhCNFJKeG1sd3M4RTFlbi8zY1gveEsyWHFReE5FNTJIKzJiU3JEdUZFRUhJNnZ1NWp4N2ZqN05rCnR1MkJkYlBkdms5N2hKcVJ5d3JIbzFKaGJHWGE4RkZXbkVsekRIUjFMWVZmQlhVaEpZZklZaDB2R0hSRmVnVjgKSENIY25HL3EycGdKbTlET084R21aM3U4RlBMaHBTUWt1aHNEUXdJREFRQUJBb0lCQVFERjUzb1h0WW1tKzUzRwoxL0IxM1ZrMm8zQ3Axa2QxNGVJVldwQUVHek9jMEFJelNmV2xPUHVPYU5YaTJ4aW80Z2daaVpiOUJwT1Z5N2pHCmVvWjh5UjcyUFBmbTdPbU1zekVzNGFod1VDRUVKdGdtM2FJbmhsVkk1UXZpMTZqbVpRYlJHQUN3aWFNV1B2NWoKY2I3ZlFIQU9yZXR4SXZRTlNoYUpaV1BhRUg0bGx1QzMwSEVNb09JMmR4TVY4N3J2djIxbTdHd1FPeXU5cUVsNwp4cUpVUVIyaXY5MzNBZlV5OUFxWndQNHB0S0FqYTJERS9ZU1BsMGxIYVFCU3dNVzlCYVE4VkMvd3pmRGNSR3c2Cm1OQXhJb09zVGpySHFsOVUyYVBTYnJlaS9RYU43VERldXpnQ3lkVDZhRHVqOVo0WFZXQlNEQSsyaFpUZDc5VUQKVHZ4U3BMREJBb0dCQU9UMkNRV240VlN1Ty8xcTQ3YkhnQ2g2YTJuUlBkbFlaam5QZVR5eDAvSE9rSy8xTGZCZAptcCthUVdNS3RBOGNuYVQ3ZFBGYURiYXBMZ1lDUXZ6d2MxVGtlZ0NuQ3dTdjFLV1pjUVVTcEJkN2o1OXVGWm1GCmhJbTVuV2RnTWhYREdlVHRoRXJpVER4dXBxMWk1U1FocHRYeE5aTnBkcDBrMFR2ZldLRTUxeDV6QW9HQkFONVUKbnhGMVNQZXlHVnhoV001Y2szYldPNHVGUUo1Vm9lN3V3bTdvZXhBb09uQ0dBN1cxdG53TWp4N0pmMG4xbm52UQozNVR1YWkrRHlYdnBHZlJqQkZ0bmlJYTVlSlNzRi9vS0REM3Y5bzBVQjdyckRXOXNmQStqYXd6QUx0WFFvK3pRCmIxemt0TEtKeDIwNVZoOFU5L1lVY2kzZi9YWkdYMkdVRTA3TzFBUHhBb0dCQU51SzAySFo5U1dXb0MxQjVqR28KSUVvd0FIa0p5dzF5Unl0ZHRybXRKalp4eEtrRUp0V1pXNTk0Y1FSQUNpR0haZDRCdzhOOWZ6TE1ERFoweXJqdwo4eFhPc3ZHWE91aDJsU2RvOTBkTzlZc1N6c2VuN2d3MFM3OG4vVGRYdFE3SzhqUmlUM3ppZXdsamJHMUxLNzYyCmlkd1JHemRMWkJJUWNKVEJkNkc0N1gzakFvR0JBS1dDc1hnNXE3eFpwVytVT0p4SFpyQU5BLzcxa0FsUERtSGsKOUhIRU4vanJPYllTemlnendrbk92NnpYckI3Tzd0Q1Z5aHdBOEtPMnBBUE9vRGZDanJmTTkySDBLTVBrNldTRwpubDV0aVVtMUk1d081ODJQSVR3ekY3cENSNXQ4MnN1c3ozcUQ5OUVCcUtpekNsM1JLbGJUR2J6MUJxZEo5QytjCklGT0d2V2JCQW9HQkFLMnJhamR6aGlRbXVQQStxRlRPUFJYdlhVVkNQMXovZDkrUTlGa0YweWtNdGczTTVKYkYKdG9vZmkzN3FBc0UwRlN1WWk4dWk4czZ5SnlWeGkzQUx5VkhIWTlGNUw3QjN6UFdsamtFNDF5QXFTcDlGcDFzYQpZQVF0bGkrb25waUlUN1hQN2hvL2h0K2NGQnVHZkZTSEVYcGFPYS95dHI0TmlRM1U0eDZDazN4ZQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo="
      tls_crt: "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUVwRENDQW95Z0F3SUJBZ0lSQVBFREpybUp1d3RQL1U1NnQ2YVRkU1l3RFFZSktvWklodmNOQVFFTEJRQXcKT3pFTE1Ba0dBMVVFQmhNQ1FWVXhFekFSQmdOVkJBZ01DbE52YldVdFUzUmhkR1V4RnpBVkJnTlZCQW9NRGs1bApkR055WVdOclpYSWdURlJFTUI0WERUSTBNREl4TlRBNU16YzFPVm9YRFRJMU1ESXhOREE1TXpjMU9Wb3dIekVkCk1Cc0dBMVVFQXhNVVpHSmhZWE10WTJWeWRHbG1hV05oZEdVdFkyNHdnZ0VpTUEwR0NTcUdTSWIzRFFFQkFRVUEKQTRJQkR3QXdnZ0VLQW9JQkFRREcyUXJTU2tkTzNzYUczUWIwT1l6RWNvb0FZMm11SWpWajQ1cjhXTXMrSU5yTQpQMzNSaWw1Vm9CUWtGS3lNN1Zoa200VFpmTzNrY2ZDTGZEVmY4Y3FLbjh1a2ZnRU81TWJvOE9rTUxWQXpaaG54Ckg4VitUM0VDRXRlWTlzYVNnMWhDLy9PNXJTdFNWOHJ5TVdvbml6SGRnRE5kUVZKSjQ1MTBKcFAyb1A4ckZQTFUKNnF5YlF1MEtQR0VjSGhFbkdhWEN6d1RWNmYvZHhmL0VyWmVwREUwVG5ZZjdadEtzTzRVUVFjanErN21QSHQrUApzMlMyN1lGMXM5MitUM3VFbXBITENzZWpVbUZzWmRyd1VWYWNTWE1NZEhVdGhWOEZkU0VsaDhoaUhTOFlkRVY2CkJYd2NJZHljYityYW1BbWIwTTQ3d2FabmU3d1U4dUdsSkNTNkd3TkRBZ01CQUFHamdiNHdnYnN3RGdZRFZSMFAKQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCL3dRRk1BTUJBZjh3SFFZRFZSME9CQllFRkFETCt6dTM5bEIvRVUrUAplRzVEVjlYSmdFZjlNQjhHQTFVZEl3UVlNQmFBRlAwY1hZeUNnbllKTDVaRDZObnVjVUtIS0JFYk1GZ0dBMVVkCkVRUlJNRStDQ1d4dlkyRnNhRzl6ZElJYlpHSmhZWE10Ylc5dVoyOHRZV1JoY0hSbGNpNXRiMjVuYjJSaWdoOWsKWW1GaGN5MXRiMjVuYnkxaFpHRndkR1Z5TG0xdmJtZHZaR0l1YzNaamh3Ui9BQUFCTUEwR0NTcUdTSWIzRFFFQgpDd1VBQTRJQ0FRQW5YdEtsRjUrREU2T2ZKUXZGTTluNWptN1d0MW1kWGM3L1VnbExFQzV2bWRFckdtYi9PMmJ0CjdEVG5xSnF2cCtIUis1aFVOK29EOU8zQ0dDQU1XbDZxcEhzZGhweE4wVUhGNTVYUkxNQThDQ1MyQlppa1VKWFkKa0RrSWh0eFNieFFxRUFGNjNjYnlvRTRZRVFLaVdxSmdRajVUbGJEVXhsSldqUlNDZVVIUXd2Z01wSmpUTzd5TApoN1EwT3psNC8yN1I4ZnZZamwvZkJYQmNIWnlOYURQWG9FU2dDcXJta29IVlVnSWt6dEd1eWZOUEUxR1o3bzhTCjUrclRHQXJTZjJCMkJPaEpaU0dqMWdxY2s3UEQyUE9HbTNmU01haEZWZTJNeGRrSm5LZGFSUHJBTlF0cE1wZmUKUThyY3dpcTRMWWVqTlFmTVJ6Nk4zMmZGM2o1d1FKdlZmdTE4U1ZCNWU2RXBYTlplWlNVa3JDTXBWTmJzUVlOSApwcEtMVkxxdmNyNmdpL21Oc29EQ1IrQ2x1OEI1NkcwS0RkTU5jS3pOMG42c0NESlpmMmJxU1FmWDFacWdPai95CjRCM3RTdURqdEZkRjY4SlptZ09rU2FGSlhTMHp1Y01RYlVzTkk0U2FPWFg2WEhkQmZNNVp2V0RhbUt3SkoxU24KeldqbkhnSTVadkljb1l3eTNmOTlGYW1KMjQ5VVdOR0NETXlrL1hsZklpY0dBSTdqS3JmN1NIdzlCT01OdDQxQgp5V25VeWZNTDh1OGNrdlZrempJWFRET1R0Zzd5dTd3b1FaTi8xRkhKQVVUcTBMbUxqY2NVaEhJbGU0ZUphMnRQCk5NUHh0U25QazQvTlFYUWdnZDMyNzhrM0IyN2RncGZRTm9kSDRRaUJpTHFqNVVBSGV1VW83dz09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K"
tls:
  mode: requireTLS
  generateCerts:
    enabled: false
  certificates:
    tls_key: "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcFFJQkFBS0NBUUVBNFhTUW1sdUxydXZnZHBLNmlZVTBMTUMxUk4wcHVmRmhnQkZqT2dpbW81WkNIeXk0CkhtTGJqNnBJR3pKU3ZXSEk1Y0JONU9icW96bGFvaTljaHlHL25CalgvSXdPN2Y4czR5OWxVN0IvVFdndkpZblUKQnBoSml1M1lpbllpUDMwMlpQK1U0ckgvSkZDNmNQYXJlMG5OcUJHMDRrTFBHeGRSZVd1WUlGS1B2OTlNdDhQegphbU9sdHZoMDloTElUL0pMMHdRU1Z6cWtYSWtUV0UyRGJYOVNDZ1dYWEN0VG5Fc01EREdUYURIN2RqelBCb2VNCkpubGpDcGZGdVVOaXI3U1RxVUZrMUdEcjV6QTNyWjJGMXlldVpkNHA5Yml3dnJNQndCclBQeXZkNzN6em4vcTgKRDUvcTRqTjlqL2pKd0hpMFZiZ3hwZWE1QmJlTmpDb1F0Z1phb3dJREFRQUJBb0lCQVFEUkdkcFB2MTVESXZQeApKVThxNHNjc1JxTVl0b0xQdVVjemoyelhVMVN5WGxiL01PdW5Dd3NXS05sdGwvUFRQOUVpL1lPQkxJWXNVckp6CnY3ZHlnV09FTkNxR1NhUkRLaXNJbmxtOUQvSlI2YkhvZi9lTkVrc0xObU5pc0FROW5EVUo0VjNHRDA1UzhTaXEKUXExeTBGV1VicSswTmtCOW9OZm81RmlZaWRwWEdkYi9rQUd6MjljSkVrK1pzUzh5UWgzWGl5WTdmQ3dIQlZhUwo5Q05xM1RkdHhBVjA4bjV3aXFLOWp3SHM0NUpwdGYxeVFSR1A0QXRtaGFqN3NmKzRsTkNzQmlwYUNuQlYwSmFXCkxGbWlEakVnanQ0bStzRC9udytBR2NndGVSNGx5R3l1RWJGRCtJMk55YjFaRzI1T2FyTCtqNisyUWdhUVpYUGwKMU1CcDlJVVJBb0dCQVBXV0NCd1p0dmF3QS9iakdjdi9GT0MrdlNHeGZGUlZXYi9PRnRoRTkxOWJrSVVnWFdvNwpxN0FwYXVlVlg1Q1BtcW1zekxDcit5WlF6UlVsbVQ2V2pFZDFDSnZuaXhhbEw3K0x0aFZDYktRci9XNXFSckI1CmNVMjZ5QmlKOWhCT2tlY1JsQUlWNGs1eWVBVWlwS2M0NkhqaGM5MS9EWXJMeXpDZCtFWEhlL3ZiQW9HQkFPc0UKQU5sR0xrL2ROUW5JNU5NY1pPbWZERytxWVBoOXZ6dkkzMDI3K3Vuak0rcE1ZYjY1c1BpWkswN00wRittZDJScApUR1MvSDk1ZFJlTnRiQ09GNWFyZDd1QlZPanJrd2wyeVV0c2JIQ2duejBhNzFwT2VuK0tEVzI1R0hlZEdLWFhZCnhqRVdrZzlTWnBoNGFSdUFOOW9rUTBIK3pmbFdqd0E4OGtpTUR2clpBb0dCQU5nNDZSajhsdmRwRDRSK2ZNYjcKNWdEZVRwenNyRjkvNmc0U3dGQlhvRWpIMEYwMW1xbWVzZEhmRlcyaU9VcUk5UTR3d3VORitGREswVld1RGRkcQpLMFg2eDhLa1FQU0dLWjBHd0NERm8rdURnNVdFWW9xYjBlTXk4VnVSbENEVlhHWktOcnNEVTRYb0NMM1V1NDB6CmNKS0ZSVU1keXVtSjluTHVrcG0xUWZjREFvR0FiVmtPZ0FtOGNLSnZGQjlxQUtRY2UrcnA0V2IzK1lhZ25OT0kKdXVWMUNMQVRMcmZkWHQyTmJ3M3RiWnUwZEZ6Qy9uQlVBQ2hCVHJnOVZXVkxSSGYvZFhJUHZFZExjYTJRbGdIcgp0VkMyMkNRMXVDYWIzMUdWK05HL2o5NkYrVjdXMmFORURBRUJjcW1YWE9maGw4OGZyWnJqeEdnbk5CVkhNZ2dwCmZ3SFQwbmtDZ1lFQTZFSmI2dXJLWDdZTFRwcFlNR1FOYWJHSzNqci9YZXFVTmZ3eEUzZGJHTUdydll3RHJCUkIKejRUTTVsMCtOQkF3Yldqb1hmZkhsWUJtWVQwRlIwVEoxZ25CQVNsblY1TXVsWC9vT0JTODU1Yk92R1JVdm1reAptMmtvMEpDK2tLWFpqdWRRK2N0cUMyREV2eGhOZ2ZwVEdFUW5UVWFZMi9PT1NKNlJoaU0zK1lzPQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo="
    tls_crt: "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUVrRENDQW5pZ0F3SUJBZ0lRTjRORE9FVWVEdm03eWRtc1JPVGhxVEFOQmdrcWhraUc5dzBCQVFzRkFEQTcKTVFzd0NRWURWUVFHRXdKQlZURVRNQkVHQTFVRUNBd0tVMjl0WlMxVGRHRjBaVEVYTUJVR0ExVUVDZ3dPVG1WMApZM0poWTJ0bGNpQk1WRVF3SGhjTk1qUXdNakUxTURjeU9ESTFXaGNOTWpVd01qRTBNRGN5T0RJMVdqQVZNUk13CkVRWURWUVFERXdwdGIyNW5iMlJpTFdOaE1JSUJJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0MKQVFFQTRYU1FtbHVMcnV2Z2RwSzZpWVUwTE1DMVJOMHB1ZkZoZ0JGak9naW1vNVpDSHl5NEhtTGJqNnBJR3pKUwp2V0hJNWNCTjVPYnFvemxhb2k5Y2h5Ry9uQmpYL0l3TzdmOHM0eTlsVTdCL1RXZ3ZKWW5VQnBoSml1M1lpbllpClAzMDJaUCtVNHJIL0pGQzZjUGFyZTBuTnFCRzA0a0xQR3hkUmVXdVlJRktQdjk5TXQ4UHphbU9sdHZoMDloTEkKVC9KTDB3UVNWenFrWElrVFdFMkRiWDlTQ2dXWFhDdFRuRXNNRERHVGFESDdkanpQQm9lTUpubGpDcGZGdVVOaQpyN1NUcVVGazFHRHI1ekEzcloyRjF5ZXVaZDRwOWJpd3ZyTUJ3QnJQUHl2ZDczenpuL3E4RDUvcTRqTjlqL2pKCndIaTBWYmd4cGVhNUJiZU5qQ29RdGdaYW93SURBUUFCbzRHMU1JR3lNQTRHQTFVZER3RUIvd1FFQXdJQ3BEQVAKQmdOVkhSTUJBZjhFQlRBREFRSC9NQjBHQTFVZERnUVdCQlJXYmZ3dkNLRFYvSXFHNmVnRmNIczE1emR3S0RBZgpCZ05WSFNNRUdEQVdnQlQ5SEYyTWdvSjJDUytXUStqWjduRkNoeWdSR3pCUEJnTlZIUkVFU0RCR2dnbHNiMk5oCmJHaHZjM1NDRDIxdmJtZHZaR0l0WTJ4MWMzUmxjb0lPYlc5dVoyOXpMbTF2Ym1kdlpHS0NFbTF2Ym1kdmN5NXQKYjI1bmIyUmlMbk4yWTRjRWZ3QUFBVEFOQmdrcWhraUc5dzBCQVFzRkFBT0NBZ0VBWnRWM25vYXRlNndpS1drOAo1UXM5NWx2R1lkaXFtT1BqU09lY1REZzEycExySWQxRExBRFhFb0g5WlJBVkU4L2Y2U01nZnRJenJaQUoxdlpqClQ5aWhnSWNkUWw5WnU2Zjk1eWdadFZwbm1tMkxTd05QMi9HZ0hmYXNVUVk3MGQwK2VhaHFCbkJrRndvemRVc1QKdjRSUWhGL1FXUU5vNEpWYlZNNkFLLzVKcHVpdnFlTExIaFVMNzlac0FoTzM4b3AyZVpHTm92T0ZqSGJCNmo4TgpmVTV2Z012L1JmRjRJVXg1WUI1eFplQTh3N3BVdERNSE8vZXNqVWVvOWExZ0xpSksrU0o4TG9lR05MQUphRlJICmFUMUNQZVl1MEpURGc0ZmQwSUpqTVJsNFFyV21BcVFQdWtTeVZ3cmQrYkNudFYyY0kwY1BKMUFmZEFwSG1pYmwKdTlHb3RFR084eWVlZGZkNWJoSjVxSUlCOXRBWk1MbVBZMGNNdDBEaDFPaVluOXZOdnlDYnBrS0FxQ1lBV2VlTwo0bVVkRXJnVGN2eWVCc2VqcXBwb2NtMEp4N2l0NFB5NHk4VXl5Q3cyTFhvaHppcGFHMmo0Q2xJblVnWG9hN0RCCnRVNTUyTWRjVjNYV2dzR2loa25OVXNYamdsMGpTbk9LTTUrejhOWklVckRlQ0FqK2NqY1FEWVZETUs4bElyYU4KOU5aL1E1d2RrR0hJMTNxVDdHK3hHYUxUR0hZam1VaFhQZzRiL2Y4cFdGZ2VhamhoZk9qRHhDM1NuaUxLeGFDWApSL2NjdlJQemRDdGJQTTYydXhRQ0xNM0RTM21NaTZIc3NxZ0g2QWQ2UDNFSkxNT0JiMEs5Q1ZsZ3RkWXhyRVE4CjlWN0tRWEJxQUFaMlBzSE1DZkF4aDNPMGxUST0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
    ca_crt: "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUZWekNDQXorZ0F3SUJBZ0lVWk9CSzNiallmUVkxOStFd0w4UmNKQnZVc004d0RRWUpLb1pJaHZjTkFRRUwKQlFBd096RUxNQWtHQTFVRUJoTUNRVlV4RXpBUkJnTlZCQWdNQ2xOdmJXVXRVM1JoZEdVeEZ6QVZCZ05WQkFvTQpEazVsZEdOeVlXTnJaWElnVEZSRU1CNFhEVEl6TVRBeE1URXdOVEUxTWxvWERUTXpNVEF3T0RFd05URTFNbG93Ck96RUxNQWtHQTFVRUJoTUNRVlV4RXpBUkJnTlZCQWdNQ2xOdmJXVXRVM1JoZEdVeEZ6QVZCZ05WQkFvTURrNWwKZEdOeVlXTnJaWElnVEZSRU1JSUNJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBZzhBTUlJQ0NnS0NBZ0VBdUpLdgpIWGU5S0Z2czdEZDZibkJnQ3ViYlQ5VmwxU2JGb0FJRFk2RGZqTGpNTHdSMEJJZGRrM3RnOW4wWjR5U0lCK1EvCjVRcmhnbVlDeGhWR0JBWnBnVStmU29BK05KR1F6UE1RbVpJdGtxbk1ZZWw2cnV1NmVIdXZnc09UeWRqMTFocGUKSVJEdVZvWEMyU1pvbm95TkUxYTlzbWdOQWdSOUEzeXdNK0tJVkVrMEZiSFF4TGZkN2ZtVEQ3ZVYyQjRxYmhYMQphMENvVGtxYkcwVDdzUldvRkJETlBQdkJtSnpYa0FXYXVySS9lKzBlc3FKakpLa0FaSXh2allWaG1saFdFcDdnCnNkZzNOYXc0NEIzckVlazI5NG9lTUhuc1pJeTVUdWx1RVdXMHMyWmtzRi82T1lrR1J1MHV3c3RqdU1DS0c4YzkKUVlmamFGZG5sMk9TZ1YzSW1reDZIeHovVEsyMzVnd0haNU5TREZFeSt0RUV6MDRBaVhmdU9zZzZLdEtIem5JQgpOMWZaK2xDaDhuUmt2aENRUmtJanByMHRnZmVINDZXc29QbjcxQmRMbFhTcTc4OUQ0ZEkwbCsrZSs1MmMyZGFFCitpaVJtWDRDbGkyS3ZmSWxYMk42STdTT0dMRFZxR2RVU2dhRHNKS3JNQkNwQjZRdTloRGlRdzAyaDFtM21jYXIKdDErYjFQN01NS2RLY1NrclZMd2pkU2ZEV2Q1S3hYVlkzaFoxMkg5dkRWZHpoVkR5N1VQWWhVcHE5RmZFTm1MYwpzUDZOaXVsbU5VRFhMdk12ZkJ2dS9USVI0SVMvbTZaUEM1OHdoUjhYamthTUo2SU0xd1lJcWNSeElYVW42WE81CjdFbEYzWFJ1a0dqUW4ydUc0NTZuRWNCdWF1Tlpob3ZESzRBWm9Wa0NBd0VBQWFOVE1GRXdIUVlEVlIwT0JCWUUKRlAwY1hZeUNnbllKTDVaRDZObnVjVUtIS0JFYk1COEdBMVVkSXdRWU1CYUFGUDBjWFl5Q2duWUpMNVpENk5udQpjVUtIS0JFYk1BOEdBMVVkRXdFQi93UUZNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dJQkFCQU41empnCmpIZTVrVjhzbmx3bzRHME5hMmZlSWtNRTVmbHhiR25vRXEyV0RscXZjYkdYVmRZVy92dGErYnkrT3hUdUlXL2QKbE84MExEY2RhUEhmaStZQXB4WXYzTFBLSWp1RHJqNmNnbFFDcEhoZGs2VXdmY0FHRmxVMmZucUZiNVBYZllaWQppNXJrL0JjQ1dFUlVMMnlpdlBGZnRZTjdTYzdmMmRlYkQwL3J1dk9MbDliR3ZJaFdmaXpGNi9hQlExa3c3Ym9TCmlTMm0vQ3I3L2l1T1Nta3RWdFBqMEN2Wmxpbkp1SnJyb242K3JCamhkZ3VzQXl2bTRXUjN1UUpHdm5IaDVrUnMKSURuZWFHOHl5UDlETStUU2pNMkJua1J6SjdUOXl2dUFwSzdjaE1ha0FjTklaR3VndXJuSUQ5NVBPMkM5eVE3bgp4VkJRSzlSR0YvcmRwdFV2Y0ZjTGJWbDJRd2YyMlhBbkwzNDNtQThIcVBsYUNZTmtGQUZodTJoYS9zVVpITE1UCnBSM1ZiMGR1ZEExZWtOMjE5RDRZMWJ5cHBKNlJjYk1DYWlQanZ0UFlXekcwREJhS0dTM2VFNUxHTjh5amNFSE4KeTY1MmN3N2VXOURhOGJ0MVlKaGZIREV4UFJRaUlCakdteis0bVAwNnNMNWNFSTFaVWJ0NHd2OU82SGJpMHFsdwp3MTlHdzBVN21ITS9CVGUwNkYwN241V0NScHJUdWYwZnYrNEI4ZXBtWkVKV0dyQ1FXSXpac0FtblJFUkh4WERiCjNuOEREazBHN041MEZjOHZiNWgrblMxT2xqcm1HMnRkTHF4TWkzZGhSazhlNFRma204LzBTYmZaNTRUSis1ZW4KM1N5Ri9aUm5qUk4yRUdRWjNLVUIzYS9Tc3p3SDJ5ekFTM0xyCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K" 
disasterRecovery:
  tls:
    certificates:
      tls_crt: "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUVzekNDQXB1Z0F3SUJBZ0lSQU9SZUxnUWNsUm1iclJLclR4VUxLRDR3RFFZSktvWklodmNOQVFFTEJRQXcKT3pFTE1Ba0dBMVVFQmhNQ1FWVXhFekFSQmdOVkJBZ01DbE52YldVdFUzUmhkR1V4RnpBVkJnTlZCQW9NRGs1bApkR055WVdOclpYSWdURlJFTUI0WERUSTBNREl4TlRBNU16YzFPRm9YRFRJMU1ESXhOREE1TXpjMU9Gb3dJakVnCk1CNEdBMVVFQXhNWGJXOXVaMjh0WkhJdFkyVnlkR2xtYVdOaGRHVXRZMjR3Z2dFaU1BMEdDU3FHU0liM0RRRUIKQVFVQUE0SUJEd0F3Z2dFS0FvSUJBUURJOGUwNGl3dHZkSFozUXlDeERuSUxNK0NqY0dHU2N4b01OTjE4dy9mRQpGR1pkdEY2NzF5RVIrRnJUclg1YXZCZGFBMlpyQkhGWXZnclc4RzAxV0RkeXJPVFA2NlYwdkRRMS8xNVhTazZwCjZMUlZJVitUaFBrd3dNNlE2Rk1DVy81d1U2SENDVnRPVW9UdW9aM2NQVXlIazV5RTJhV0YwR3I3dUdNSHVsZlIKQ2lFVTNKdmE5RE5sTEVoREpraW5XZjFrM281T3VXUWVCZjBHbWxTa3dLazMwaEFJQmh5cWptQVRsb3ZBTW9WZAp3ZG5SSy9tbnFMSXM1NzRBR3hQWVlyRjhFVzd5d040LzhURG9LaVB2Zm5aRVJ5Y2NqRTRyMHRTS1o1bXdhZmVBCkVVMDh1WFVIYmxxZWV5Sk5CMlMxYjBPblhnNVd4RXJXMjFkRmNFS1FKeSs1QWdNQkFBR2pnY293Z2Njd0RnWUQKVlIwUEFRSC9CQVFEQWdLa01BOEdBMVVkRXdFQi93UUZNQU1CQWY4d0hRWURWUjBPQkJZRUZER0x1Q2FDMzYwRwpTcitTWTdoRzRaUW1sSGVsTUI4R0ExVWRJd1FZTUJhQUZQMGNYWXlDZ25ZSkw1WkQ2Tm51Y1VLSEtCRWJNR1FHCkExVWRFUVJkTUZ1Q0NXeHZZMkZzYUc5emRJSWhiVzl1WjI5a1lpMWthWE5oYzNSbGNpMXlaV052ZG1WeWVTNXQKYjI1bmIyUmlnaVZ0YjI1bmIyUmlMV1JwYzJGemRHVnlMWEpsWTI5MlpYSjVMbTF2Ym1kdlpHSXVjM1pqaHdSLwpBQUFCTUEwR0NTcUdTSWIzRFFFQkN3VUFBNElDQVFBZ1JnNUROdmhHaWEwQ3dHV1NkcWVwTlJhNkZLQTlkb3VCClRLd3ZqTm0vaFJXc3FUSTQzWjZZVHV6R25GU3NvY2xyOURpQ3VmTXBGcUVmeDlnVDNENDVoSmFocG5MWVh0RTUKRE93M1VJVDBrbW15cVplNXlyK1ZIR2JzL21WOU1wTWhTdUFLV3NpellhQ0FzeU5qTENzSHFvTzFFZCtSOXVqTQpSc1M1cHJBS3JJRFp0eFVUQm5BRlpmQ1pGS21TUzFpamlPTUxqdG14VURIVXZyT3BzWWNDUFVER1VPbWtsVEs0CjRZeWdENnN1eTNiOVM0Mnk1N0JEa2FxbXJWdmovcTZ1VCs0R1Bvb0h5K0xSTWo3L2loRi85V2J0eHhMV0xDSVAKRGMvVVN5NWpyNUx3ckI1a25aYWVoVW1vQTFhMkh0MitlZHZ3QnJzTFBoZ01DOHMyVHROUmkxcVB6NjJFeENLbgpxRnJSejBCb3BwbTlDUkg2bzlLTC9ITk4ycmZyaDgvM2YrWEwvbkN0bStIRUpMTWp3WDQ2US9Ka1NZZm5ick5FCjZ2eHY5L0t1SXpyYTdkYjhkN2lKeVlvby91SlZ1eXNJSGJqTGl1VVVxOThoVDRsZGhaUG5zdjM2WWJjN1dSK1kKRjFNNitEa1NwR2JhUEVuSm9OMXhqTytyNUlmcW9ybXVGVDRncWIrQWJwRUhpTnpBSmQ0cExuMXdJakRieW1scgpGOUdhdFpxZXRBbDI2K2NDblhIMGtNMko2SHJrOElrK2xNNFh1b0VuUWZlYWlPanBHaU9SQ0NRMXRMOEQycGhYCjBXUmkvb0cwZUtRcFpKUnlVV2FZNExKR0pxOVdGS2lydWNpSSszTFFKOHFGN2YxM0lLcE5CdFdsQlBvZnBtRXAKcUxGWFpHZ0RCUT09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K"
      tls_key: "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcFFJQkFBS0NBUUVBeVBIdE9Jc0xiM1IyZDBNZ3NRNXlDelBnbzNCaGtuTWFERFRkZk1QM3hCUm1YYlJlCnU5Y2hFZmhhMDYxK1dyd1hXZ05tYXdSeFdMNEsxdkJ0TlZnM2Nxemt6K3VsZEx3ME5mOWVWMHBPcWVpMFZTRmYKazRUNU1NRE9rT2hUQWx2K2NGT2h3Z2xiVGxLRTdxR2QzRDFNaDVPY2hObWxoZEJxKzdoakI3cFgwUW9oRk55YgoydlF6WlN4SVF5WklwMW45Wk42T1RybGtIZ1g5QnBwVXBNQ3BOOUlRQ0FZY3FvNWdFNWFMd0RLRlhjSFowU3Y1CnA2aXlMT2UrQUJzVDJHS3hmQkZ1OHNEZVAvRXc2Q29qNzM1MlJFY25ISXhPSzlMVWltZVpzR24zZ0JGTlBMbDEKQjI1YW5uc2lUUWRrdFc5RHAxNE9Wc1JLMXR0WFJYQkNrQ2N2dVFJREFRQUJBb0lCQVFEQTd1ZmhQaitBaDhXbQp4S0VDM3VmSXNjcWhvaWxNdjQ3bTRXczNlOERNVnZuaVJtZ2Uybk02R2NhN0x2eitpVkd5YjBsS3Z6MUZBMUxOCkJKTVdnTmpjRmZ5clZZbkxCMWpwNzRMWk5OTktkOCtOWFRtekhoMVVIZ3MzUHBsVXpwY0Jxb3JKRHNySDdKc04KczhjcHl3Rkx1d0t3MjNmOWZ4cjVEUlNvT3RaT01VTGI1dnUyWDkycWxPRXVmRXZDVWs1OTZHeWtZbS9MdXlRLwpOL3RlRXV5OWg0ajcvZjZ4WlZicHU4azBlYkVlT29ldUszSjVXeUtJN0NLS1lhelRXa0dZRDdPNkxYVGI5OTA3Cm96L0ErRmZtQ3VqSUZKZTZTOFVHS3Jua1daNEdGak1zRHhBUmRCWjlPSjVSZFg3azF3Q1VIUTErMnhIejNPKzcKQStQdFNzR3hBb0dCQVBQWlRlT0doOHhzdFo4Tnh2UzVhT0FESUhwTmM3S3YyemZBeFAvRG1iOTA2NVV0azZydwpSRDRsQ1lqY1FqVlllaGs1WFF0dm5UMDd0eWN3UktYeEJWMHNIZTRJT2htMExOYXYzM0VzV1pNeE9hZ05ob0xWClR4QnpMMm1Fdlo4TEtPVXU1ZkZ6ZGNLZjcrWGgyYVBhb01mckZEcG9Zc3pBZHYrMDZtMmdmcnZiQW9HQkFOTDEKVUNwM24vMUhvL2UzQktSb2RSY2h3bFRhR3YyTXB0K0hYYjRjelR2dzcreTBrUFpzMDcyb1hqcWdXK1FQMStFOApLTEZ2Q3dCb0d6MkdVQWRrWmZ1cEcrVkFZWkg2WkZsOVZyWHZTMEd5R1haUzZSYWo4bEM0SWtSemhlNGdUT3hoCnVVdno1ejhOdzJvUzdZRDBmckFvQjMxNFRCc1JJOTRoa2llN1B3RDdBb0dCQUxsemdLc1RlMS9iSlYycnFxNGYKL0VTeDNCZG5wQ0EzWWk5S3FnZ2lDR0gxVjkyQ1poWFEyUFd5VVVnR3kwdXEyR0Vxb1RxN1RnaHR5K00vOEZXTApzaHFrSExjVkJxclp2bWdnSlh6Nno3MEQ2T2VJTWM1Nno4Q2crV1Awa2duTkFQTWI4Y0Rwb0p1OTYwTVh1dC9FCnZCYVBFRGxEZmpCZUI2SjlRdlRRNU5HVkFvR0FSbndTYWU0SU5hOGZHT0E4bTlZTzhVaWxUb2FGS0J3N2tVb0EKUjBvR1JMWE81QzY4bEtsdDRkdUVpR0FWODlCYlYvVXF2NFlUamZJNno4YTFySktlQklUUFBqelJuTjJsYzhVTwpHTUc0U2w0QVplbHoyYzJ6WThieUpCN1pLK1A4NzZvREtGNTQ4RGRnQ0d3RWtPYWdBYW1PUHh6WGlOK2tOVTdRCkw1Zy9oOHNDZ1lFQXM2dnBiOHlPcEJGd1dyZ1pkUDRqekZLNWFxeWFoUXV1aGkwQmo4bnEzQ1hyUlppWXBlcHoKZmZmYVJwU2hZcVVHN3phKzVDWDhOY0Ywc0xPTXQzbE5PaEZHM0F1Y2Z2T0M1eXB5MjA2TFpLbFhFQmhTK3Q5cgppOUNkNlRFTnZkQlNNYVNHV2Q3ekpsS3ViZ21xVklMM3hpVmdMVlNnRnMrUWJ5a0hxYmR1REFjPQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo="
```

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

### Upgrade Using App Deployer

 Set **DEPLOY_MODE** to **Rolling Update** and repeat the steps from [Deployment Using App Deployer](#deployment-using-app-deployer).

## Upgrade Legacy Mongodb-Cluster-Installation-Scripts With Operator

The information about upgrading the legacy mongodb-cluster-installation-scripts with operator is specified below.

### Prerequisites

The prerequisites are as follows:

* DVM with access to the cloud.
* OpenShift Client (oc) is installed to the DVM.

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
  
7. Run the App Deployer in the **Rolling Update** mode. For example of parameters, see [Parameters' Examples](#parameters-examples).
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
