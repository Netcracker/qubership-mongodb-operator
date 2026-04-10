{{- define "find_image" -}}
  {{- $image := .default -}}

  {{- if .vals.deployDescriptor -}}
    {{- if index .vals.deployDescriptor .deployName -}}
      {{- $image = (index .vals.deployDescriptor .deployName "image") -}}
    {{- else if index .vals.deployDescriptor .SERVICE_NAME -}}
      {{- $image = (index .vals.deployDescriptor .SERVICE_NAME "image") -}}
    {{- end -}}
  {{- end -}}

  {{ printf "%s" $image }}
{{- end -}}

{{/*
Common Mongodb resources labels
*/}}
{{- define "mongodb.defaultLabels" -}}
{{- if .Values.ARTIFACT_DESCRIPTOR_VERSION }}
app.kubernetes.io/version: {{ default "" .Values.ARTIFACT_DESCRIPTOR_VERSION | trunc 63 | trimAll "-_." }}
{{- end }}
app.kubernetes.io/part-of: {{ default "mongodb-services" .Values.PART_OF }}
app.kubernetes.io/managed-by: {{ default "operator" .Values.MANAGED_BY }}
app.kubernetes.io/managed-by-operator: "mongodb-services-operator"
{{- end -}}

{{/*
Configure MongoDB service 'enableDisasterRecovery' property
*/}}
{{- define "mongodb.enableDisasterRecovery" -}}
  {{- if or (eq .Values.disasterRecovery.mode "active") (eq .Values.disasterRecovery.mode "standby") (eq .Values.disasterRecovery.mode "disabled") -}}
    {{- printf "true" }}
  {{- else -}}
    {{- printf "false" }}
  {{- end -}}
{{- end -}}

{{/*
Find a MongoDB disaster recovery service operator image in various places.
Image can be found from:
* SaaS/App deployer (or groovy.deploy.v3) from .Values.disasterRecoveryImage
* DP.Deployer from .Values.deployDescriptor.disasterRecoveryImage.image
* or from default values .Values.global.disasterRecovery.image
*/}}
{{- define "disasterRecovery.image" -}}
  {{- if .Values.deployDescriptor -}}
    {{- if .Values.disasterRecoveryImage -}}
      {{- printf "%s" .Values.disasterRecoveryImage -}}
    {{- else -}}
      {{- printf "%s" (index .Values.deployDescriptor.disasterRecoveryImage.image) -}}
    {{- end -}}
  {{- else -}}
    {{- printf "%s" .Values.disasterRecovery.image -}}
  {{- end -}}
{{- end -}}

{{/*
Configure MongoDB statefulset names in disaster recovery health check format.
*/}}
{{- define "mongodb.statefulsetNames" -}}
    {{- $cnfs := .Values.schemaSettings.cnfReplicaSize }}
    {{- $datas := .Values.schemaSettings.dataReplicaSize }}
    {{- $shards := .Values.schemaSettings.shardCount }}
    {{- $lst := list }}
    {{- range $i, $e := until ($shards | int) }}
        {{- $tmpTemplate := (printf "statefulset datars%d" (add $i 1))}}
        {{- range $j, $k := until ($datas | int) }}
            {{- $lst = append $lst (printf "%s%d" $tmpTemplate (add $j 0))  }}
        {{- end }}
    {{- end }}
    {{- range $i, $e := until ($cnfs | int) }}
        {{- $lst = append $lst (printf "statefulset cnfrs%d" (add $i 0))  }}
    {{- end }}
    {{- join "," $lst }}
{{- end -}}
[MongoDB Operator] Get Vault kv path to secret
Arguments:
Dictionary with:
1. "cloudPublicHost" - Server Hostname
2. "namespace" - Namespace
2. "serviceAccount" - Service Account
3. "secretName" - Name of the secret
Usage example:
{{include "vault_kv_path" (dict "cloudPublicHost" .Values.CLOUD_PUBLIC_HOST "namespace" .Release.Namespace "serviceAccount" .Values.operator.operatorName "secretName" "sample-secret") }}
*/}}
{{- define "vault_kv_path" -}}
    {{ printf "secret/nc-%s/%s/%s/%s#password" .cloudPublicHost .namespace .serviceAccount .secretName }}
{{- end -}}

{{/*
[MongoDB Operator] Get Vault DB Engine path
Arguments:
Dictionary with:
1. "cloudPublicHost" - Server Hostname
2. "namespace" - Namespace
2. "serviceAccount" - Service Account
3. "username" - Username
Usage example:
{{include "vault_db_path" (dict "cloudPublicHost" .Values.CLOUD_PUBLIC_HOST "namespace" .Release.Namespace "serviceAccount" .Values.operator.operatorName "username" "sample-user") }}
*/}}
{{- define "vault_db_path" -}}
    {{ printf "database/static-creds/nc-%s_%s_%s_%s#password" .cloudPublicHost .namespace .serviceAccount .username }}
{{- end -}}


