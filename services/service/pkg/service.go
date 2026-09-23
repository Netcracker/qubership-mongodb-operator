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

package pkg

import (
	"fmt"
	"strings"

	driver "github.com/Netcracker/qubership-mongodb-driver"
	"github.com/Netcracker/qubership-mongodb-supplementary/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-supplementary/pkg/backup"
	"github.com/Netcracker/qubership-mongodb-supplementary/pkg/common"
	"github.com/Netcracker/qubership-mongodb-supplementary/pkg/dbaas"
	"github.com/Netcracker/qubership-mongodb-supplementary/pkg/prometheus_exporter"
	"github.com/Netcracker/qubership-mongodb-supplementary/pkg/robotTests"
	"github.com/Netcracker/qubership-mongodb-supplementary/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type MongoServicesCompound struct {
	core.DefaultCompound
}

type MongoServiceBuilder struct {
	core.ExecutableBuilder
}

func (r *MongoServiceBuilder) Build(ctx core.ExecutionContext) core.Executable {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbSupplService)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	client := ctx.Get(constants.ContextClient).(client.Client)
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)

	log.Debug("Mongo Executable build process is started")

	// Checking of common params that affect deployment of some microservices
	//TODO same as in cassandra
	var depth int = 1
	names := make(map[string]interface{})
	core.GetFieldsAndNamesByTag(names, "true", "common", spec.Spec, &depth)
	isAnyParamChanged, commonParamCheckErr := core.HasSpecChanged(
		ctx, func(cfgTemplate *v1.ConfigMap) bool {
			resultCheck := false
			for specKey, specToCheck := range names {
				specHasChanges := core.CompareSpecToCM(ctx, cfgTemplate, specToCheck, specKey)

				if specHasChanges {
					resultCheck = true
				}
			}
			return resultCheck
		},
	)
	core.PanicError(commonParamCheckErr, log.Error, "Error happened during checking common parameters for changes")
	ctx.Set(utils.IsAnyCommonParameterChanged, isAnyParamChanged)

	// It is needed for test proposes. Implementations is changed for module tests
	// TODO: Force key change based on deploy type?
	defaultUtilsHelper := &core.DefaultKubernetesHelperImpl{
		ForceKey: false,
		OwnerKey: true,
		Client:   client,
	}

	ctx.Set(utils.KubernetesHelperImpl, defaultUtilsHelper)

	// Default mongo auth
	//ctx.Set(utils.ContextMongoCMD, fmt.Sprintf(utils.MongoCMD, spec.Spec.AuthDb))

	// Setting IP template for metrics collection
	ipTemplate := "%(ip)s"
	if spec.Spec.IpV6 {
		ipTemplate = "[" + ipTemplate + "]"
	}
	ctx.Set(utils.MonitoringIPTemplate, ipTemplate)

	// Default Schema Type
	if spec.Spec.SchemaSettings.SchemaType == "" {
		log.Warn("Schema type is not set in CR. Setting default one...")
		spec.Spec.SchemaSettings.SchemaType = v1alpha1.HA
	}

	mongoHost := fmt.Sprintf("%s.%s", utils.Mongos, request.Namespace)
	port := 27017
	if !spec.Spec.SchemaSettings.Sharded && spec.Spec.SchemaSettings.SchemaType != v1alpha1.Single {
		mongoHost = strings.Join(utils.GetDATAReplicaSetHostName(spec.Spec.SchemaSettings.DataReplicaSize, 0, spec.Spec.SchemaSettings.ThisDomainName, request.Namespace), ",")
		port = 0
	}

	secret, secErr := core.ReadSecret(client, utils.MongoRootCreds, request.Namespace)
	core.PanicError(secErr, log.Error, fmt.Sprintf("Failed to read %s", utils.MongoRootCreds))

	pass := string(secret.Data[utils.Password])
	user := string(secret.Data[utils.Username])

	//TODO
	mongodService := &driver.MongoServiceImpl{
		Logger: log,
		Configuration: &driver.MongoConfigurationImpl{
			Hostname:   mongoHost,
			Port:       port,
			User:       user,
			Pass:       pass,
			AuthDb:     "admin",
			TLSEnabled: spec.Spec.TLS.Enabled,
			CAPath:     utils.RootCertPath + "ca.crt",
		},
	}

	ctx.Set(utils.ContextMongoHost, mongoHost)
	ctx.Set(utils.ContextMongoService, mongodService)
	ctx.Set(utils.ContextCredsManager, &utils.CredsManager{})

	var compound core.ExecutableCompound = &MongoServicesCompound{}

	if spec.Spec.SchemaSettings.SchemaType == v1alpha1.DR && spec.Spec.DisasterRecovery.Mode == utils.DisableMode {
		return compound
	}

	if spec.Spec.DisasterRecovery.Mode == utils.ActiveMode || spec.Spec.SchemaSettings.SchemaType != v1alpha1.DR {
		compound.AddStep(&common.AddServicesUsers{})
	}

	if spec.Spec.Backup.Install {
		compound.AddStep((&backup.BackupBuilder{}).Build(ctx))
	}

	if spec.Spec.Dbaas.Install {
		compound.AddStep(&dbaas.UpdateDbaasUserStep{})
		compound.AddStep((&dbaas.DbaasBuilder{}).Build(ctx))
	}

	if spec.Spec.PrometheusExporter.Install {
		compound.AddStep((&prometheus_exporter.PrometheusExporterBuilder{}).Build(ctx))
	}

	if spec.Spec.RobotTests.Install {
		compound.AddStep((&robotTests.RobotBuilder{}).Build(ctx))
	}

	if spec.Spec.DisasterRecovery.Mode == utils.ActiveMode || spec.Spec.SchemaSettings.SchemaType != v1alpha1.DR {
		compound.AddStep(&common.AddServicesUsers{})
	}

	return compound
}

type PreDeployBuilder struct {
	core.ExecutableBuilder
}

func (r *PreDeployBuilder) Build(ctx core.ExecutionContext) core.Executable {
	var compound core.ExecutableCompound = &MongoServicesCompound{}
	// spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbSupplService)
	client := ctx.Get(constants.ContextClient).(client.Client)

	//TODO DR
	defaultKubernetesHelper := &core.DefaultKubernetesHelperImpl{
		ForceKey: false,
		OwnerKey: true,
		Client:   client,
	}
	ctx.Set(utils.KubernetesHelperImpl, defaultKubernetesHelper)

	return compound
}
