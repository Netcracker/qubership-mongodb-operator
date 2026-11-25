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

package utils

import (
	"context"
	"time"

	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func WaitForDeploymentReady(ctx core.ExecutionContext, labelSelectors map[string]string, namespace string, numberOfPods int, waitSeconds int) error {
	helperImpl := ctx.Get(KubernetesHelperImpl).(*core.DefaultKubernetesHelperImpl)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)

	return wait.PollImmediate(time.Second, time.Second*time.Duration(waitSeconds), func() (done bool, err error) {
		podsReady, err := checkPodsByLabel(helperImpl.Client, labelSelectors, namespace, numberOfPods, v1.PodRunning,
			func(status v1.ContainerStatus) (bool, error) {
				return status.Ready, nil
			})
		if err != nil || !podsReady {
			return false, err
		}

		dep := &appsv1.Deployment{}
		err = helperImpl.Client.Get(context.Background(), client.ObjectKey{
			Namespace: namespace,
			Name:      BackupDaemon,
		}, dep)
		if err != nil {
			log.Warn("Failed to get Deployment", zap.Error(err))
			return false, err
		}

		// Controller hasn't processed the latest spec yet
		if dep.Status.ObservedGeneration < dep.Generation {
			log.Debug("Waiting for deployment controller to observe new generation",
				zap.Int64("observed", dep.Status.ObservedGeneration),
				zap.Int64("generation", dep.Generation),
			)
			return false, nil
		}

		if dep.Status.UpdatedReplicas < *dep.Spec.Replicas {
			log.Debug("Waiting for new replicas to be updated",
				zap.Int32("updatedReplicas", dep.Status.UpdatedReplicas),
				zap.Int32("desired", *dep.Spec.Replicas),
			)
			return false, nil
		}

		if dep.Status.Replicas > dep.Status.UpdatedReplicas {
			log.Info("Waiting for old replicas to terminate",
				zap.Int32("replicas", dep.Status.Replicas),
				zap.Int32("updatedReplicas", dep.Status.UpdatedReplicas),
			)
			return false, nil
		}

		if dep.Status.AvailableReplicas < *dep.Spec.Replicas {
			log.Debug("Waiting for replicas to become available",
				zap.Int32("availableReplicas", dep.Status.AvailableReplicas),
				zap.Int32("desired", *dep.Spec.Replicas),
			)
			return false, nil
		}

		log.Info("Deployment rollout complete",
			zap.String("deployment", BackupDaemon),
			zap.Int32("availableReplicas", dep.Status.AvailableReplicas),
		)

		return true, nil
	})
}

func checkPodsByLabel(K8sclient client.Client, labelSelectors map[string]string, namespace string, numberOfPods int, podPhase v1.PodPhase, containerCheckFunc func(status v1.ContainerStatus) (bool, error)) (done bool, err error) {
	podList := &v1.PodList{}

	listOps := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabelsSelector{Selector: labels.SelectorFromSet(labelSelectors)},
	}
	if err := K8sclient.List(context.Background(), podList, listOps...); err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	if len(podList.Items) == numberOfPods {
		for _, pod := range podList.Items {
			if pod.Status.Phase == podPhase {
				for _, containerStatus := range pod.Status.ContainerStatuses {
					result, err := containerCheckFunc(containerStatus)
					if err != nil {
						return false, err
					}
					if !result {
						return false, nil
					}
				}
			} else {
				return false, nil
			}
		}
		return true, nil
	}
	return false, nil
}
