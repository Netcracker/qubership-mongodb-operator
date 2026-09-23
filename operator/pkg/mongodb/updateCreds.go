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

	"github.com/Netcracker/qubership-credential-manager/pkg/manager"
	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type UpdateMongoDBCredentials struct {
	core.DefaultExecutable
}

func (r *UpdateMongoDBCredentials) Execute(ctx core.ExecutionContext) error {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)
	scheme := ctx.Get(constants.ContextSchema).(*runtime.Scheme)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	schema := spec.Spec.SchemaSettings.SchemaType
	sharded := spec.Spec.SchemaSettings.Sharded

	creds, rErr := utils.ReadSecret(ctx, spec.Spec.MongoDB.MongoRootSecretName, request.Namespace)
	core.PanicError(rErr, log.Error, "MongoDB Root user credentials secret reading failed")

	username := string(creds.Data[utils.Username])

	log.Info("Updating mongodb root password")

	err := manager.ActualizeCreds(spec.Spec.MongoDB.MongoRootSecretName, func(newSecret, oldSecret *v1.Secret) error {

		var secret client.Object = newSecret
		err := controllerutil.SetControllerReference(spec, secret, scheme)
		if err != nil {
			return fmt.Errorf("failed to set owner reference to new secret %v, err: %w", newSecret.Name, err)
		}
		return mongoImpl.UpdateRootPassword(spec.Spec.AuthDb, username, string(oldSecret.Data[utils.Password]), string(newSecret.Data[utils.Password]),
			sharded, schema != v1alpha1.Single, spec.Spec.SchemaSettings.ShardCount)
	})
	return err
}
