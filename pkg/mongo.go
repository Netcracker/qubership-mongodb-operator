package impl

import (
	"fmt"

	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/common"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/dr"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/mongodb"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/steps"
	"go.uber.org/zap"
	v1core "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
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
	/*  TODO: Add checking of schema type: non-ha, ha, dr, arbiter with node
	Do we need to extract different builder strategies based on schema type?
	Different deploy types: clean, update, force?
	Clean deploy is supported only for now
	*/
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	kubeConfig := ctx.Get(constants.ContextKubeClient).(*rest.Config)
	client := ctx.Get(constants.ContextClient).(client.Client)

	log.Debug("Mongo Executable build process is started")

	// Checking of common params that affect deployment of some microservices
	//TODO same as in cassandra
	var depth int = 1
	names := make(map[string]interface{})
	core.GetFieldsAndNamesByTag(names, "true", "common", spec.Spec, &depth)
	isAnyParamChanged, commonParamCheckErr := core.HasSpecChanged(
		ctx, func(cfgTemplate *v1core.ConfigMap) bool {
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
		ForceKey: spec.Spec.StopOnFailedResourceUpdate,
		OwnerKey: true,
		Client:   client,
	}
	defaultMongoHelper := &utils.MongoUtilsHelperImpl{
		KubernetesHelperImpl: defaultUtilsHelper,
		Namespace:            request.Namespace,
		KubeConfig:           kubeConfig,
		Logger:               log,
		Client:               client,
		Cmd:                  fmt.Sprintf(utils.MongoCMD, utils.MongoBinary(spec.Spec.MongoDB.DockerImage), spec.Spec.AuthDb),
		Tries:                5,
		RetryInterval:        10,
		WaitSeconds:          spec.Spec.WaitSeconds,
		Sharded:              spec.Spec.SchemaSettings.Sharded,
		Single:               spec.Spec.SchemaSettings.SchemaType == v1alpha1.Single,
	}

	if spec.Spec.TLS.Enabled {
		// cmd := defaultMongoHelper.Cmd
		defaultMongoHelper.Cmd = fmt.Sprint(defaultMongoHelper.Cmd, fmt.Sprintf(" --tls --tlsCAFile %s%s --tlsAllowInvalidCertificates", utils.RootCertPath, spec.Spec.TLS.RootCAFileName))
	}

	ctx.Set(utils.KubernetesHelperImpl, defaultUtilsHelper)
	ctx.Set(utils.MongoHelperImpl, defaultMongoHelper)

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

	// Arbiter Selector
	var res func(maxReplicaSize int) int
	if spec.Spec.SchemaSettings.SchemaType == v1alpha1.Arbiter {
		if spec.Spec.SchemaSettings.ArbiterIndex >= 0 {
			// TODO: Validations for arbiter index value? maxReplicaSize = max(CNFRS, DATARS)
			res = func(maxReplicaSize int) int { return spec.Spec.SchemaSettings.ArbiterIndex }
		} else {
			res = func(maxReplicaSize int) int {
				return maxReplicaSize / 2
			}
		}
	} else {
		res = func(maxReplicaSize int) int {
			return -1
		}
	}
	ctx.Set(utils.ArbiterIndexSelectorFunc, res)

	// TODO: Cleanup steps?

	var compound core.ExecutableCompound = &MongoServicesCompound{}

	if spec.Spec.SchemaSettings.SchemaType == v1alpha1.DR && spec.Spec.DisasterRecovery.Mode == utils.DisableMode {
		return compound
	}

	compound.AddStep(&mongodb.CalculateMaxReplicaSizeStep{})

	if spec.Spec.MongoDB.Install {
		compound.AddStep((&mongodb.MongoDBBuilder{}).Build(ctx))
	}

	compound.AddStep(&mongodb.UpdateContextAuthMongo{})
	log.Debug("Mongo Executable has been built")

	return compound
}

type PreDeployBuilder struct {
	core.ExecutableBuilder
}

func (r *PreDeployBuilder) Build(ctx core.ExecutionContext) core.Executable {
	var compound core.ExecutableCompound = &MongoServicesCompound{}
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	kubeConfig := ctx.Get(constants.ContextKubeClient).(*rest.Config)
	client := ctx.Get(constants.ContextClient).(client.Client)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)

	//TODO DR
	defaultKubernetesHelper := &core.DefaultKubernetesHelperImpl{
		ForceKey: spec.Spec.StopOnFailedResourceUpdate,
		OwnerKey: true,
		Client:   client,
	}
	defaultMongoHelper := &utils.MongoUtilsHelperImpl{
		KubernetesHelperImpl: defaultKubernetesHelper,
		Namespace:            request.Namespace,
		KubeConfig:           kubeConfig,
		Logger:               log,
		Client:               client,
		Tries:                5,
		RetryInterval:        10,
		WaitSeconds:          spec.Spec.WaitSeconds,
		Sharded:              spec.Spec.SchemaSettings.Sharded,
		Single:               spec.Spec.SchemaSettings.SchemaType == v1alpha1.Single,
	}
	ctx.Set(utils.KubernetesHelperImpl, defaultKubernetesHelper)
	ctx.Set(utils.MongoHelperImpl, defaultMongoHelper)

	if !ctx.Get(constants.ContextSpecHasChanges).(bool) {
		if spec.Spec.VaultRegistration.Enabled {
			compound.AddStep(&steps.SetPasswordFromVaultRole{
				Registration:          spec.Spec.VaultRegistration,
				RoleName:              spec.Spec.VaultDBEngine.Role,
				CtxVarToStorePassword: spec.Spec.MongoDB.MongoRootSecretName,
			})
		}
		compound.AddStep(&mongodb.UpdateContextAuthMongo{})
	}

	compound.AddStep(&common.RunFiberServer{})

	return compound
}

