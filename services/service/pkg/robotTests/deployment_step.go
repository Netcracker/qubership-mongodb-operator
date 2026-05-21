package robotTests

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Netcracker/qubership-mongodb-supplementary/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-supplementary/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	cUtils "github.com/Netcracker/qubership-nosqldb-operator-core/pkg/utils"
	"go.uber.org/zap"
	v12 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type RobotDeployment struct {
	core.DefaultExecutable
}

func (r *RobotDeployment) Execute(ctx core.ExecutionContext) error {
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbSupplService)
	robot := spec.Spec.RobotTests
	helperImpl := ctx.Get(utils.KubernetesHelperImpl).(core.KubernetesHelper)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	mongoHost := ctx.Get(utils.ContextMongoHost).(string)

	log.Info("RobotTests Deployment initialization step started")

	// Environment variable Start
	envs := []v12.EnvVar{
		v12.EnvVar{
			Name: "OPENSHIFT_WORKSPACE_WA",
			ValueFrom: &v12.EnvVarSource{
				FieldRef: &v12.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
	}

	datarsHost := []string{}
	for i := 0; i < spec.Spec.SchemaSettings.ShardCount; i++ {
		datarsHost = append(datarsHost, fmt.Sprintf(utils.DataNameKey, i+1))

	}

	simpleKV := map[string]string{
		"MONGO_HOST":                  mongoHost,
		"BACKUP_HOST":                 fmt.Sprintf("%s.%s.svc", utils.BackupDaemon, request.Namespace),
		"DBAAS_HOST":                  fmt.Sprintf("%s.%s.svc", utils.DbaasName, request.Namespace),
		"EXTERNAL_BACKUP_PATH":        robot.ExternalBackupPath,
		"TAGS":                        robot.Tags,
		"WAIT_TIMEOUT":                strconv.Itoa(spec.Spec.WaitSeconds),
		"DATARS_HOST":                 strings.Join(datarsHost[:], ","),
		"MAIN_OS_SIDE":                robot.MainSide,
		"LEFT_NODES_PATTERN":          robot.LeftNodesPattern,
		"RIGHT_NODES_PATTERN":         robot.RightNodesPattern,
		"STATUS_CUSTOM_RESOURCE_PATH": fmt.Sprintf("apps/v1/%s/deployments/robot-tests", request.Namespace),
		"STATUS_WRITING_ENABLED":      "true",
		"PORT":                        fmt.Sprint(cUtils.GetHTTPPort(spec.Spec.TLS.Enabled)),
		"CONFIG_NAME":                 "mongodb-tests-config",
		"SUPPLEMENTARY_CONFIG_NAME":   "supplementary-tests-config",
	}

	for key, value := range simpleKV {
		envs = append(envs,
			v12.EnvVar{
				Name:  key,
				Value: value,
			})
	}
	secretVolumes := map[string]string{
		spec.Spec.MongoDB.MongoRootSecretName: "/var/run/secrets/mongodb/mongo-root",
	}

	if spec.Spec.Backup.Install {
		secretVolumes[spec.Spec.Backup.BackupApiSecretName] =
			"/var/run/secrets/mongodb/backup-api"
	}

	if spec.Spec.Dbaas.Install {
		secretVolumes[spec.Spec.Dbaas.DbaasAggregatorSecretName] =
			"/var/run/secrets/mongodb/dbaas-aggregator"
	}
	secretVolumeMode := int32(256)
	volumes := []v12.Volume{}
	volumeMounts := []v12.VolumeMount{}

	for secretName, mountPath := range secretVolumes {

		volumeName := sanitizeVolumeName(secretName)

		volumes = append(volumes, v12.Volume{
			Name: volumeName,
			VolumeSource: v12.VolumeSource{
				Secret: &v12.SecretVolumeSource{
					SecretName:  secretName,
					DefaultMode: &secretVolumeMode,
				},
			},
		})

		volumeMounts = append(volumeMounts, v12.VolumeMount{
			Name:      volumeName,
			MountPath: mountPath,
			ReadOnly:  true,
		})
	}

	// Environment variable  End

	var tolerations []v12.Toleration
	if spec.Spec.Policies != nil {
		tolerations = spec.Spec.Policies.Tolerations
	}

	dc := RobotTemplate(
		request.Namespace,
		robot.DockerImage,
		*robot.Resources,
		robot.NodeLabels,
		tolerations,
		envs,
		spec.Spec.PodSecurityContext,
		spec.Spec.TLS,
		spec.Spec.VaultRegistration,
		spec.Spec.RobotTests.PriorityClassName,
		spec.Spec.ServiceAccountName,
		spec.Spec.RobotTests.Affinity,
		spec.Spec.ImagePullPolicy,
		volumeMounts,
		volumes)

	err := helperImpl.DeleteDeploymentAndPods(dc.Name, dc.Namespace, spec.Spec.WaitSeconds)
	core.PanicError(err, log.Error, "RobotTests deployment config processing failed")

	cUtils.VaultPodSpec(&dc.Spec.Template.Spec, []string{"/docker-entrypoint.sh", "run-robot"}, spec.Spec.VaultRegistration)

	err = utils.CreateRuntimeObjectContextWrapper(ctx, dc, dc.ObjectMeta)
	core.PanicError(err, log.Error, "RobotTests deployment config processing failed")

	log.Debug("Waiting for robot is ready")
	err = helperImpl.WaitForTestsReady(
		dc.Name,
		dc.Namespace,
		spec.Spec.WaitSeconds)
	core.PanicError(err, log.Error, "RobotTests failed")

	return nil
}

func sanitizeVolumeName(name string) string {
	return strings.ReplaceAll(name, ".", "-")
}
