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

package prometheus_exporter

import (
	"github.com/Netcracker/qubership-mongodb-supplementary/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-supplementary/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type PrometheusExporterCompound struct {
	core.DefaultCompound
}

type PrometheusExporterBuilder struct {
	core.ExecutableBuilder
}

func (r *PrometheusExporterBuilder) Build(ctx core.ExecutionContext) core.Executable {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbSupplService)
	client := ctx.Get(constants.ContextClient).(client.Client)
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)

	creds, rErr := core.ReadSecret(client, spec.Spec.PrometheusExporter.MonitoringSecretName, request.Namespace)
	core.PanicError(rErr, log.Error, "Prometheus credentials secret reading failed")

	exporter := PrometheusExporterCompound{}

	exporter.AddStep(&PrometheusExporterService{})

	users := []utils.UserToAdd{
		{
			User:       string(creds.Data[utils.Username]),
			Pass:       func() string { return string(creds.Data[utils.Password]) },
			Role:       string(creds.Data[utils.Role]),
			ShardLocal: true,
		},
	}

	for _, user := range users {
		utils.AddServicesUsersToContext(
			ctx,
			user,
		)
	}

	exporter.AddStep(&PrometheusExporterDeployment{})

	return &exporter
}

func (r *PrometheusExporterCompound) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbSupplService)
	microServiceCheck, microserviceCheckErr := core.CheckSpecChange(ctx, spec.Spec.PrometheusExporter, utils.MongoPrometheusExporter)
	commonCheck := ctx.Get(utils.IsAnyCommonParameterChanged).(bool)

	if microserviceCheckErr != nil {
		return microServiceCheck, microserviceCheckErr
	} else {
		return microServiceCheck || commonCheck, nil
	}
}
