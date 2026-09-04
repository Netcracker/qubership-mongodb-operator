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
	"github.com/Netcracker/qubership-mongodb-supplementary/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-supplementary/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/steps"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type MongoBackup struct {
	core.MicroServiceCompound
}

type BackupBuilder struct {
	core.ExecutableBuilder
}

func (r *BackupBuilder) Build(ctx core.ExecutionContext) core.Executable {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbSupplService)
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	client := ctx.Get(constants.ContextClient).(client.Client)

	storage := spec.Spec.Backup.Storage

	backupCreds, bErr := core.ReadSecret(client, spec.Spec.Backup.BackupSecretName, request.Namespace)
	restoreCreds, rErr := core.ReadSecret(client, spec.Spec.Backup.RestoreUserSecretName, request.Namespace)
	core.PanicError(bErr, log.Error, "Backup credentials secret reading failed")
	core.PanicError(rErr, log.Error, "Restore credentials secret reading failed")

	users := []utils.UserToAdd{
		{
			User:       string(backupCreds.Data[utils.Username]),
			Pass:       func() string { return string(backupCreds.Data[utils.Password]) },
			Role:       string(backupCreds.Data[utils.Role]),
			ShardLocal: false,
		},
		{
			User:       string(restoreCreds.Data[utils.Username]),
			Pass:       func() string { return string(restoreCreds.Data[utils.Password]) },
			Role:       string(restoreCreds.Data[utils.Role]),
			ShardLocal: false,
		},
	}

	for _, user := range users {
		utils.AddServicesUsersToContext(
			ctx,
			user,
		)
	}

	pvcSelector := map[string]string{
		utils.Name: utils.BackupDaemon,
	}

	backup := MongoBackup{}
	backup.ServiceName = utils.Backup
	helperImpl := ctx.Get(utils.KubernetesHelperImpl).(core.KubernetesHelper)
	backup.CalcDeployType = func(ctx core.ExecutionContext) (core.MicroServiceDeployType, error) {
		return helperImpl.GetDeploymentTypeByPVC(ctx, backup.ServiceName, pvcSelector)
	}
	// For common steps like pvc creation we need to set up context keys used in this steps
	backup.AddStep(&PrepareContextForBackupService{})

	if !spec.Spec.Backup.Storage.EmptyDir {
		pvcStep := &steps.CreatePVCStep{
			Storage:           storage,
			NameFormat:        utils.BackupPvcNameFormat,
			LabelSelector:     pvcSelector,
			ContextVarToStore: utils.BackupPvcNames,
			PVCCount: func(ctx core.ExecutionContext) int {
				return 1
			},
			WaitTimeout:  spec.Spec.WaitSeconds,
			Owner:        nil,
			WaitPVCBound: spec.Spec.Backup.Storage.WaitPVCBound,
		}
		if spec.Spec.DeletePVConUninstall {
			pvcStep.Owner = spec
		}
		backup.AddStep(pvcStep)
		backup.AddStep(&steps.StoreNodesStep{
			Storage:           storage,
			ContextVarToStore: utils.BackupPVNodes,
		})
	}

	backup.AddStep(&BackupService{})
	//backup.AddStep(&BackupSecrets{})
	backup.AddStep(&BackupConfigMaps{})

	backup.AddStep(&BackupDeployment{})

	return &backup
}

func (r *MongoBackup) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbSupplService)
	microServiceCheck, microserviceCheckErr := core.CheckSpecChange(ctx, spec.Spec.Backup, utils.BackupDaemon)
	commonCheck := ctx.Get(utils.IsAnyCommonParameterChanged).(bool)

	if microserviceCheckErr != nil {
		return microServiceCheck, microserviceCheckErr
	} else {
		return microServiceCheck || commonCheck, nil
	}
}
