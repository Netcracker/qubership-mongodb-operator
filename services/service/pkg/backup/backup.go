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
	"context"
	"fmt"
	"time"

	"github.com/Netcracker/qubership-mongodb-supplementary/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-supplementary/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/steps"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
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
	kubeClient := ctx.Get(constants.ContextClient).(ctrlclient.Client)

	storage := spec.Spec.Backup.Storage

	backupCreds, bErr := core.ReadSecret(kubeClient, spec.Spec.Backup.BackupSecretName, request.Namespace)
	restoreCreds, rErr := core.ReadSecret(kubeClient, spec.Spec.Backup.RestoreUserSecretName, request.Namespace)
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
		backupWaitSeconds := spec.Spec.WaitSeconds
		backup.AddStep(&steps.WaitForPVCExpansionStep{
			WaitTimeout: backupWaitSeconds,
			PVCNamesVar: utils.BackupPvcNames,
			OnNeedsRestart: func(ctx core.ExecutionContext) error {
				req := ctx.Get(constants.ContextRequest).(reconcile.Request)
				log := ctx.Get(constants.ContextLogger).(*zap.Logger)

				depl := &appsv1.Deployment{}
				err := kubeClient.Get(context.Background(), types.NamespacedName{Name: utils.BackupDaemon, Namespace: req.Namespace}, depl)
				if err != nil {
					if k8sapierrors.IsNotFound(err) {
						log.Info(fmt.Sprintf("Deployment %s not found, skipping restart", utils.BackupDaemon))
						return nil
					}
					return fmt.Errorf("getting deployment %s: %w", utils.BackupDaemon, err)
				}

				// scale down to 0
				log.Info(fmt.Sprintf("Scaling deployment %s down for volume resize", utils.BackupDaemon))
				zero := int32(0)
				patch := ctrlclient.MergeFrom(depl.DeepCopy())
				depl.Spec.Replicas = &zero
				if err := kubeClient.Patch(context.Background(), depl, patch); err != nil {
					return fmt.Errorf("scaling down %s: %w", utils.BackupDaemon, err)
				}
				if err := wait.PollImmediate(2*time.Second, time.Duration(backupWaitSeconds)*time.Second, func() (bool, error) {
					cur := &appsv1.Deployment{}
					if err := kubeClient.Get(context.Background(), types.NamespacedName{Name: utils.BackupDaemon, Namespace: req.Namespace}, cur); err != nil {
						return false, err
					}
					return cur.Status.Replicas == 0, nil
				}); err != nil {
					return fmt.Errorf("waiting for %s to scale down: %w", utils.BackupDaemon, err)
				}

				// scale up with retry
				const maxAttempts = 5
				for attempt := 1; attempt <= maxAttempts; attempt++ {
					delay := time.Duration(10*(1<<uint(attempt-1))) * time.Second
					log.Info(fmt.Sprintf("Attempt %d/%d: waiting %s before starting %s", attempt, maxAttempts, delay, utils.BackupDaemon))
					time.Sleep(delay)

					cur := &appsv1.Deployment{}
					if err := kubeClient.Get(context.Background(), types.NamespacedName{Name: utils.BackupDaemon, Namespace: req.Namespace}, cur); err != nil {
						return fmt.Errorf("getting deployment %s: %w", utils.BackupDaemon, err)
					}
					scaleUpPatch := ctrlclient.MergeFrom(cur.DeepCopy())
					one := int32(1)
					cur.Spec.Replicas = &one
					if err := kubeClient.Patch(context.Background(), cur, scaleUpPatch); err != nil {
						return fmt.Errorf("scaling up %s: %w", utils.BackupDaemon, err)
					}

					checkErr := wait.PollImmediate(5*time.Second, 2*time.Minute, func() (bool, error) {
						updated := &appsv1.Deployment{}
						if err := kubeClient.Get(context.Background(), types.NamespacedName{Name: utils.BackupDaemon, Namespace: req.Namespace}, updated); err != nil {
							return false, err
						}
						return updated.Status.AvailableReplicas >= 1 && updated.Status.ReadyReplicas >= 1, nil
					})
					if checkErr == nil {
						log.Info(fmt.Sprintf("Deployment %s started successfully on attempt %d", utils.BackupDaemon, attempt))
						return nil
					}
					log.Warn(fmt.Sprintf("Deployment %s not healthy on attempt %d, retrying", utils.BackupDaemon, attempt))
					if attempt < maxAttempts {
						scaleDownPatch := ctrlclient.MergeFrom(cur.DeepCopy())
						cur.Spec.Replicas = &zero
						_ = kubeClient.Patch(context.Background(), cur, scaleDownPatch)
						_ = wait.PollImmediate(2*time.Second, time.Duration(backupWaitSeconds)*time.Second, func() (bool, error) {
							c := &appsv1.Deployment{}
							if err := kubeClient.Get(context.Background(), types.NamespacedName{Name: utils.BackupDaemon, Namespace: req.Namespace}, c); err != nil {
								return false, err
							}
							return c.Status.Replicas == 0, nil
						})
					}
				}
				return fmt.Errorf("deployment %s failed to start after %d attempts", utils.BackupDaemon, maxAttempts)
			},
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
