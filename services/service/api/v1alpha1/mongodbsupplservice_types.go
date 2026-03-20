/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SchemaType string

const (
	HA      SchemaType = "ha"
	DR      SchemaType = "dr"
	Arbiter SchemaType = "arbiter"
	Single  SchemaType = "single"
)

type Policies struct {
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
}

// Schema of deploy
type SchemaSettings struct {
	SchemaType      SchemaType `json:"schemaType,omitempty"`
	Sharded         bool       `json:"sharded,omitempty"`
	CnfReplicaSize  int        `json:"cnfReplicaSize,omitempty"`
	DataReplicaSize int        `json:"dataReplicaSize,omitempty"`
	ShardCount      int        `json:"shardCount,omitempty"`
	ThisDomainName  string     `json:"thisDomainName,omitempty"`
	OtherDomainName string     `json:"otherDomainName,omitempty"`
}

type MongoDB struct {
	Install             bool   `json:"install,omitempty"`
	MongoRootSecretName string `json:"mongoRootSecretName,omitempty"`
}

// DisasterRecovery shows Disaster Recovery configuration
type DisasterRecovery struct {
	Mode   string `json:"mode"`
	Status string `json:"status,omitempty"`
	NoWait bool   `json:"noWait,omitempty"`
}

// MongodbSupplServiceSpec defines the desired state of MongodbSupplService
type MongodbSupplServiceSpec struct {
	MongoDB                   `json:"mongodb"`
	Backup                    `json:"backup"`
	Dbaas                     `json:"dbaas"`
	PrometheusExporter        `json:"prometheusExporter"`
	RobotTests                `json:"robotTests"`
	TLS                       types.TLS               `json:"tls,omitempty"`
	DeploymentVersion         string                  `json:"deploymentVersion,omitempty"`
	Recycler                  types.Recycler          `json:"recycler,omitempty"`
	WaitSeconds               int                     `json:"waitSeconds,omitempty"`
	PodSecurityContext        *v1.PodSecurityContext  `json:"podSecurityContext,omitempty"`
	ServiceAccountName        string                  `json:"serviceAccountName"`
	ImagePullPolicy           v1.PullPolicy           `json:"imagePullPolicy,omitempty" common:"true"`
	Policies                  *Policies               `json:"policies,omitempty" common:"true"`
	VaultRegistration         types.VaultRegistration `json:"vaultRegistration" common:"true"`
	VaultDBEngine             types.VaultDBEngine     `json:"vaultDBEngine"  common:"true"`
	SchemaSettings            SchemaSettings          `json:"schemaSettings,omitempty" common:"true"`
	DisasterRecovery          *DisasterRecovery       `json:"disasterRecovery,omitempty"`
	DeletePVConUninstall      bool                    `json:"deletePVConUninstall,omitempty"`
	AuthDb                    string                  `json:"authDb,omitempty"`
	IpV6                      bool                    `json:"ipV6,omitempty"`
	ArtifactDescriptorVersion string                  `json:"artifactDescriptorVersion,omitempty"`
	PartOf                    string                  `json:"partOf,omitempty"`
	ManagedBy                 string                  `json:"managedBy,omitempty"`
	Instance                  string                  `json:"instance,omitempty"`
	CloudPublicHost           string                  `json:"cloudPublicHost,omitempty"`
}

// MongodbSupplServiceStatus defines the observed state of MongodbSupplService
type MongodbSupplServiceStatus struct {
	Conditions []types.ServiceStatusCondition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MongodbSupplService is the Schema for the mongodbsupplservices API
type MongodbSupplService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MongodbSupplServiceSpec   `json:"spec,omitempty"`
	Status MongodbSupplServiceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MongodbSupplServiceList contains a list of MongodbSupplService
type MongodbSupplServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MongodbSupplService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MongodbSupplService{}, &MongodbSupplServiceList{})
}

type BackupDaemonTLS struct {
	// a name of Kubernetes secret that holds a CA certificate, a Signed MongoDB Backup Daemon certificate and a private key.
	BackupDaemonCASecretName string `json:"backupDaemonCASecretName,omitempty"`
}

type DbaasAdapterTLS struct {
	// a name of Kubernetes secret that holds a CA certificate, a Signed MongoDB Dbaas Adapter certificate and a private key.
	DbaasAdapterCASecretName string `json:"dbaasAdapterCASecretName,omitempty"`
}

