package prometheus_exporter

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/Netcracker/qubership-mongodb-supplementary/api/v1alpha1"
	cUtils "github.com/Netcracker/qubership-nosqldb-operator-core/pkg/utils"

	"github.com/Netcracker/qubership-mongodb-supplementary/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"go.uber.org/zap"
	v12 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type PrometheusExporterDeployment struct {
	core.DefaultExecutable
}

func (r *PrometheusExporterDeployment) Validate(ctx core.ExecutionContext) error {
	return nil
}

type PrometheusExporterURI struct {
	ShardMembers []string `json:"shardMembers,omitempty"`
	CnfrsMembers string   `json:"cnfrsMembers,omitempty"`
}

func (r *PrometheusExporterDeployment) Execute(ctx core.ExecutionContext) error {
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbSupplService)
	exporter := spec.Spec.PrometheusExporter
	helperImpl := ctx.Get(utils.KubernetesHelperImpl).(core.KubernetesHelper)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	credsManager := ctx.Get(utils.ContextCredsManager).(utils.CredsManagerI)

	log.Info("PrometheusExporter Deployment initialization step started")

	envs := []v12.EnvVar{
		v12.EnvVar{
			Name: "NAMESPACE",
			ValueFrom: &v12.EnvVarSource{
				FieldRef: &v12.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
	}

	shards := []string{}
	members := []string{}

	if spec.Spec.SchemaSettings.SchemaType != v1alpha1.Single {
		// arbFunc := ctx.Get(utils.ArbiterIndexSelectorFunc).(func(int) int)
		// arbiterNodeIndex := arbFunc(ctx.Get(utils.MaxReplicaSize).(int))

		cnfCount := spec.Spec.SchemaSettings.CnfReplicaSize
		members = utils.GetCNFReplicaSetHostNames(cnfCount, spec.Spec.SchemaSettings.ThisDomainName, request.Namespace)

		shardCount := spec.Spec.SchemaSettings.ShardCount
		dataCount := spec.Spec.SchemaSettings.DataReplicaSize

		for s := 0; s < shardCount; s++ {
			rs := utils.GetDATAReplicaSetHostName(dataCount, s, spec.Spec.SchemaSettings.ThisDomainName, request.Namespace)
			// if arbiterNodeIndex >= 0 {
			// 	rs = core.RemoveElementFromSlice(rs, arbiterNodeIndex)
			// }
			shards = append(shards, strings.Join(rs[:], ","))
		}
	}
	cnfrsMemberList := ""
	if !spec.Spec.SchemaSettings.Sharded {
		cnfrsMemberList = ""
	} else {
		cnfrsMemberList = strings.Join(members[:], ",")
	}
	uri := PrometheusExporterURI{
		ShardMembers: shards,
		CnfrsMembers: cnfrsMemberList,
	}

	jsonBytes, jsonErr := json.Marshal(uri)

	if jsonErr != nil {
		return &core.ExecutionError{Msg: "Error happened on URI calculation. Error: " + jsonErr.Error()}
	}

	simpleKV := map[string]string{
		"MONGO_URI": string(jsonBytes),
	}

	if !spec.Spec.SchemaSettings.Sharded {
		simpleKV["EXPORT_MONGOS"] = strconv.FormatBool(false)
	} else if spec.Spec.SchemaSettings.SchemaType == v1alpha1.DR {
		simpleKV["EXPORT_MONGOS"] = strconv.FormatBool(spec.Spec.DisasterRecovery.Mode == utils.ActiveMode)
	}

	for key, value := range simpleKV {
		envs = append(envs,
			v12.EnvVar{
				Name:  key,
				Value: value,
			})
	}

	secretVolumes := map[string]string{
		spec.Spec.PrometheusExporter.MonitoringSecretName:         "/var/run/secrets/mongodb/mongo-monitoring",
		spec.Spec.PrometheusExporter.PrometheusExporterSecretName: "/var/run/secrets/mongodb/prom-exporter",
	}

	volumes := []v12.Volume{}
	volumeMounts := []v12.VolumeMount{}

	for secretName, mountPath := range secretVolumes {

		volumeName := utils.SanitizeVolumeName(secretName)

		volumes = append(volumes, v12.Volume{
			Name: volumeName,
			VolumeSource: v12.VolumeSource{
				Secret: &v12.SecretVolumeSource{
					SecretName: secretName,
				},
			},
		})

		volumeMounts = append(volumeMounts, v12.VolumeMount{
			Name:      volumeName,
			MountPath: mountPath,
			ReadOnly:  true,
		})
	}

	// Environment variable End
	var tolerations []v12.Toleration
	if spec.Spec.Policies != nil {
		tolerations = spec.Spec.Policies.Tolerations
	}

	var healthzPort int
	if !spec.Spec.SchemaSettings.Sharded {
		healthzPort = 9218
	} else if simpleKV["EXPORT_MONGOS"] == "false" && len(members) > 0 {
		healthzPort = 9217
	} else {
		healthzPort = 9216
	}

	dc := PrometheusExporterDeploymentTemplate(
		request.Namespace,
		exporter.DockerImage,
		exporter.AdditionalNodeLabels,
		spec.Spec.ServiceAccountName,
		*exporter.Resources,
		envs,
		spec.Spec.PodSecurityContext,
		tolerations,
		spec.Spec.ImagePullPolicy,
		spec.Spec.TLS,
		spec.Spec.PrometheusExporter.PriorityClassName,
		spec.Spec.Instance,
		healthzPort,
		volumeMounts,
		volumes)

	err := credsManager.AddCredHashToPodTemplate([]string{spec.Spec.MongoDB.MongoRootSecretName}, &dc.Spec.Template)
	if err != nil {
		log.Error(fmt.Sprintf("can't add secret HASH to annotations for %s", dc.Name), zap.Error(err))
		return err
	}

	cUtils.VaultPodSpec(&dc.Spec.Template.Spec, []string{"bash", "-c", "/opt/run.sh"}, spec.Spec.VaultRegistration)

	err = utils.CreateRuntimeObjectContextWrapper(ctx, dc, dc.ObjectMeta)

	if err != nil {
		return &core.ExecutionError{Msg: "Error happened on processing prometheus exporter deployment config. Error: " + err.Error()}
	}

	log.Debug("Waiting for prometheus exporter is ready")
	podLabelSelector := map[string]string{
		utils.Name: utils.MongoPrometheusExporter,
	}
	err = helperImpl.WaitForPodsReady(
		podLabelSelector,
		request.Namespace,
		1,
		spec.Spec.WaitSeconds)

	if err != nil {
		return &core.ExecutionError{Msg: "Error happened while waiting prometheus exporter pod is ready. Error: " + err.Error()}
	}

	return nil
}
