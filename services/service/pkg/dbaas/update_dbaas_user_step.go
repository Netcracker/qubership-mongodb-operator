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
	"context"

	driver "github.com/Netcracker/qubership-mongodb-driver"
	"github.com/Netcracker/qubership-mongodb-supplementary/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-supplementary/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type UpdateDbaasUserStep struct {
	core.DefaultExecutable
}

func (u UpdateDbaasUserStep) Execute(ctx core.ExecutionContext) error {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbSupplService)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	client := ctx.Get(constants.ContextClient).(client.Client)
	creds, rErr := core.ReadSecret(client, spec.Spec.Dbaas.DbaasAdminSecretName, request.Namespace)
	core.PanicError(rErr, log.Error, "Dbaas credentials secret reading failed")
	mongoService := ctx.Get(utils.ContextMongoService).(driver.MongoService)

	_, err := mongoService.GrantRoleToUser(context.Background(), "admin", string(creds.Data[utils.Username]), "readAnyDatabase")
	return err
}

func (u UpdateDbaasUserStep) Condition(ctx core.ExecutionContext) (bool, error) {
	return core.GetCurrentDeployType(ctx) == core.Update, nil
}
