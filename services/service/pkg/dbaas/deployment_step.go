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

package dbaas

import (
	"fmt"
	"strconv"

	cUtils "github.com/Netcracker/qubership-nosqldb-operator-core/pkg/utils"

	"github.com/Netcracker/qubership-mongodb-supplementary/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-supplementary/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"go.uber.org/zap"
	v12 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type DbaasDeployment struct {
	core.DefaultExecutable
}

func (r *DbaasDeployment) Execute(ctx core.ExecutionContext) error {
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbSupplService)
	dbaas := spec.Spec.Dbaas
	helperImpl := ctx.Get(utils.KubernetesHelperImpl).(core.KubernetesHelper)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	mongoHost := ctx.Get(utils.ContextMongoHost).(string)
	tlsEnabled := cUtils.IsTLSEnableForDBAAS(spec.Spec.Dbaas.DbaasAggregatorRegistrationAddress, spec.Spec.TLS.Enabled)
	dbaasPort := cUtils.GetHTTPPort(tlsEnabled)
	credsManager := ctx.Get(utils.ContextCredsManager).(utils.CredsManagerI)

	log.Info("Dbaas Deployment initialization step started")

	// Environment variable Start
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

	envs = append(envs,
		cUtils.GetPlainTextEnvVar("MULTI_USERS_ENABLED", strconv.FormatBool(dbaas.MultiUsers)),
	)

	simpleKV := map[string]string{
		"MONGO_HOST":        mongoHost,
		"MONGO_PORT":        "27017",
		"CLOUD_PUBLIC_HOST": spec.Spec.CloudPublicHost,
		"BACKUP_DAEMON_ADDRESS": fmt.Sprintf("%s://%s.%s:%d", cUtils.GetHTTPProtocol(spec.Spec.TLS.Enabled), utils.BackupDaemon, request.Namespace,
			cUtils.GetHTTPPort(spec.Spec.TLS.Enabled)),
		"DBAAS_AGGREGATOR_PHYSICAL_DATABASE_IDENTIFIER": core.OptionalString(dbaas.DbaasAggregatorPhysicalDatabaseIdentifier, request.Namespace),
		"DBAAS_ADAPTER_ADDRESS":                         fmt.Sprintf("%s://%s.%s:%d", cUtils.GetHTTPProtocol(tlsEnabled), utils.DbaasName, request.Namespace, dbaasPort),
		"DBAAS_AGGREGATOR_REGISTRATION_ADDRESS":         dbaas.DbaasAggregatorRegistrationAddress,
		"DBAAS_AGGREGATOR_REGISTRATION_FIXED_DELAY_MS":  fmt.Sprintf("%v", dbaas.DbaasAggregatorRegistrationFixedDelayMS),
		"DBAAS_AGGREGATOR_REGISTRATION_RETRY_DELAY_MS":  fmt.Sprintf("%v", dbaas.DbaasAggregatorRegistrationRetryDelayMS),
		"DBAAS_AGGREGATOR_REGISTRATION_RETRY_TIME_MS":   fmt.Sprintf("%v", dbaas.DbaasAggregatorRegistrationRetryTimeMS),
		"PORT": fmt.Sprint(dbaasPort),
	}
	for key, value := range simpleKV {
		envs = append(envs,
			v12.EnvVar{
				Name:  key,
				Value: value,
			})
	}

	secretVolumes := map[string]string{
		spec.Spec.Dbaas.DbaasAggregatorSecretName:   "/var/run/secrets/mongodb/dbaas-aggregator",
		spec.Spec.Dbaas.DbaasAdminSecretName:        "/var/run/secrets/mongodb/mongo-admin",
		spec.Spec.Dbaas.DbaasRegistrationSecretName: "/var/run/secrets/mongodb/dbaas-registration",
	}

	if spec.Spec.Backup.Install {
		secretVolumes[spec.Spec.Backup.BackupApiSecretName] =
			"/var/run/secrets/mongodb/backup-api"
	}

	secretVolumeMode := int32(256)
	volumes := []v12.Volume{}
	volumeMounts := []v12.VolumeMount{}

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

	dc := DbaasDeploymentTemplate(
		&spec.Spec,
		request.Namespace,
		dbaas.DockerImage,
		dbaas.AdditionalNodeLabels,
		*dbaas.Resources,
		envs,
		tolerations,
		numberOfReplicas,
		dbaasPort,
		spec.Spec.Dbaas.PriorityClassName,
		spec.Spec.Dbaas.Affinity,
		volumeMounts,
		volumes,
	)

	err := credsManager.AddCredHashToPodTemplate([]string{spec.Spec.MongoDB.MongoRootSecretName}, &dc.Spec.Template)
	if err != nil {
		log.Error(fmt.Sprintf("can't add secret HASH to annotations for %s", dc.Name), zap.Error(err))
		return err
	}

	cUtils.TLSSpecUpdate(&dc.Spec.Template.Spec, utils.RootCertPath, spec.Spec.TLS)
	if tlsEnabled {
		cUtils.TLSServerSpecUpdate(&dc.Spec.Template.Spec, spec.Spec.TLS, spec.Spec.Dbaas.TLS.DbaasAdapterCASecretName, utils.ServerCertsPath)
	}

	err = utils.CreateRuntimeObjectContextWrapper(ctx, dc, dc.ObjectMeta)

	if err != nil {
		return &core.ExecutionError{Msg: "Error happened on processing dbaas deployment config. Error: " + err.Error()}
	}

	log.Debug("Waiting for dbaas is ready")
	podLabelSelector := map[string]string{
		utils.Name: utils.DbaasName,
	}
	err = helperImpl.WaitForPodsReady(
		podLabelSelector,
		request.Namespace,
		int(numberOfReplicas),
		spec.Spec.WaitSeconds)

	if err != nil {
		return &core.ExecutionError{Msg: "Error happened while waiting dbaas pod is ready. Error: " + err.Error()}
	}

	return nil
}