type MongoDRCompound struct {
	core.DefaultCompound
}

func (r *MongoDRCompound) Condition(ctx core.ExecutionContext) (bool, error) {
	//TODO do not run in clean install
	return true, nil
}

func (r *MongoDRCompound) Validate(ctx core.ExecutionContext) error {
	return nil
}

type DRBuilder struct {
	core.ExecutableBuilder
}

func (r *DRBuilder) Build(ctx core.ExecutionContext) core.Executable {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)

	//TODO we need to wrap compound so its condition func can be called
	defaultCompound := &core.DefaultCompound{}
	compound := MongoDRCompound{}

	if spec.Spec.SchemaSettings.SchemaType == v1alpha1.DR {
		if spec.Spec.DisasterRecovery.Mode == utils.DisableMode {
			compound.AddStep(&dr.ScaleMongosStep{Replicas: 0})
			compound.AddStep(&dr.ScaleBackupDaemonStep{Replicas: 0})
			compound.AddStep(&dr.ScaleDbaasAdapterStep{Replicas: 0})
			compound.AddStep(&dr.ScaleCNFRSStep{Replicas: 0})
			compound.AddStep(&dr.ScaleDATARSStep{Replicas: 0})
		} else {
			compound.AddStep(&dr.ReconfigureCnfrsStep{})
			compound.AddStep(&dr.ReconfigureDataRSStep{})
			if spec.Spec.DisasterRecovery.Mode == utils.ActiveMode {
				// compound.AddStep(&dr.UpdateConfigRSInDATARSStep{})
				compound.AddStep(&dr.UpdateShardsStep{})
				compound.AddStep(&dr.RestartConfigRSStep{})
				compound.AddStep(&dr.ScaleMongosStep{Replicas: spec.Spec.SchemaSettings.MongosReplicas})
				// compound.AddStep(&dr.RestartDATARSStep{})
				compound.AddStep(&dr.ScaleBackupDaemonStep{Replicas: 1})
				compound.AddStep(&dr.ScaleDbaasAdapterStep{Replicas: 1})

			} else if spec.Spec.DisasterRecovery.Mode == utils.StandbyMode {
				compound.AddStep(&dr.ScaleMongosStep{Replicas: 0})
				compound.AddStep(&dr.ScaleBackupDaemonStep{Replicas: 0})
				compound.AddStep(&dr.ScaleDbaasAdapterStep{Replicas: 0})
			}
			compound.AddStep(&dr.UpdatePrometheusExporterStep{ExportMongos: spec.Spec.DisasterRecovery.Mode == utils.ActiveMode})
			compound.AddStep(&dr.WaitExpectedClusterStatusStep{Status: utils.Up})
		}

	}
	defaultCompound.AddStep(&compound)
	return defaultCompound
}