type Backup struct {
	Install                        bool                       `json:"install,omitempty"`
	StorageDirectory               string                     `json:"storageDirectory,omitempty"`
	DockerImage                    string                     `json:"dockerImage,omitempty"`
	BackupSchedule                 string                     `json:"backupSchedule,omitempty"`
	IncrementalBackupSchedule      string                     `json:"incBackupSchedule,omitempty"`
	EvictionPolicy                 string                     `json:"evictionPolicy,omitempty"`
	IncrementalEvictionPolicy      string                     `json:"incEvicitionPolicy,omitempty"`
	MongoDatabasePrefixPattern     string                     `json:"mongoDatabasePrefixPattern,omitempty"`
	Storage                        *types.StorageRequirements `json:"storage,omitempty"`
	Resources                      *v1.ResourceRequirements   `json:"backupResources,omitempty"`
	NumParallelConnections         int                        `json:"numParallelConnections,omitempty"`
	GranularNumParallelConnections int                        `json:"granularNumParallelConnections,omitempty"`
	MongoBackupDB                  string                     `json:"mongoBackupDB,omitempty"`
	MongoSourceDB                  string                     `json:"mongoSourceDB,omitempty"`
	EnableFullRestore              bool                       `json:"enableFullRestore,omitempty"`
	ConfigCollections              []string                   `json:"configCollections,omitempty"`
	AdditionalNodeLabels           map[string]string          `json:"additionalNodeLabels,omitempty"`
	S3                             S3backup                   `json:"s3,omitempty"`
	TLS                            BackupDaemonTLS            `json:"tls,omitempty"`
	GranularBackupSchedule         string                     `json:"granularBackupSchedule,omitempty"`
	GranularBackupScheduledDbs     []string                   `json:"granularBackupScheduledDbs,omitempty"`
	PriorityClassName              string                     `json:"priorityClassName,omitempty"`
	BackupApiSecretName            string                     `json:"backupApiSecretName,omitempty"`
	RestoreUserSecretName          string                     `json:"restoreUserSecretName,omitempty"`
	BackupSecretName               string                     `json:"backupSecretName,omitempty"`
	Affinity                       *v1.Affinity               `json:"affinity,omitempty"`
}

type S3backup struct {
	Enabled         bool   `json:"enabled,omitempty"`
	SecretName      string `json:"secretName,omitempty"`
	BucketName      string `json:"bucketName,omitempty"`
	AccessKeyId     string `json:"accessKeyId,omitempty"`
	AccessKeySecret string `json:"accessKeySecret,omitempty"`
	EndpointUrl     string `json:"endpointUrl,omitempty"`
	SslVerify       bool   `json:"sslVerify,omitempty"`
	SslSecretName   string `json:"sslSecretName,omitempty"`
	SslCert         string `json:"sslCert,omitempty"`
}

type Dbaas struct {
	Install                                   bool                     `json:"install,omitempty"`
	DockerImage                               string                   `json:"dockerImage,omitempty"`
	AdditionalNodeLabels                      map[string]string        `json:"additionalNodeLabels,omitempty"`
	Resources                                 *v1.ResourceRequirements `json:"dbaasResources,omitempty"`
	DbaasAggregatorRegistrationAddress        string                   `json:"dbaasAggregatorRegistrationAddress,omitempty"`
	DbaasPhysicalDatabasesCustomLabels        map[string]string        `json:"dbaasPhysicalDatabasesCustomLabels,omitempty"`
	DbaasAggregatorPhysicalDatabaseIdentifier string                   `json:"dbaasAggregatorPhysicalDatabaseIdentifier,omitempty"`
	DbaasAggregatorRegistrationFixedDelayMS   int                      `json:"dbaasAggregatorRegistrationFixedDelayMS,omitempty"`
	DbaasAggregatorRegistrationRetryDelayMS   int                      `json:"dbaasAggregatorRegistrationRetryDelayMS,omitempty"`
	DbaasAggregatorRegistrationRetryTimeMS    int                      `json:"dbaasAggregatorRegistrationRetryTimeMS,omitempty"`
	ApiVersion                                string                   `json:"apiVersion,omitempty"`
	MultiUsers                                bool                     `json:"multiUsers,omitempty"`
	TLS                                       DbaasAdapterTLS          `json:"tls,omitempty"`
	PriorityClassName                         string                   `json:"priorityClassName,omitempty"`
	DbaasAdminSecretName                      string                   `json:"dbaasAdminSecretName,omitempty"`
	DbaasAggregatorSecretName                 string                   `json:"dbaasAggregatorSecretName,omitempty"`
	DbaasRegistrationSecretName               string                   `json:"dbaasRegistrationSecretName,omitempty"`
	DbaasAdminRoleSecretName                  string                   `json:"dbaasAdminRoleSecretName,omitempty"`
	Affinity                                  *v1.Affinity             `json:"affinity,omitempty"`
}

type PrometheusExporter struct {
	Install                      bool                     `json:"install,omitempty"`
	DockerImage                  string                   `json:"dockerImage,omitempty"`
	AdditionalNodeLabels         map[string]string        `json:"additionalNodeLabels,omitempty"`
	Resources                    *v1.ResourceRequirements `json:"exporterResources,omitempty"`
	MongoConnectionTimeout       int                      `json:"mongoConnectionTimeout,omitempty"`
	PriorityClassName            string                   `json:"priorityClassName,omitempty"`
	MonitoringSecretName         string                   `json:"monitoringSecretName,omitempty"`
	PrometheusExporterSecretName string                   `json:"prometheusExporterSecretName,omitempty"`
}

type RobotTests struct {
	Install            bool                     `json:"install,omitempty"`
	DockerImage        string                   `json:"dockerImage,omitempty"`
	Resources          *v1.ResourceRequirements `json:"resources,omitempty"`
	Tags               string                   `json:"tags,omitempty"`
	NodeLabels         map[string]string        `json:"nodeLabels,omitempty"`
	ExternalBackupPath string                   `json:"externalBackupPath,omitempty"`
	MainSide           string                   `json:"mainSide,omitempty"`
	LeftNodesPattern   string                   `json:"leftNodesPattern,omitempty"`
	RightNodesPattern  string                   `json:"rightNodesPattern,omitempty"`
	PriorityClassName  string                   `json:"priorityClassName,omitempty"`
	Affinity           *v1.Affinity             `json:"affinity,omitempty"`
}
