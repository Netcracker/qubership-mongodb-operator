package utils

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"

	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	coreUtils "github.com/Netcracker/qubership-nosqldb-operator-core/pkg/utils"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func __max(weight_list []int, poped_indexes []int, shard_size int) int {
	max_index := -1
	max_weight := -1
	if shard_size != 1 {
		for x := 0; x < len(weight_list); x++ {
			if core.ContainsInt(poped_indexes, x) == false && weight_list[x] > max_weight {
				max_weight = weight_list[x]
				max_index = x
			}
		}
	} else {
		for x := len(weight_list) - 1; x > -1; x-- {
			if core.ContainsInt(poped_indexes, x) == false && weight_list[x] > max_weight {
				max_weight = weight_list[x]
				max_index = x
			}
		}
	}
	return max_index
}

func intArrWithValues(size int, v int) []int {
	a := make([]int, size)
	for i := range a {
		a[i] = v
	}
	return a
}

func GetDummyPvcMap(pvcNames []string) []map[string]string {
	pvcNamesDummyMap := make([]map[string]string, len(pvcNames))
	for i, pvc := range pvcNames {
		pvcNamesDummyMap[i] = map[string]string{KubeHostName: pvc}
	}

	return pvcNamesDummyMap
}

func ElementsDistribution(elements []map[string]string, replicasCount int, replicaNum int, shardsCount int, shardNum int) (string, string) {
	pv_nodes := elements
	pv_length := len(pv_nodes)
	repl_size := replicasCount
	shard_size := shardsCount

	pv_length_side := (pv_length - 1) / 2
	repl_size_side := (repl_size - 1) / 2

	weights := intArrWithValues(pv_length_side, 0)

	extra_nodes := pv_length_side - repl_size_side

	result := [][]string{}
	for shard := 0; shard < shard_size; shard++ {
		shard_nodes := make([]string, len(pv_nodes))
		for node := 0; node < len(pv_nodes); node++ {
			shard_nodes[node] = getMapValues(pv_nodes[node])[0]
		}

		poped_indexes := intArrWithValues(extra_nodes, -1)

		for node := 0; node < extra_nodes; node++ {
			pop_index := __max(weights, poped_indexes, shard_size)
			poped_indexes[node] = pop_index

			shard_nodes[pop_index] = ""
			shard_nodes[len(shard_nodes)-1-pop_index] = ""
		}

		for i := 0; i < pv_length_side; i++ {
			if core.ContainsInt(poped_indexes, i) == false {
				weights[i] += 1
			}
		}

		shard_nodes_result := []string{}
		for x := 0; x < len(shard_nodes); x++ {
			if shard_nodes[x] != "" {
				shard_nodes_result = append(shard_nodes_result, shard_nodes[x])
			}
		}

		result = append(result, shard_nodes_result)
	}
	return getMapKeys(pv_nodes[replicaNum])[0], result[shardNum][replicaNum]
}

func getMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}

