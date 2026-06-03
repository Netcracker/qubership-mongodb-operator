package utils

import (
	"fmt"
	"strings"

	"github.com/Netcracker/qubership-mongodb-supplementary/api/v1alpha1"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateRuntimeObjectContextWrapper(ctx core.ExecutionContext, object client.Object, meta v1.ObjectMeta) error {
	scheme := ctx.Get(constants.ContextSchema).(*runtime.Scheme)
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbSupplService)
	helper := ctx.Get(KubernetesHelperImpl).(core.KubernetesHelper)
	specPointer := &(*spec)

	return helper.CreateRuntimeObject(scheme, specPointer, object, meta)
}

type UserToAdd struct {
	User       string
	Pass       func() string
	Role       string
	ShardLocal bool
}

func AddServicesUsersToContext(ctx core.ExecutionContext, usr UserToAdd) {
	iUsers := ctx.Get(ServicesUsersContextList)
	users := map[string]UserToAdd{}
	if iUsers != nil {
		users = iUsers.(map[string]UserToAdd)
	}
	users[usr.User] = usr
	ctx.Set(ServicesUsersContextList, users)
}

// ShardLocal flag determines if this role should be created on shards
type RoleToAdd struct {
	Role       string
	Privileges string
	Roles      string
	ShardLocal bool
}

func AddServicesRolesToContext(ctx core.ExecutionContext, role RoleToAdd) {
	iRoles := ctx.Get(ServicesRolesContextList)
	roles := map[string]RoleToAdd{}
	if iRoles != nil {
		roles = iRoles.(map[string]RoleToAdd)
	}
	roles[role.Role] = role
	ctx.Set(ServicesRolesContextList, roles)
}

func GetReplicaSetHostNames(replicaSize int, fullDomain string) []string {
	rs := make([]string, replicaSize)
	for i := 0; i < replicaSize; i++ {
		rs[i] = fmt.Sprintf(fullDomain, i)
	}

	return rs
}

func GetCNFReplicaSetHostNames(replicaSize int, domain string, namespace string) []string {
	fullDomain := fmt.Sprintf("%s-0.%s.%s.svc.%s:27017", CnfNameWithIndexFormat, CnfNameKey, namespace, domain)
	return GetReplicaSetHostNames(replicaSize, fullDomain)
}

func GetDATAReplicaSetHostName(replicaSize int, shardIndex int, domain string, namespace string) []string {
	serviceName := fmt.Sprintf(DataNameKey, shardIndex+1)
	podNamePattern := fmt.Sprintf(DataNameWithIndexesFormat, shardIndex+1, "%v-0")
	fullDomain := fmt.Sprintf("%s.%s.%s:27017", podNamePattern, serviceName, fmt.Sprintf("%s.svc.%s", namespace, domain))

	return GetReplicaSetHostNames(replicaSize, fullDomain)
}

func GetUriScheme(tlsEnabled bool) v12.URIScheme {
	if tlsEnabled {
		return v12.URISchemeHTTPS
	}
	return v12.URISchemeHTTP
}

func SanitizeVolumeName(name string) string {
	return strings.ReplaceAll(name, ".", "-")
}
