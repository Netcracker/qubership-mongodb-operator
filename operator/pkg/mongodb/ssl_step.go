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
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type CreateSSLSecretStep struct {
	core.DefaultExecutable
}

func (r *CreateSSLSecretStep) Execute(ctx core.ExecutionContext) error {
	var request reconcile.Request = ctx.Get(constants.ContextRequest).(reconcile.Request)
	helperImpl := ctx.Get(utils.KubernetesHelperImpl).(core.KubernetesHelper)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)

	log.Info("Mongo SSL step started")
	var sslString string
	if spec.Spec.SchemaSettings.SchemaType == v1alpha1.DR {
		url := fmt.Sprintf("http://%s.%s.svc.%s:8069", utils.OperatorServiceName,
			request.Namespace, spec.Spec.SchemaSettings.OtherDomainName)

		client := utils.NewOperatorClinet(url)
		result, err := client.GetKeyFile()
		if err == nil {
			sslString = result
		} else if spec.Spec.DisasterRecovery.Mode == utils.StandbyMode && core.GetCurrentDeployType(ctx) == core.CleanDeploy {
			sslResult, sslErr := helperImpl.OpensslCommand([]string{"rand", "-base64", "755"})
			core.PanicError(sslErr, log.Error, "Can't create mongo-secret")
			sslString = string(sslResult)
		} else {
			core.PanicError(err, log.Error, fmt.Sprintf("Failed to recieve keyfile from %s. Make sure that other operator is running ", url))
		}

	} else {
		result, err := helperImpl.OpensslCommand([]string{"rand", "-base64", "755"})
		core.PanicError(err, log.Error, "Can't create mongo-secret")
		sslString = string(result)
	}

	secretTemplate := utils.SecretTemplate(
		utils.MongoSecret,
		map[string]string{
			utils.MongoSecretKeyFile: string(sslString),
		},
		request.Namespace)

	err := utils.CreateRuntimeObjectContextWrapper(ctx, secretTemplate, secretTemplate.ObjectMeta)
	core.PanicError(err, log.Error, "Can't create mongo-secret")

	log.Info("Mongo SSL has been successfully created")
	return nil
}