func getMapValues(m map[string]string) []string {
	values := make([]string, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}

	return values
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func MongoReplicaNodeSelector(elements []map[string]string, replicasCount int, replicaNum int, shardsCount int, shardNum int) map[string]string {

	if len(elements) < 1 {
		return map[string]string{}
	} else {
		labelKey, labelValue := ElementsDistribution(elements, replicasCount, replicaNum, shardsCount, shardNum)
		return map[string]string{
			labelKey: labelValue,
		}
	}
}

func MongoReplicaContainerArgs(replicaType string, data string, nameKey string, nameWithIndexes string, mongoSecret string,
	mongoSecretKeyFile string, wiredCacheGb float64, ipv6 bool, customDataRSParams []string, tls *v1alpha1.TLS, OpLogSizeMb int64, keyfileAuth bool) string {

	authOpt1 := "--auth"
	authOpt2 := "--keyFile"
	authOpt3 := "'/" + data + "/" + nameWithIndexes + "/" + mongoSecretKeyFile + "'"
	if !keyfileAuth {
		authOpt2 = "--clusterAuthMode"
		authOpt3 = X509AuthMode
	}

	bindOpt := "--bind_ip_all"

	if ipv6 {
		bindOpt = "--bind_ip_all --ipv6"
	}

	var setParameters string
	if len(customDataRSParams) > 0 {
		for _, v := range customDataRSParams {
			setParameters += fmt.Sprintf("--setParameter %s ", v)
		}
	}

	cmd := "if  [ ! -d /" + data + "/" + nameWithIndexes + " ]; then  mkdir /" + data + "/" + nameWithIndexes + "; fi && cp /opt/" + mongoSecret + "/" + mongoSecretKeyFile + " /" + data + "/" + nameWithIndexes + "/" + mongoSecretKeyFile + " && chmod 0600 /" + data + "/" + nameWithIndexes + "/" + mongoSecretKeyFile + " && mongod " + bindOpt + " --port 27017 --dbpath  /" + data + "/" + nameWithIndexes + " --replSet " + nameKey + " " + replicaType + " --wiredTigerCacheSizeGB " + fmt.Sprintf("%f", wiredCacheGb) + " " + authOpt1 + " " + authOpt2 + " " + authOpt3 + " " + setParameters + getTLSOPtions(tls)
	if OpLogSizeMb > 0 {
		cmd += " --oplogSize " + fmt.Sprintf("%d", OpLogSizeMb)
	}

	return cmd

}

func HAMongosContainerArgs(tmp string, data string, mongoSecret string, mongoSecretKeyFile string, configNodes string, ipv6 bool, tls *v1alpha1.TLS, keyfileAuth bool) string {
	authOpt2 := "--keyFile"
	authOpt3 := "'/" + tmp + "/" + data + "/" + mongoSecretKeyFile + "'"
	if !keyfileAuth {
		authOpt2 = "--clusterAuthMode"
		authOpt3 = X509AuthMode
	}

	bindOpt := "--bind_ip_all"

	if ipv6 {
		bindOpt = "--bind_ip_all --ipv6"
	}

	return "if  [ ! -d /" + tmp + "/" + data + " ]; then  mkdir /" + tmp + "/" + data + "; fi && cp /opt/" + mongoSecret + "/" + mongoSecretKeyFile + " /" + tmp + "/" + data + "/" + mongoSecretKeyFile + " && chmod 0600 /" + tmp + "/" + data + "/" + mongoSecretKeyFile + " && mongos " + bindOpt + " --port 27017 " + authOpt2 + " " + authOpt3 + " --configdb " + configNodes + getTLSOPtions(tls)
}

func SingleMongosContainerArgs(wiredCacheGb float64, ipv6 bool, tls *v1alpha1.TLS) string {

	v6Opt := ""

	if ipv6 {
		v6Opt = "--ipv6"
	}

	return "mongod --bind_ip_all " + v6Opt + " --port 27017 --dbpath /data --auth --wiredTigerCacheSizeGB " + fmt.Sprintf("%f", wiredCacheGb) + getTLSOPtions(tls)
}

func getTLSOPtions(tls *v1alpha1.TLS) string {
	var tlsOptions string
	if tls.Enabled {
		tlsOptions = fmt.Sprintf(" --tlsMode %s --tlsCertificateKeyFile %s%s --tlsCAFile %s%s --tlsAllowConnectionsWithoutCertificates", tls.Mode, RootCertPath, tls.CombinedKeyAndCRTFileName, RootCertPath, "ca.crt")
	}
	return tlsOptions
}

func MongosRegisterNodesString(nameKey string, replicaSize int, namespace string) string {
	configNodes := nameKey + "/"
	domain := fmt.Sprintf(MongoDomainTemplate, namespace)
	splitter := ""
	for i := 0; i < replicaSize; i++ {
		configNodes += fmt.Sprintf("%s%s.%s.%s:27017", splitter, fmt.Sprintf(StatefulSetPodNameTemplate, fmt.Sprintf("%s%v", nameKey, i)), nameKey, domain)
		splitter = ","
	}

	return configNodes
}

func HostNameBuilder(name string, index int, port bool) string {
	portStr := ""
	if port {
		portStr = ":27017"
	}
	hostName := fmt.Sprintf("%s%v-0.%s%s", name, index, name, portStr)

	return hostName
}

// TODO do we need this?
func HostsList(name string, replicaSize int, port bool) []string {
	members := []string{}

	for i := 0; i < replicaSize; i++ {
		hostName := HostNameBuilder(name, i, port)
		members = append(members, hostName)
	}

	return members
}

func CalcProperMongoWiredCacheSize(specMemLimit int64, specCacheGb string, maxDefault float64) float64 {
	var specCacheGbf float64
	if specCacheGb != "" {
		var err error
		specCacheGbf, err = strconv.ParseFloat(specCacheGb, 64)
		if err != nil {
			panic(err)
		}
	}
	var wiredCacheGb float64
	if specCacheGbf == 0 {
		valGb := float64(specMemLimit / 1024 / 1024 / 1024)
		half := valGb / 2
		wiredCacheGb = float64(half - 1)
	} else {
		wiredCacheGb = specCacheGbf
	}

	return math.Max(wiredCacheGb, maxDefault)
}

type OutputResult struct {
	Ok     int    `json:"ok,omitempty"`
	Code   int    `json:"code,omitempty"`
	ErrMsg string `json:"errmsg,omitempty"`
}

func CreateUserForMongoReplicaCommand(authDb string, user string, pass string, role string) string {
	return fmt.Sprintf(
		"db.getSiblingDB('%s').createUser({user: '%s', pwd: '%s', roles: [%s]});",
		authDb, user, pass, role)
}

func CreateRoleForMongoReplicaCommand(authDb string, role string, privileges string, roles string) string {
	return fmt.Sprintf(
		"db.getSiblingDB('%s').createRole({ role: '%s', privileges: [ %s ], roles: [ %s ]});",
		authDb, role, privileges, roles)
}

// TODO vault
func GrantRoleToUser(ctx core.ExecutionContext, cmd, user, role string) error {
	mongoImpl := ctx.Get(MongoHelperImpl).(MongoHelper)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)

	_, err := mongoImpl.RunOnMongos(
		fmt.Sprintf("db.grantRolesToUser('%s', ['%s'])", user, role),
	)

	if err != nil {
		log.Warn(fmt.Sprintf("A problem during granting %s role to %s user check: %+v", role, user, err))
		return err
	}

	return nil
}