{{- define "nosql.core.secret.vault" -}}
{{ $_ := set . "userEnv" "" }}
{{ $_ := set . "userPass" "" }}
{{include "nosql.core.secret.vault.fromEnv" $_ }}
{{- end -}}

{{/*
[NoSQL Operator Core] Vault secret template
Arguments:
Dictionary with:
1. "vlt" - .vaultRegistration section
3. "secret" section includes next elements:
    .secretName (required)
    .password (required)
    .username (optional)
    .role (optinal)
    .authDb (optional)
    .vaultPasswordPath (optional)
4. "isInternal" is a required boolean parameter
Usage example:
{{template "nosql.core.secret.vault" (dict "vlt" .Values.vaultRegistration "secret" (dict "secretName" "mongodb-root-credentials.v1" "password" .Values.mongodb.rootPassword) "userEnv" .Values.INFRA_MONGO_DB_USERNAME "passEnv" .Values.INFRA_MONGO_DB_PASSWORD))}}
*/}}
{{- define "nosql.core.secret.vault.fromEnv" -}}
apiVersion: v1
kind: Secret
metadata:
  name: {{ .secret.secretName }}
stringData:
  {{- if .vlt.enabled }}
    {{- if .secret.vaultPasswordPath }}
  password: {{ .secret.vaultPasswordPath | quote }}
    {{- else }}
        {{- if (.isInternal) }}
            {{- if (.secret.db) }}
  password: 'vault:{{include "vault_db_path" (set . "username" .secret.username) }}'
            {{- else }}
  password: 'vault:{{include "vault_kv_path" (set . "secretName" .secret.secretName) }}'
            {{- end }}
        {{- else }}
  password: {{ include "fromEnv" (dict "envName" .passEnv "default" .secret.password) | quote }}
        {{- end }}
    {{- end }}
  {{- if .secret.nonVaultPassword }}
  nonVaultPassword: {{ include "fromEnv" (dict "envName" .passEnv "default" .secret.password) | quote }}
  {{- end }}
  {{- else }}
  password: {{ include "fromEnv" (dict "envName" .passEnv "default" .secret.password) | quote }}
  {{- end }}
  {{- if .secret.username }}
  username: {{ include "fromEnv" (dict "envName" .userEnv "default" .secret.username) | quote }} 
  {{- end }}
  {{- if .secret.role }}
  role: {{ .secret.role | quote }}
  {{- end }}
  {{- if .secret.authDb }}
  auth-database: {{ .secret.authDb | quote }}
  {{- end }}
type: Opaque
{{- end -}}

{{/*
[NoSQL Operator Core] Internal secret template
{{template "nosql.core.secret.internal" (dict "vlt" .Values.vaultRegistration "secret" .Values.redis)}}
*/}}
{{- define "nosql.core.secret.internal" -}}
{{include "nosql.core.secret.vault" (set . "isInternal" true)}}
{{- end -}}

{{/*
[NoSQL Operator Core] External secret template
{{template "nosql.core.secret.external" (dict "vlt" .Values.vaultRegistration "secret" .Values.redis)}}
*/}}
{{- define "nosql.core.secret.external" -}}
{{include "nosql.core.secret.vault" (set . "isInternal" false)}}
{{- end -}}

{{/*
[MongoDB Operator Core] from env of from values
Dictionary with:
1. "envName" - name of env var to get value from
2.  "default" - default value from values.yaml
{{template "fromEnv" (dict "envName" ".Values.VAULT_ADDR" "default" .Values.vaultRegistration.token) }}
*/}}
{{- define "fromEnv" -}}
  {{- $envValue := .envName -}}
  {{- if and (ne ($envValue | toString) "<nil>") (ne ($envValue | toString) "") -}}
    {{- .envName -}}
  {{- else -}}
    {{- .default -}}
  {{- end -}}
{{- end -}}

{{/*
Dictionary with:
Uses value from values.yaml if defined, otherwise value from environment variable if defined, else - default
1. "dotVar" - parameter defined with dots like dbaas.install
2. "enVar" - parameter defined as environment variable like DBAAS_ENABLED
3.  "default" - default value
{{template "fromValuesThenEnvElseDefault" (dict "dotVar" .Values.dbaas.install "envVar" .Values.DBAAS_ENABLED "default" true ) }}
*/}}
{{- define "fromValuesThenEnvElseDefault" -}}
  {{- if and (ne (.dotVar | toString) "<nil>") (ne (.dotVar | toString) "") -}}
    {{- .dotVar -}}
  {{- else if and (ne (.envVar | toString) "<nil>") (ne (.envVar | toString) "") -}}
    {{- .envVar -}}
  {{- else -}}
    {{- .default -}}
  {{- end -}}
{{- end -}}

