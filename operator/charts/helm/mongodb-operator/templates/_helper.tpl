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
DNS names used to generate SSL certificate with "Subject Alternative Name" field
*/}}
{{- define "mongodb.generateCNFRSFQDNs" -}}
  {{- $size := int .Values.schemaSettings.cnfReplicaSize -}}
  {{- $namespace := .Release.Namespace -}}
  {{- $domain := .Values.schemaSettings.thisDomainName | default "cluster.local" -}}
  {{- $fqdnList := list -}}
  {{- range $i := until $size -}}
  {{- $fqdn := printf "cnfrs%d-0.cnfrs.%s.svc.%s" $i $namespace $domain -}}
  {{- $fqdnList = append $fqdnList $fqdn -}}
  {{- end -}}
  {{- $fqdnList | toYaml -}} 
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


{{- define "nosql.core.secret" -}}
{{ $_ := set . "userEnv" "" }}
{{ $_ := set . "passEnv" "" }}
{{ include "nosql.core.secret.fromEnv" $_ }}
{{- end -}}

{{/*
[NoSQL Operator Core] Secret template

Arguments:
Dictionary with:
1. "secret" section includes next elements:
    .secretName (required)
    .password (required)
    .username (optional)
    .role (optional)
    .authDb (optional)

Usage example:
{{template "nosql.core.secret" (dict "secret" (dict "secretName" "mongodb-root-credentials.v1" "password" .Values.mongodb.rootPassword) "userEnv" .Values.INFRA_MONGO_DB_USERNAME "passEnv" .Values.INFRA_MONGO_DB_PASSWORD)}}
*/}}
{{- define "nosql.core.secret.fromEnv" -}}
apiVersion: v1
kind: Secret
metadata:
  name: {{ .secret.secretName }}
stringData:
  password: {{ include "fromEnv" (dict "envName" .passEnv "default" .secret.password) | quote }}
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

{{template "nosql.core.secret.internal" (dict "secret" .Values.redis)}}
*/}}
{{- define "nosql.core.secret.internal" -}}
{{ include "nosql.core.secret" . }}
{{- end -}}


{{- define "mongodb.generateDATARSFQDNs" -}}
  {{- $size := int .Values.schemaSettings.dataReplicaSize -}}
  {{- $shCount := int .Values.schemaSettings.shardCount -}}
  {{- $namespace := .Release.Namespace -}}
  {{- $domain := .Values.schemaSettings.thisDomainName | default "cluster.local" -}}
  {{- $fqdnList := list -}}
  {{- range $sh := until $shCount -}}
    {{- range $r := until $size -}}
      {{- $fqdn := printf "datars%d%d-0.datars%d.%s.svc.%s" (add1 $sh) $r (add1 $sh) $namespace $domain -}}
      {{- $fqdnList = append $fqdnList $fqdn -}}
    {{- end -}}
  {{- end -}}
  {{- $fqdnList | toYaml -}} 
{{- end -}}

{{- define "mongodb.certDnsNames" -}}
  {{- $dnsNames := list "localhost" "mongodb-cluster" (printf "%s.%s" "mongos" .Release.Namespace) (printf "%s.%s.svc" "mongos" .Release.Namespace) -}}
  {{- $dnsNames = concat $dnsNames (default list .Values.tls.generateCerts.subjectAlternativeName.additionalDnsNames) -}}
  {{- $dnsNames | toYaml -}} 
{{- end -}}


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

{{/*
Service Account for Site Manager depending on smSecureAuth
*/}}
{{- define "disasterRecovery.siteManagerServiceAccount" -}}
  {{- if .Values.disasterRecovery.httpAuth.smServiceAccountName -}}
    {{- .Values.disasterRecovery.httpAuth.smServiceAccountName -}}
  {{- else -}}
    {{- if .Values.disasterRecovery.httpAuth.smSecureAuth -}}
      {{- "site-manager-sa" -}}
    {{- else -}}
      {{- "sm-auth-sa" -}}
    {{- end -}}
  {{- end -}}
{{- end -}}

{{- define "mongodb.defaultLabels" -}}
{{- if .Values.ARTIFACT_DESCRIPTOR_VERSION }}
app.kubernetes.io/version: {{ default "" .Values.ARTIFACT_DESCRIPTOR_VERSION | trunc 63 | trimAll "-_." }}
{{- end }}
app.kubernetes.io/part-of: {{ default "mongodb" .Values.PART_OF }}
app.kubernetes.io/managed-by: {{ default "operator" .Values.MANAGED_BY }}
{{- end -}}

{{- define "mongo.monitoredImages" -}}
  {{- if .Values.deployDescriptor -}}
    {{- printf "deployment mongodb-operator mongodb-operator %s, " (include "find_image" (dict "deployName" "dockerMongoOperator" "SERVICE_NAME" "dockerMongoOperator" "vals" .Values "default" "not_found")) -}}
    {{- if and (not (eq .Values.schemaSettings.schemaType "single")) .Values.mongodb.install -}}
      {{- printf "statefulset datars10 datars10 %s, " (include "find_image" (dict "deployName" "dockerMongodb" "SERVICE_NAME" "dockerMongodb" "vals" .Values "default" "not_found")) -}}
    {{- end -}}
    {{- if or (eq .Values.schemaSettings.schemaType "ha") (eq .Values.schemaSettings.schemaType "dr") -}}
      {{- if .Values.schemaSettings.sharded -}}
        {{- printf "statefulset cnfrs0 cnfrs0 %s, " (include "find_image" (dict "deployName" "dockerMongodb" "SERVICE_NAME" "dockerMongodb" "vals" .Values "default" "not_found")) -}}
      {{- end -}}
    {{- end -}}
  {{- end -}}
{{- end -}}

{{- define "mongodb.x509.generateCNFRSFQDNs" -}}
{{- $size := int .Values.schemaSettings.cnfReplicaSize -}}
{{- $namespace := .Release.Namespace -}}
{{- $domain := .Values.schemaSettings.thisDomainName | default "cluster.local" -}}

- localhost
{{- range $i := until $size }}
- {{ printf "cnfrs%d-0.cnfrs.%s.svc.%s" $i $namespace $domain }}
{{- end }}

{{- end }}


{{- define "mongodb.x509.generateShardFQDNs" -}}
{{- $shard := .shard -}}
{{- $ctx := .context -}}

{{- $size := int $ctx.Values.schemaSettings.dataReplicaSize -}}
{{- $namespace := $ctx.Release.Namespace -}}
{{- $domain := $ctx.Values.schemaSettings.thisDomainName | default "cluster.local" -}}

- localhost
{{- range $r := until $size }}
- {{ printf "datars%d%d-0.datars%d.%s.svc.%s" $shard $r $shard $namespace $domain }}
{{- end }}

{{- end }}

{{- define "mongodb.x509.generateMongosFQDNs" -}}
{{- $namespace := .Release.Namespace -}}
{{- $domain := .Values.schemaSettings.thisDomainName | default "cluster.local" -}}

- localhost
- {{ printf "mongos.%s.svc.%s" $namespace $domain }}

{{- end }}