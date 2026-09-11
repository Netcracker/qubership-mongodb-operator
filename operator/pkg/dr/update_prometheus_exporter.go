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
	"errors"
	"fmt"
	"strconv"

	"github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	v1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type UpdatePrometheusExporterStep struct {
	core.DefaultExecutable
	ExportMongos bool
}

func (s *UpdatePrometheusExporterStep) Execute(ctx core.ExecutionContext) error {
	helperImpl := ctx.Get(utils.KubernetesHelperImpl).(core.KubernetesHelper)
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)

	err := helperImpl.UpdateDeploymentByLabels(map[string]string{utils.Microservice: utils.MongoPrometheusExporter},
		request.Namespace,
		func(depl *v1.Deployment) {
			envs := depl.Spec.Template.Spec.Containers[0].Env
			for i := 0; i < len(envs); i++ {
				if envs[i].Name == "EXPORT_MONGOS" {
					envs[i].Value = strconv.FormatBool(s.ExportMongos)
					break
				}
			}
			depl.Spec.Template.Spec.Containers[0].Env = envs
		})

	var nfe *core.NotFoundError
	if errors.As(err, &nfe) {
		return nil
	}

	if err != nil {
		panic(&core.DRExecutionError{Msg: fmt.Sprintf("Failed to scale deployment %s, err: %s", utils.MongoPrometheusExporter, err.Error())})
	}

	return nil
}
