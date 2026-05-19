/*
Copyright 2022.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MongodbDeploymentSpec defines the desired state of MongodbDeployment
type MongodbDeploymentSpec struct {
	DisasterRecovery           *DisasterRecovery       `json:"disasterRecovery,omitempty"`
	DeploymentVersion          string                  `json:"deploymentVersion,omitempty"`
	SchemaSettings             SchemaSettings          `json:"schemaSettings,omitempty" common:"true"`
	Recycler                   Recycler                `json:"recycler,omitempty"`
	StopOnFailedResourceUpdate bool                    `json:"stopOnFailedResourceUpdate,omitempty"`
	WaitSeconds                int                     `json:"waitSeconds,omitempty"`
	AuthDb                     string                  `json:"authDb,omitempty"`
	IpV6                       bool                    `json:"ipV6,omitempty"`
	PodSecurityContext         *v1.PodSecurityContext  `json:"podSecurityContext,omitempty"`
	Policies                   *Policies               `json:"policies,omitempty" common:"true"`
	VaultRegistration          types.VaultRegistration `json:"vaultRegistration,omitempty" common:"true"`
	VaultDBEngine              types.VaultDBEngine     `json:"vaultDBEngine"  common:"true"`
	ServiceAccountName         string                  `json:"serviceAccountName"`
	ImagePullPolicy            v1.PullPolicy           `json:"imagePullPolicy,omitempty" common:"true"`
	MongoDB                    `json:"mongodb"`
	TLS                        `json:"tls,omitempty" common:"true"`
	CloudPublicHost            string `json:"cloudPublicHost,omitempty"`
	DeletePVConUninstall       bool   `json:"deletePVConUninstall,omitempty"`
}

// MongodbDeploymentStatus defines the observed state of MongodbDeployment
type MongodbDeploymentStatus struct {
	DisasterRecoveryStatus types.DisasterRecoveryStatus   `json:"disasterRecoveryStatus,omitempty"`
	Conditions             []types.ServiceStatusCondition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MongodbDeployment is the Schema for the mongodbdeployments API
type MongodbDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MongodbDeploymentSpec   `json:"spec,omitempty"`
	Status MongodbDeploymentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MongodbDeploymentList contains a list of MongodbDeployment
type MongodbDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MongodbDeployment `json:"items"`
}

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
	//schemaType: "ha", "single", "dr", "arbiter"
	SchemaType      SchemaType `json:"schemaType,omitempty"`
	Sharded         bool       `json:"sharded,omitempty"`
	CnfReplicaSize  int        `json:"cnfReplicaSize,omitempty"`
	DataReplicaSize int        `json:"dataReplicaSize,omitempty"`
	ShardCount      int        `json:"shardCount,omitempty"`
	ArbiterIndex    int        `json:"arbiterIndex,omitempty"`
	ThisDomainName  string     `json:"thisDomainName,omitempty"`
	OtherDomainName string     `json:"otherDomainName,omitempty"`
	// Parameter defined mongos replicas to deploy.
	// If mutiple pods write and read simultaneously the same database they can access not latest data.
	// Session affinity parameter on Mongos service has timeout of 3 hours, which means that application can be redirected to another mongos if this timeout is reached.
	MongosReplicas int `json:"mongosReplicas,omitempty"`
}

// DisasterRecovery shows Disaster Recovery configuration
type DisasterRecovery struct {
	Mode   string `json:"mode"`
	Status string `json:"status,omitempty"`
	NoWait bool   `json:"noWait,omitempty"`
}

type Recycler struct {
	Install   bool                     `json:"install,omitempty"`
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
}

type TLS struct {
	types.TLS `json:",omitempty"`
	// Enables TLS used for all network connections. Possible values: disabled / allowTLS /  preferTLS / requireTLS
	Mode string `json:"mode,omitempty"`
	//a key in the Kubernetes secret `tls.combinedKeyAndCRTFileName` that holds the Signed certificate + Private key.
	CombinedKeyAndCRTFileName string `json:"combinedKeyAndCRTFileName,omitempty"`
}

type MongoDB struct {
	Install                 bool                       `json:"install,omitempty"`
	CnfResources            *v1.ResourceRequirements   `json:"cnfResources,omitempty"`
	DataResources           *v1.ResourceRequirements   `json:"dataResources,omitempty"`
	ArbiterResources        *v1.ResourceRequirements   `json:"arbiterResources,omitempty"`
	MongosResources         *v1.ResourceRequirements   `json:"mongosResources,omitempty"`
	DockerImage             string                     `json:"dockerImage,omitempty"`
	ContainerTimeoutSeconds int                        `json:"containerTimeoutSeconds,omitempty"`
	ContainerPeriodSeconds  int                        `json:"containerPeriodSeconds,omitempty"`
	Storage                 *types.StorageRequirements `json:"storage,omitempty"`
	CnfWiredTigerCacheGb    string                     `json:"cnfWiredTigerCacheGb,omitempty"`
	DataWiredTigerCacheGb   string                     `json:"dataWiredTigerCacheGb,omitempty"`
	CnfOpLogSizeMb          int64                      `json:"cnfOpLogSizeMb,omitempty"`
	DataOpLogSizeMb         int64                      `json:"dataOpLogSizeMb,omitempty"`
	SingleWiredTigerCacheGb string                     `json:"singleWiredTigerCacheGb,omitempty"`
	AdditionalNodeLabels    map[string]string          `json:"additionalNodeLabels,omitempty"`
	CustomDataRSParameters  []string                   `json:"customDataRSParameters,omitempty"`
	PriorityClassName       string                     `json:"priorityClassName,omitempty"`
	MongoRootSecretName     string                     `json:"mongoRootSecretName,omitempty"`
	Affinity                *v1.Affinity               `json:"affinity,omitempty"`
	ClusterAuthMode         string                     `json:"clusterAuthMode,omitempty"`
}

func init() {
	SchemeBuilder.Register(&MongodbDeployment{}, &MongodbDeploymentList{})
}
