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
	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Add replicas from DR site to Main
type AddCNFReplicas struct {
	core.DefaultExecutable
}

func (r *AddCNFReplicas) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	return spec.Spec.SchemaSettings.Sharded && spec.Spec.DisasterRecovery.Mode == utils.ActiveMode, nil
}

func (r *AddCNFReplicas) Execute(ctx core.ExecutionContext) error {
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)

	rs := utils.GetCNFReplicaSetHostNames(spec.Spec.SchemaSettings.CnfReplicaSize, spec.Spec.SchemaSettings.OtherDomainName, request.Namespace)

	return mongoImpl.AddDRReplicas(map[string]string{utils.Microservice: utils.CnfNameKey}, rs)
}