type UserToAdd struct {
	User       string
	Pass       func() string
	Role       string
	ShardLocal bool
	AddToVault bool
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

func CreateRuntimeObjectContextWrapper(ctx core.ExecutionContext, object client.Object, meta v12.ObjectMeta) error {
	scheme := ctx.Get(constants.ContextSchema).(*runtime.Scheme)
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	helper := ctx.Get(KubernetesHelperImpl).(core.KubernetesHelper)
	specPointer := &(*spec)

	return helper.CreateRuntimeObject(scheme, specPointer, object, meta)
}

func ReadSecret(ctx core.ExecutionContext, name string, namespace string) (*v1.Secret, error) {
	kubeClient := ctx.Get(constants.ContextClient).(client.Client)
	secret := &v1.Secret{}
	err := kubeClient.Get(context.TODO(),
		types.NamespacedName{Name: name, Namespace: namespace}, secret)

	return secret, err
}

func IsAnyCommonParamChanged(ctx core.ExecutionContext, specsToCheckMap map[string]interface{}) (bool, error) {

	resultCheck := false

	for specKey, specToCheck := range specsToCheckMap {
		specHasChanges, checkErr := core.CheckSpecChange(ctx, specToCheck, specKey)

		if checkErr != nil {
			return false, checkErr
		}

		if specHasChanges == true {
			resultCheck = true
		}
	}

	return resultCheck, nil
}

func HalfIndexFunc(repSize int) string {
	half := repSize / 2
	primaryIndexes := ""
	splitter := ""
	for i := half; i < repSize; i++ {
		primaryIndexes += splitter + fmt.Sprintf("%v", i)
		splitter = ","
	}
	primaryInit := fmt.Sprintf("[%s].indexOf(member._id) >= 0", primaryIndexes)

	return primaryInit
}

func RemoveElementFromSlice(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
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

func GetDRCNFReplicaSetHostNames(replicaSize int, mainDomain string, drDomain string, namespace string) []string {
	mainReplicas := GetCNFReplicaSetHostNames(replicaSize, mainDomain, namespace)
	drReplicas := GetCNFReplicaSetHostNames(replicaSize, drDomain, namespace)
	return append(mainReplicas, drReplicas...)
}

func GetDRDATAReplicaSetHostNames(replicaSize int, shardIndex int, mainDomain string, drDomain string, namespace string) []string {
	mainReplicas := GetDATAReplicaSetHostName(replicaSize, shardIndex, mainDomain, namespace)
	drReplicas := GetDATAReplicaSetHostName(replicaSize, shardIndex, drDomain, namespace)
	return append(mainReplicas, drReplicas...)
}

func QuoteReplicaSet(replicasetHostnames []string) []string {
	for i, name := range replicasetHostnames {
		replicasetHostnames[i] = fmt.Sprintf("'%s'", name)
	}

	return replicasetHostnames
}

func TLSClientSpecUpdate(podSpec *v1.PodSpec, tls v1alpha1.TLS) {
	if !tls.Enabled {
		return
	}

	podSpec.Volumes = append(podSpec.Volumes,
		v1.Volume{
			Name: tls.CertificateSecretName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: tls.CertificateSecretName,
				},
			},
		},
	)

	podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts,
		v1.VolumeMount{
			Name:      tls.CertificateSecretName,
			ReadOnly:  true,
			MountPath: RootCertPath,
		},
	)

	if len(podSpec.Containers[0].ReadinessProbe.ProbeHandler.Exec.Command) > 0 {
		podSpec.Containers[0].ReadinessProbe.ProbeHandler.Exec.Command = append(podSpec.Containers[0].ReadinessProbe.ProbeHandler.Exec.Command, "--tls", "--tlsCAFile", RootCertPath+tls.RootCAFileName)
	}
}

