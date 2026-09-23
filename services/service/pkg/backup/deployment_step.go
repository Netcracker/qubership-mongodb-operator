// Copyright 2024-2025 NetCracker Technology Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package backup

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Netcracker/qubership-mongodb-supplementary/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-supplementary/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	coreUtils "github.com/Netcracker/qubership-nosqldb-operator-core/pkg/utils"
	"go.uber.org/zap"
	v12 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type BackupDeployment struct {
	core.DefaultExecutable
}

func (r *BackupDeployment) Execute(ctx core.ExecutionContext) error {
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbSupplService)
	backup := spec.Spec.Backup
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	mongoHost := ctx.Get(utils.ContextMongoHost).(string)
	credsManager := ctx.Get(utils.ContextCredsManager).(utils.CredsManagerI)

	log.Info("Backup Deployment initialization step started")

	// configNodesCount := ctx.Get(utils.BackupConfigNodes).(int)

	// configNodes := utils.MongosRegisterNodesString(utils.CnfNameKey, configNodesCount, request.Namespace)
	// log.Debug(fmt.Sprintf("Config nodes: %s", configNodes))

	configCollections := strings.Join(backup.ConfigCollections[:], " ")
	log.Debug(fmt.Sprintf("Config collections: %s", configCollections))

	mongoBackupDb := backup.MongoBackupDB
	mongoSourceDb := backup.MongoSourceDB

	if mongoBackupDb == "" {
		mongoBackupDb = mongoHost
	}
	if mongoSourceDb == "" {
		mongoSourceDb = mongoHost
	}

	var envs []v12.EnvVar

	envs = append(envs,
		coreUtils.GetPlainTextEnvVar("INC_BACKUP_SCHEDULE", backup.IncrementalBackupSchedule),
		coreUtils.GetPlainTextEnvVar("BACKUP_SCHEDULE", backup.BackupSchedule),
		coreUtils.GetPlainTextEnvVar("EVICTION_POLICY", backup.EvictionPolicy),
		coreUtils.GetPlainTextEnvVar("INC_EVICTION_POLICY", backup.IncrementalEvictionPolicy),
		coreUtils.GetPlainTextEnvVar("MONGO_DATABASE_PREFIX_PATTERN", backup.MongoDatabasePrefixPattern),
		coreUtils.GetPlainTextEnvVar("CONFIG_NODES", fmt.Sprintf("cnfrs.%s", request.Namespace)),
		coreUtils.GetPlainTextEnvVar("CONFIG_COLLECTIONS", configCollections), //TODO this parameter is not used in backup daemon src
		coreUtils.GetPlainTextEnvVar("ENABLE_FULL_RESTORE", strconv.FormatBool(backup.EnableFullRestore)),
		coreUtils.GetPlainTextEnvVar("MONGO_BACKUP_DB", mongoBackupDb),
		coreUtils.GetPlainTextEnvVar("MONGO_SOURCE_DB", mongoSourceDb),
		coreUtils.GetPlainTextEnvVar("NUM_PARALLEL_CONNECTIONS", fmt.Sprintf("%v", backup.NumParallelConnections)),
		coreUtils.GetPlainTextEnvVar("GRANULAR_NUM_PARALLEL_CONNECTIONS", fmt.Sprintf("%v", backup.GranularNumParallelConnections)),
		coreUtils.GetPlainTextEnvVar("MONGO_AUTH_DB", "admin"),
		coreUtils.GetPlainTextEnvVar("STORAGE", backup.StorageDirectory),
		coreUtils.GetPlainTextEnvVar("GRANULAR_SCHEDULE", backup.GranularBackupSchedule),
		coreUtils.GetPlainTextEnvVar("SCHEDULED_DBS", strings.Join(backup.GranularBackupScheduledDbs[:], ",")),
		coreUtils.GetPlainTextEnvVar("DATA_VALIDATION_ENABLED", strconv.FormatBool(backup.DataValidationEnabled)),
	)

	secretVolumes := map[string]string{
		spec.Spec.Backup.BackupSecretName:      "/var/run/secrets/mongodb/mongo-backup",
		spec.Spec.Backup.RestoreUserSecretName: "/var/run/secrets/mongodb/mongo-restore",
		spec.Spec.Backup.BackupApiSecretName:   "/var/run/secrets/mongodb/backup-api",
	}

	if spec.Spec.IpV6 {
		envs = append(envs, coreUtils.GetPlainTextEnvVar("BROADCAST_ADDRESS", "::"))
	}

	if backup.S3.Enabled {
		envs = append(envs,
			coreUtils.GetPlainTextEnvVar("S3_ENABLED", strconv.FormatBool(backup.S3.Enabled)),
			coreUtils.GetPlainTextEnvVar("S3_BUCKET", backup.S3.BucketName),
			coreUtils.GetPlainTextEnvVar("S3_URL", backup.S3.EndpointUrl),
			coreUtils.GetPlainTextEnvVar("BACKUP_DAEMON_SECRETS_DIR", "/var/run/secrets/mongodb/s3"),
		)
		secretVolumes[backup.S3.SecretName] =
			"/var/run/secrets/mongodb/s3"

		if backup.S3.SslVerify {
			envs = append(envs, coreUtils.GetPlainTextEnvVar("S3_CERTS_PATH", "/s3Certs"))

		}
	}

	// Environment variable End
	nodeSelector := map[string]string{}
	var pvcName string
	if !backup.Storage.EmptyDir {
		nodeLabels := ctx.Get(fmt.Sprintf(utils.BackupPVNodes)).([]map[string]string)

		if len(nodeLabels) > 0 {
			nodeSelector = nodeLabels[0]
		}
	}

	if !backup.Storage.EmptyDir {
		pvcName = ctx.Get(utils.BackupPvcNames).([]string)[0]
	}

	var tolerations []v12.Toleration
	if spec.Spec.Policies != nil {
		tolerations = spec.Spec.Policies.Tolerations
	}

	var numberOfReplicas int32
	if spec.Spec.DisasterRecovery.Mode == utils.ActiveMode {
		numberOfReplicas = 1
	} else {
		numberOfReplicas = 0
	}

	volumes := []v12.Volume{}
	volumeMounts := []v12.VolumeMount{}
	secretVolumeMode := int32(256)
	for secretName, mountPath := range secretVolumes {
		volumeName := utils.SanitizeVolumeName(secretName)
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

	dc := BackupDeploymentTemplate(
		&spec.Spec,
		pvcName,
		request.Namespace,
		core.ConcatMaps(
			spec.Spec.Backup.AdditionalNodeLabels, // TODO
			nodeSelector),
		// utils.MongoReplicaNodeSelector(nodeLabels, 1, 0, 1, 0)),
		envs,
		tolerations,
		backup.StorageDirectory,
		backup.Storage.EmptyDir,
		numberOfReplicas,
		spec.Spec.Backup.PriorityClassName,
		spec.Spec.Backup.Affinity,
		volumeMounts,
		volumes)

	err := credsManager.AddCredHashToPodTemplate([]string{spec.Spec.MongoDB.MongoRootSecretName}, &dc.Spec.Template)
	if err != nil {
		log.Error(fmt.Sprintf("can't add secret HASH to annotations for %s", dc.Name), zap.Error(err))
		return err
	}

	err = utils.CreateRuntimeObjectContextWrapper(ctx, dc, dc.ObjectMeta)

	if err != nil {
		return &core.ExecutionError{Msg: "Error happened on processing backup deployment config. Error: " + err.Error()}
	}

	log.Debug("Waiting for backup is ready")
	podLabelSelector := map[string]string{
		utils.Name: utils.BackupDaemon,
	}
	err = utils.WaitForDeploymentReady(
		ctx,
		podLabelSelector,
		request.Namespace,
		int(numberOfReplicas),
		spec.Spec.WaitSeconds)

	if err != nil {
		return &core.ExecutionError{Msg: "Error happened while waiting backup pod is ready. Error: " + err.Error()}
	}

	return nil
}
