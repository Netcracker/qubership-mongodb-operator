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

const (
	robotTestsAtpStorageSecretName = "mongodb-robot-tests-atp-storage-secret"
	robotTestsPodSecretsMountPath  = "/etc/secrets/robot-tests-pod-secrets"
)

type RobotDeployment struct {
	core.DefaultExecutable
}

func (r *RobotDeployment) Execute(ctx core.ExecutionContext) error {
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbSupplService)
	robot := spec.Spec.RobotTests
	helperImpl := ctx.Get(utils.KubernetesHelperImpl).(core.KubernetesHelper)
	credsManager := ctx.Get(utils.ContextCredsManager).(utils.CredsManagerI)
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

	if robot.AtpReport.Enabled {
		volumes, volumeMounts, envs = appendAtpReportPodConfig(robot, volumes, volumeMounts, envs)
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
		spec.Spec.RobotTests.PriorityClassName,
		spec.Spec.ServiceAccountName,
		spec.Spec.RobotTests.Affinity,
		spec.Spec.ImagePullPolicy,
		volumeMounts,
		volumes)

	secretNames := []string{spec.Spec.MongoDB.MongoRootSecretName}
	if robot.AtpReport.Enabled {
		secretNames = append(secretNames, robotTestsAtpStorageSecretName)
	}
	err := credsManager.AddCredHashToPodTemplate(
		secretNames,
		&dc.Spec.Template,
	)
	if err != nil {
		log.Error(fmt.Sprintf("can't add secret HASH to annotations for %s", dc.Name), zap.Error(err))
		return err
	}

	err = helperImpl.DeleteDeploymentAndPods(dc.Name, dc.Namespace, spec.Spec.WaitSeconds)
	core.PanicError(err, log.Error, "RobotTests deployment config processing failed")

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

func appendAtpReportPodConfig(
	robot v1alpha1.RobotTests,
	volumes []v12.Volume,
	volumeMounts []v12.VolumeMount,
	envs []v12.EnvVar,
) ([]v12.Volume, []v12.VolumeMount, []v12.EnvVar) {
	atpSecretMode := int32(420)
	volumes = append(volumes, v12.Volume{
		Name: "robot-tests-pod-secrets",
		VolumeSource: v12.VolumeSource{
			Projected: &v12.ProjectedVolumeSource{
				DefaultMode: &atpSecretMode,
				Sources: []v12.VolumeProjection{
					{
						Secret: &v12.SecretProjection{
							LocalObjectReference: v12.LocalObjectReference{
								Name: robotTestsAtpStorageSecretName,
							},
							Items: []v12.KeyToPath{
								{Key: "atp-storage-username", Path: "ATP_STORAGE_USERNAME"},
								{Key: "atp-storage-password", Path: "ATP_STORAGE_PASSWORD"},
							},
						},
					},
				},
			},
		},
	})
	volumeMounts = append(volumeMounts, v12.VolumeMount{
		Name:      "robot-tests-pod-secrets",
		MountPath: robotTestsPodSecretsMountPath,
		ReadOnly:  true,
	})

	envs = append(envs,
		cUtils.GetPlainTextEnvVar("INTEGRATION_TESTS_SECRETS_DIR", robotTestsPodSecretsMountPath),
		cUtils.GetPlainTextEnvVar("ATP_REPORT_ENABLED", "true"),
		cUtils.GetPlainTextEnvVar("ATP_STORAGE_PROVIDER", robot.AtpReport.AtpStorage.Provider),
		cUtils.GetPlainTextEnvVar("ATP_STORAGE_SERVER_URL", robot.AtpReport.AtpStorage.ServerUrl),
		cUtils.GetPlainTextEnvVar("ATP_STORAGE_SERVER_UI_URL", robot.AtpReport.AtpStorage.ServerUiUrl),
		cUtils.GetPlainTextEnvVar("ATP_STORAGE_BUCKET", robot.AtpReport.AtpStorage.Bucket),
		cUtils.GetPlainTextEnvVar("ATP_STORAGE_REGION", robot.AtpReport.AtpStorage.Region),
		cUtils.GetPlainTextEnvVar("ATP_REPORT_VIEW_UI_URL", robot.AtpReportViewUiUrl),
		cUtils.GetPlainTextEnvVar("ENVIRONMENT_NAME", robot.EnvironmentName),
		cUtils.GetPlainTextEnvVar("ENABLE_JIRA_INTEGRATION", strconv.FormatBool(robot.EnableJiraIntegration)),
	)

	return volumes, volumeMounts, envs
}

func sanitizeVolumeName(name string) string {
	return strings.ReplaceAll(name, ".", "-")
}