{{/*
[MongoDB Operator Core] from env of from values
Dictionary with:
1. "envName" - name of env var to get value from
2.  "default" - default value from values.yaml
{{template "ifEnvThenDefault" (dict "envName" .Values.VAULT_ADDR "then" (printf %s_%s .Values.VAULT_ADDR "const" ) "default" .Values.vaultRegistration.token) }}
*/}}
{{- define "ifEnvThenDefault" -}}
  {{- $value := .default -}}
  {{- if .envName -}}
    {{- $value = .then -}}
  {{- else -}}
    {{- $value = .default -}}
  {{- end -}}
  {{- if $value -}}
  {{ printf "%s" $value }}
  {{- end -}}
{{- end -}}

{{/*
DNS names used to generate SSL certificate with "Subject Alternative Name" field
*/}}
{{- define "dbaasAdapter.certDnsNames" -}}
  {{- $dnsNames := list "localhost" (printf "%s.%s" "dbaas-mongo-adapter" .Release.Namespace) (printf "%s.%s.svc" "dbaas-mongo-adapter" .Release.Namespace) -}}
  {{- $dnsNames = concat $dnsNames .Values.tls.generateCerts.subjectAlternativeName.additionalDnsNames -}}
  {{- $dnsNames | toYaml -}}
{{- end -}}
{{- define "backupDaemon.certDnsNames" -}}
  {{- $dnsNames := list "localhost" "mongodb-backup-daemon" (printf "%s.%s" "mongodb-backup-daemon" .Release.Namespace) (printf "%s.%s.svc" "mongodb-backup-daemon" .Release.Namespace) -}}
  {{- $dnsNames = concat $dnsNames .Values.tls.generateCerts.subjectAlternativeName.additionalDnsNames -}}
  {{- $dnsNames | toYaml -}}
{{- end -}}
{{- define "mongodb.certDnsNames" -}}
  {{- $dnsNames := list "localhost" "mongodb-cluster" (printf "%s.%s" "mongos" .Release.Namespace)  (printf "%s.%s.svc" "mongos" .Release.Namespace) -}}
  {{- $dnsNames = concat $dnsNames .Values.tls.generateCerts.subjectAlternativeName.additionalDnsNames -}}
  {{- $dnsNames | toYaml -}}
{{- end -}}
{{- define "mongodr.certDnsNames" -}}
  {{- $dnsNames := list "localhost" (printf "%s.%s" "mongodb-disaster-recovery" .Release.Namespace) (printf "%s.%s.svc" "mongodb-disaster-recovery" .Release.Namespace) -}}
  {{- $dnsNames = concat $dnsNames .Values.tls.generateCerts.subjectAlternativeName.additionalDnsNames -}}
  {{- $dnsNames | toYaml -}}
{{- end -}}
{{/*
IP addresses used to generate SSL certificate with "Subject Alternative Name" field
*/}}
{{- define "common.certIpAddresses" -}}
  {{- $ipAddresses := list "127.0.0.1" -}}
  {{- $ipAddresses = concat $ipAddresses .Values.tls.generateCerts.subjectAlternativeName.additionalIpAddresses -}}
  {{- $ipAddresses | toYaml -}}
{{- end -}}

{{/*
[NoSQL Operator Core] PodDisruptionBudget
Dictionary with:
1. "name" - pdb name
2. "labels" - label selectors map
3. "minAvailable" - desired pods count
{{template "nosql.core.pdb" (dict "name" "mongodb" "labels" $labels "minAvailable" $minAvailable)}}
*/}}
{{- define "nosql.core.pdb" -}}
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: {{ .name | quote }}
  labels:
    app.kubernetes.io/part-of: "mongodb"
    app.kubernetes.io/managed-by: "operator"
spec:
  minAvailable: {{ .minAvailable }}
  selector:
    matchLabels:
      {{- range $k, $v := .labels }}
      {{ $k | quote }}: {{ $v | quote }}
      {{- end }}
{{- end -}}

