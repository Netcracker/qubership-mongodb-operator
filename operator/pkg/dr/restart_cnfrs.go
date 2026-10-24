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

	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
)

type RestartConfigRSStep struct {
	core.DefaultExecutable
	Replicas int
}

func (s *RestartConfigRSStep) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	return spec.Spec.SchemaSettings.Sharded, nil
}

func (s *RestartConfigRSStep) Execute(ctx core.ExecutionContext) error {
	mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)

	err := mongoImpl.RestartConfigRS(spec.Spec.SchemaSettings.CnfReplicaSize, spec.Spec.SchemaSettings.ThisDomainName, spec.Spec.DisasterRecovery.Mode)

	if err != nil {
		panic(&core.DRExecutionError{Msg: fmt.Sprintf("Failed to restart cnfrs, err: %s", err.Error())})
	}

	return nil
}

type RestartDATARSStep struct {
	core.DefaultExecutable
	Replicas int
}

func (s *RestartDATARSStep) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	return spec.Spec.SchemaSettings.Sharded, nil
}

func (s *RestartDATARSStep) Execute(ctx core.ExecutionContext) error {
	mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)

	err := mongoImpl.RestartDataRS(spec.Spec.SchemaSettings.DataReplicaSize, spec.Spec.SchemaSettings.ShardCount, spec.Spec.SchemaSettings.ThisDomainName, spec.Spec.DisasterRecovery.Mode)

	if err != nil {
		panic(&core.DRExecutionError{Msg: fmt.Sprintf("Failed to restart datars, err: %s", err.Error())})
	}

	return nil
}
