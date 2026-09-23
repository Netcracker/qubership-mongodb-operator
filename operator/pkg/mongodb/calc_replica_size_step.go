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

package mongodb

import (
	"fmt"

	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"go.uber.org/zap"
)

type CalculateMaxReplicaSizeStep struct {
	core.DefaultExecutable
}

func (r *CalculateMaxReplicaSizeStep) Validate(ctx core.ExecutionContext) error {
	var spec *v1alpha1.MongodbDeployment = ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)

	if spec.Spec.SchemaSettings.SchemaType != v1alpha1.Single {

		if spec.Spec.SchemaSettings.CnfReplicaSize == 0 ||
			spec.Spec.SchemaSettings.DataReplicaSize == 0 ||
			spec.Spec.SchemaSettings.ShardCount == 0 {
			return &core.ExecutionError{Msg: utils.SchemaSettingsValidationZeros}
		}

		if spec.Spec.SchemaSettings.SchemaType != v1alpha1.DR {
			if spec.Spec.SchemaSettings.CnfReplicaSize%2 == 0 ||
				spec.Spec.SchemaSettings.DataReplicaSize%2 == 0 {
				return &core.ExecutionError{Msg: utils.SchemaSettingsValidationHAandArbiter}
			}
		}
	}

	return nil
}

func (r *CalculateMaxReplicaSizeStep) Execute(ctx core.ExecutionContext) error {
	var spec *v1alpha1.MongodbDeployment = ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)

	var res int
	if spec.Spec.SchemaSettings.SchemaType == v1alpha1.Single {
		res = 1
	} else {
		res = core.MaxInt(spec.Spec.SchemaSettings.DataReplicaSize, spec.Spec.SchemaSettings.CnfReplicaSize)
	}

	ctx.Set(utils.MaxReplicaSize, res)
	ctx.Set(utils.MaxPVCCountForService, res)

	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	log.Debug(fmt.Sprintf("Max replica size = %v", res))

	return nil
}