{{/*
TLS Static Metric secret template
Arguments:
Dictionary with:
* "namespace" is a namespace of application
* "application" is name of application
* "service" is a name of service
* "enabledSsl" is ssl enabled for service
* "secret" is a name of tls secret for service
* "certProvider" is a type of tls certificates provider
* "certificate" is a name of CertManger's Certificate resource for service
Usage example:
{{template "global.tlsStaticMetric" (dict "namespace" .Release.Namespace "application" .Chart.Name "service" .global.name "enabledSsl" (include "global.sslEnabled" .) "secret" (include "global.sslSecretName" .) "certProvider" (include "services.certProvider" .) "certificate" (printf "%s-tls-certificate" (include "global.name")) }}
*/}}
{{- define "global.tlsStaticMetric" -}}
- expr: {{ ternary "1" "0" .enabledSsl }}
  labels:
    namespace: "{{ .namespace }}"
    application: "{{ .application }}"
    service: "{{ .service }}"
    {{ if .enabledSsl }}
    secret: "{{ .secret }}"
    {{ if eq .certProvider "cert-manager" }}
    certificate: "{{ .certificate }}"
    {{ end }}
    {{ end }}
  record: service:tls_status:info
{{- end -}}


{{- define "getMongosResourcesForProfile" -}}
  {{- $flavor := .dotVar }}
{{- if and (ne (.envVar | toString) "<nil>") (ne (.envVar | toString) "") -}}
  {{- $flavor = .envVar -}}
{{- end -}}
  {{- if eq $flavor "small" -}}
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: 500m
        memory: 512Mi
  {{- else if eq $flavor "medium" -}}
      requests:
        cpu: 500m
        memory: 256Mi
      limits:
        cpu: 2
        memory: 2Gi
  {{- else if eq $flavor "large" -}}
      requests:
        cpu: 1
        memory: 1Gi
      limits:
        cpu: 4
        memory: 4Gi
  {{- else if $flavor -}}
  {{- fail "value for .Values.global.profile is not one of  `small`, `medium`, `large`" }}
  {{- else -}}
      limits:
        memory: {{ .values.mongodb.mongosResources.limits.memory }}
        cpu: {{ .values.mongodb.mongosResources.limits.cpu | quote }}
      requests:
        memory: {{ .values.mongodb.mongosResources.requests.memory }}
        cpu: {{ .values.mongodb.mongosResources.requests.cpu | quote }}
  {{- end -}}
{{- end -}}

{{- define "getDatarsResourcesForProfile" -}}
  {{- $flavor := .dotVar }}
{{- if and (ne (.envVar | toString) "<nil>") (ne (.envVar | toString) "") -}}
  {{- $flavor = .envVar -}}
{{- end -}}
  {{- if eq $flavor "small" -}}
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: 500m
        memory: 512Mi
  {{- else if eq $flavor "medium" -}}
      requests:
        cpu: 500m
        memory: 256Mi
      limits:
        cpu: 2
        memory: 2Gi
  {{- else if eq $flavor "large" -}}
      requests:
        cpu: 1
        memory: 1Gi
      limits:
        cpu: 4
        memory: 4Gi
  {{- else if $flavor -}}
  {{- fail "value for .Values.global.profile is not one of  `small`, `medium`, `large`" }}
  {{- else -}}
      limits:
        memory: {{ .values.mongodb.dataResources.limits.memory }}
        cpu: {{ .values.mongodb.dataResources.limits.cpu | quote }}
      requests:
        memory: {{ .values.mongodb.dataResources.requests.memory }}
        cpu: {{ .values.mongodb.dataResources.requests.cpu | quote }}
  {{- end -}}
{{- end -}}

{{- define "getCnfrsResourcesForProfile" -}}
  {{- $flavor := .dotVar }}
{{- if and (ne (.envVar | toString) "<nil>") (ne (.envVar | toString) "") -}}
  {{- $flavor = .envVar -}}
{{- end -}}
  {{- if eq $flavor "small" -}}
      requests:
        cpu: 100m
        memory: 128Mi
      limits:
        cpu: 500m
        memory: 512Mi
  {{- else if eq $flavor "medium" -}}
      requests:
        cpu: 300m
        memory: 128Mi
      limits:
        cpu: 1
        memory: 1Gi
  {{- else if eq $flavor "large" -}}
      requests:
        cpu: 300m
        memory: 512Mi
      limits:
        cpu: 2
        memory: 2Gi
  {{- else if $flavor -}}
  {{- fail "value for .Values.global.profile is not one of  `small`, `medium`, `large`" }}
  {{- else -}}
      limits:
        memory: {{ .values.mongodb.cnfResources.limits.memory }}
        cpu: {{ .values.mongodb.cnfResources.limits.cpu | quote }}
      requests:
        memory: {{ .values.mongodb.cnfResources.requests.memory }}
        cpu: {{ .values.mongodb.cnfResources.requests.cpu | quote }}
  {{- end -}}
{{- end -}}


