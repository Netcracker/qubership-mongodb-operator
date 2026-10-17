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

package dr

import (
	"fmt"
	"time"

	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/wait"
)

type WaitExpectedClusterStatusStep struct {
	core.DefaultExecutable
	BypassCheck bool
	Status      string
}

func (u *WaitExpectedClusterStatusStep) Condition(ctx core.ExecutionContext) (bool, error) {
	return ctx.Get(utils.MongoDBDeploymentType) == nil || ctx.Get(utils.MongoDBDeploymentType).(core.MicroServiceDeployType) != core.CleanDeploy, nil
}

func (u *WaitExpectedClusterStatusStep) Execute(ctx core.ExecutionContext) error {
	mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment).Spec
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)

	mode := spec.DisasterRecovery.Mode
	if u.BypassCheck {
		mode = utils.StandbyMode
	}

	err := wait.Poll(2*time.Second, time.Duration(spec.WaitSeconds)*time.Second, func() (done bool, err error) {
		status, err := mongoImpl.GetClusterStatus(mode, spec.SchemaSettings.ThisDomainName, spec.SchemaSettings.CnfReplicaSize,
			spec.SchemaSettings.DataReplicaSize, spec.SchemaSettings.ShardCount, spec.SchemaSettings.Sharded)

		if err != nil {
			log.Warn(fmt.Sprintf("Failed to get cluster status, err is %v", err))
		}
		return status == u.Status, nil
	})

	log.Debug("Cluster is up")

	return err
}