func TLSServerSpecUpdate(depl *v1.PodSpec, tls v1alpha1.TLS, secretName string, mountPath string) {
	if !tls.Enabled {
		return
	}

	depl.Volumes = append(depl.Volumes,
		v1.Volume{
			Name: secretName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: secretName,
				},
			},
		},
	)

	depl.Containers[0].VolumeMounts = append(depl.Containers[0].VolumeMounts,
		v1.VolumeMount{
			Name:      secretName,
			ReadOnly:  true,
			MountPath: mountPath,
		},
	)

	depl.Containers[0].Env = append(depl.Containers[0].Env,
		coreUtils.GetPlainTextEnvVar("INTERNAL_TLS_ENABLED", strconv.FormatBool(tls.Enabled)),
		coreUtils.GetPlainTextEnvVar("INTERNAL_TLS_ROOTCERT", mountPath+tls.RootCAFileName),
		coreUtils.GetPlainTextEnvVar("INTERNAL_TLS_CERTIFICATE_FILENAME", mountPath+tls.SignedCRTFileName),
		coreUtils.GetPlainTextEnvVar("INTERNAL_TLS_KEY_FILENAME", mountPath+tls.PrivateKeyFileName),
		coreUtils.GetPlainTextEnvVar("INTERNAL_TLS_PATH", mountPath),
	)
}

func GetHTTPPort(tlsEnabled bool) int32 {
	var port int32 = 8080
	if tlsEnabled {
		port = 8443
	}
	return port
}

func GetHTTPProtocol(tlsEnabled bool) string {
	if tlsEnabled {
		return "https"
	}
	return "http"
}

func IsTLSEnableForDBAAS(aggregatorRegistrationAddress string, tlsEnabled bool) bool {
	return strings.Contains(aggregatorRegistrationAddress, "https") && tlsEnabled
}