{{- define "getBackupResourcesForProfile" -}}
  {{- $flavor := .dotVar }}
{{- if and (ne (.envVar | toString) "<nil>") (ne (.envVar | toString) "") -}}
  {{- $flavor = .envVar -}}
{{- end -}}
  {{- if eq $flavor "small" -}}
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: 500m
        memory: 512Mi
  {{- else if eq $flavor "medium" -}}
      requests:
        cpu: 500m
        memory: 256Mi
      limits:
        cpu: 1
        memory: 1Gi
  {{- else if eq $flavor "large" -}}
      requests:
        cpu: 500m
        memory: 512Mi
      limits:
        cpu: 2
        memory: 2Gi
  {{- else if $flavor -}}
  {{- fail "value for .Values.global.profile is not one of  `small`, `medium`, `large`" }}
  {{- else -}}
      limits:
        memory: {{ .values.backup.backupResources.limits.memory }}
        cpu: {{ .values.backup.backupResources.limits.cpu | quote }}
      requests:
        memory: {{ .values.backup.backupResources.requests.memory }}
        cpu: {{ .values.backup.backupResources.requests.cpu | quote }}
  {{- end -}}
{{- end -}}

{{- define "mongodbSupplementary.monitoredImages" -}}
  {{- if .Values.deployDescriptor -}}
    {{- if .Values.robotTests.install -}}
      {{- printf "deployment robot-tests robot-tests %s, " (include "find_image" (dict "deployName" "mongo-tests" "SERVICE_NAME" "mongo-tests" "vals" .Values "default" "not_found")) -}}
    {{- end -}}
    {{- if .Values.prometheusExporter.install -}}
      {{- printf "deployment mongodb-prometheus-exporter mongodb-prometheus-exporter %s, " (include "find_image" (dict "deployName" "mongodb-prometheus-exporter" "SERVICE_NAME" "mongodb-prometheus-exporter" "vals" .Values "default" "not_found")) -}}
    {{- end -}}
    {{- if .Values.dbaas.install -}}
      {{- printf "deployment dbaas-mongo-adapter dbaas-mongo-adapter %s, " (include "find_image" (dict "deployName" "dbaas_mongo" "SERVICE_NAME" "dbaas_mongo" "vals" .Values "default" "not_found")) -}}
    {{- end -}}
    {{- if .Values.backup.install -}}
      {{- printf "deployment mongodb-backup-daemon mongodb-backup-daemon %s, " (include "find_image" (dict "deployName" "backupDaemon" "SERVICE_NAME" "backupDaemon" "vals" .Values "default" "not_found")) -}}
    {{- end -}}
    {{- printf "deployment  %s  %s %s, " .Values.operator.operatorName .Values.operator.operatorName (include "find_image" (dict "deployName" "dockerMongodbService" "SERVICE_NAME" "dockerMongodbService" "vals" .Values "default" "not_found")) -}}
  {{- end -}}
{{- end -}}

{{/*
Return name for gateway
*/}}
{{- define "gateway.name" -}}
{{- if and .Values.GATEWAY_SYSTEM_NAME .Values.global.cloudIntegrationEnabled }}
{{- .Values.GATEWAY_SYSTEM_NAME }}
{{- else }}
{{- default "default-external-gateway" .Values.dbaas.gatewayAPI.gatewayName }}
{{- end -}}
{{- end -}}

{{/*
Return namespace for gateway
*/}}
{{- define "gateway.namespace" -}}
{{- if and .Values.GATEWAY_SYSTEM_NAMESPACE .Values.global.cloudIntegrationEnabled }}
{{- .Values.GATEWAY_SYSTEM_NAMESPACE }}
{{- else }}
{{- default "gateway-system" .Values.dbaas.gatewayAPI.gatewayNamespace }}
{{- end -}}
{{- end -}}

{{/*
Backup Daemon SSL secret name
*/}}
{{- define "getBackupSslSecretName" -}}
  {{- if .Values.backup.s3.sslCert -}}
    {{- if .Values.backup.s3.sslSecretName -}}
      {{- .Values.backup.s3.sslSecretName -}}
    {{- else -}}
      {{- printf "backup-daemon-s3-tls-secret" -}}
    {{- end -}}
  {{- else -}}
    {{- if .Values.backup.s3.sslSecretName -}}
      {{- .Values.backup.s3.sslSecretName -}}
    {{- else -}}
      {{- printf "" -}}
    {{- end -}}
  {{- end -}}
{{- end -}}
