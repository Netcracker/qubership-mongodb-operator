package mongodb

import (
	"fmt"
	"reflect"

	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/dr"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/steps"
	"go.uber.org/zap"
	v12 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type MongoDB struct {
	core.MicroServiceCompound
}

func (r *MongoDB) Validate(ctx core.ExecutionContext) error {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	//TO DO --> fix for non-tls to tls upgrade

	// schema := spec.Spec.SchemaSettings
	// mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)
	// deploymentType, err := r.CalcDeployType(ctx)
	// if err != nil {
	// 	return &core.ExecutionError{Msg: fmt.Sprintf("Failed to get deployment type, err: %s", err.Error())}
	// }
	// if spec.Spec.SchemaSettings.SchemaType != v1.DR && deploymentType != core.CleanDeploy {
	// 	(&UpdateContextAuthMongo{}).Execute(ctx)
	// 	status, err := mongoImpl.GetClusterStatus(spec.Spec.DisasterRecovery.Mode, schema.ThisDomainName, schema.CnfReplicaSize, schema.DataReplicaSize, schema.ShardCount, spec.Spec.SchemaSettings.Sharded)
	// 	if status == utils.Down || err != nil {
	// 		return &core.ExecutionError{Msg: fmt.Sprintf("cluster status is Down. Check /rs-status operator endpoint for more information., err: %s", err.Error())}
	// 	}
	// 	fcvMismatch, err := mongoImpl.CheckFCV(spec.Spec.SchemaSettings.ShardCount)
	// 	if err != nil {
	// 		return &core.ExecutionError{Msg: fmt.Sprintf("Failed to get FCVs, err: %s", err.Error())}
	// 	}
	// 	if fcvMismatch {
	// 		return &core.ExecutionError{Msg: "FeatureCompatibilityVersion mismatch detected."}
	// 	}
	// }
	if reflect.ValueOf(spec).IsNil() {
		return &core.ExecutionError{Msg: "MongoService CR spec is not found"}
	}
	mongoDbSpec := &spec.Spec.MongoDB

	if mongoDbSpec == nil {
		return &core.ExecutionError{Msg: "MongoDB spec is empty"}
	}

	return r.DefaultCompound.Validate(ctx)
}

func (r *MongoDB) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	microServiceCheck, microserviceCheckErr := core.CheckSpecChange(ctx, spec.Spec.MongoDB, utils.MongoCluster)
	commonCheck := ctx.Get(utils.IsAnyCommonParameterChanged).(bool)

	if !microServiceCheck &&
		spec.Spec.SchemaSettings.SchemaType == v1alpha1.DR && spec.Spec.DisasterRecovery.Mode != utils.ActiveMode {
		mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)
		schema := spec.Spec.SchemaSettings

		(&UpdateContextAuthMongo{}).Execute(ctx)
		status, err := mongoImpl.GetClusterStatus(spec.Spec.DisasterRecovery.Mode, schema.ThisDomainName, schema.CnfReplicaSize, schema.DataReplicaSize, schema.ShardCount, spec.Spec.SchemaSettings.Sharded)
		if err != nil {
			return true, nil
		}
		if status == utils.Down {
			return true, nil
		}
	}

	if microserviceCheckErr != nil {
		log.Info("Failed to check Spec changes for Mongo service, so starting reconcile.")
		return microServiceCheck, microserviceCheckErr
	} else {
		log.Info(fmt.Sprintf("Mongo spec changed: %v, common parameters Spec changed: %v.", microServiceCheck, commonCheck))
		return microServiceCheck || commonCheck, nil
	}
}

type UpdateRootPasswordCompound struct {
	core.DefaultCompound
}

type DeleteStatefulsetsCompound struct {
	core.DefaultCompound
}

func (r *DeleteStatefulsetsCompound) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)
	schema := spec.Spec.SchemaSettings
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)

	if spec.Spec.DisasterRecovery.Mode == utils.ActiveMode || schema.SchemaType != v1alpha1.DR ||
		core.GetCurrentDeployType(ctx) == core.CleanDeploy {
		return false, nil
	}

	url := fmt.Sprintf("http://%s.%s.svc.%s:8069", utils.OperatorServiceName, request.Namespace, spec.Spec.SchemaSettings.OtherDomainName)
	client := utils.NewOperatorClinet(url)
	status, err := client.GetStatus()
	if err != nil {
		log.Warn(fmt.Sprintf("Failed to get another cluster status, err: %s", err.Error()))
		return false, nil
	} else if status != utils.Up {
		log.Warn(fmt.Sprintf("Another cluster status is %s", status))
		return false, nil
	}

	(&UpdateContextAuthMongo{}).Execute(ctx)
	status, err = mongoImpl.GetClusterStatus(spec.Spec.DisasterRecovery.Mode, schema.ThisDomainName, schema.CnfReplicaSize, schema.DataReplicaSize, schema.ShardCount, spec.Spec.SchemaSettings.Sharded)
	if err != nil {
		log.Warn(fmt.Sprintf("Failed to get cluster status, err: %s", err.Error()))
		return true, nil
	}

	if status == utils.Down {
		log.Warn("Cluster status is Down starting PVC cleanup")
		return true, nil
	}

	return false, nil
}

type MongoDBBuilder struct {
	core.ExecutableBuilder
}

func (r *MongoDBBuilder) Build(ctx core.ExecutionContext) core.Executable {
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	singleSchema := spec.Spec.SchemaSettings.SchemaType == v1alpha1.Single

	mongo := MongoDB{}
	pvcSelector := map[string]string{
		utils.Microservice: utils.MongoCluster,
	}

	mongo.ServiceName = utils.MongoCluster
	helperImpl := ctx.Get(utils.KubernetesHelperImpl).(core.KubernetesHelper)
	mongo.CalcDeployType = func(ctx core.ExecutionContext) (core.MicroServiceDeployType, error) {
		deplType, err := helperImpl.GetDeploymentTypeByPVC(ctx, mongo.ServiceName, pvcSelector)
		ctx.Set(utils.MongoDBDeploymentType, deplType)
		return deplType, err
	}

	pvcStep := &steps.CreatePVCStep{
		Storage:           spec.Spec.MongoDB.Storage,
		NameFormat:        fmt.Sprintf(utils.MongoPvcNameFormat, request.Namespace) + "-%v",
		LabelSelector:     pvcSelector,
		ContextVarToStore: utils.PvcNames,
		PVCCount: func(ctx core.ExecutionContext) int {
			return ctx.Get(utils.MaxPVCCountForService).(int)
		},
		WaitTimeout:  spec.Spec.WaitSeconds,
		Owner:        nil,
		WaitPVCBound: spec.Spec.MongoDB.Storage.WaitPVCBound,
	}
	if spec.Spec.DeletePVConUninstall {
		pvcStep.Owner = spec
	}
	mongo.AddStep(pvcStep)
	mongo.AddStep(&steps.StoreNodesStep{
		Storage:           spec.Spec.MongoDB.Storage,
		ContextVarToStore: utils.PVNodes,
	})

	cleanupCompound := DeleteStatefulsetsCompound{}
	cleanupCompound.AddStep(&dr.DeleteDataStatefulsetsStep{})
	cleanupCompound.AddStep(&dr.DeleteConfigStatefulsetsStep{})

	//TODO check when executed - should not be executed if statefulsets not deleted
	if spec.Spec.Recycler.Install {
		var tolerations []v12.Toleration
		if spec.Spec.Policies != nil {
			tolerations = spec.Spec.Policies.Tolerations
		}

		recyclerStep := steps.PVRecyclerStep{
			DockerImage:        spec.Spec.MongoDB.DockerImage,
			Volumes:            spec.Spec.MongoDB.Storage.Volumes,
			Tolerations:        tolerations,
			PVCContextVar:      utils.PvcNames,
			PVNodesContextVar:  utils.PVNodes,
			WaitTimeout:        spec.Spec.WaitSeconds,
			PodSecurityContext: spec.Spec.PodSecurityContext,
			Resources:          spec.Spec.Recycler.Resources,
			Owner:              nil,
		}
		if spec.Spec.DeletePVConUninstall {
			recyclerStep.Owner = spec
		}
		recyclerStep.ConditionFunc = func(ctx core.ExecutionContext) (bool, error) {
			//condition is defined by cleanupCompund Condition
			return true, nil
		}
		cleanupCompound.AddStep(&recyclerStep)

		//TODO debug recycling
		cleanInstallRecycler := recyclerStep
		cleanInstallRecycler.ConditionFunc = nil
		mongo.AddStep(&cleanInstallRecycler)
	}
	mongo.AddStep(&cleanupCompound)

	mongo.AddStep(&CreateSSLSecretStep{})

	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	if singleSchema {
		mongo.AddStep((&SingleMongosStepBuilder{}).Build(ctx))
	} else {
		mongo.AddStep((&CNFStepBuilder{}).Build(ctx))
		mongo.AddStep((&DataStepBuilder{}).Build(ctx))
		mongo.AddStep((&HAMongosStepBuilder{}).Build(ctx))
	}

	creds, rErr := utils.ReadSecret(ctx, spec.Spec.MongoDB.MongoRootSecretName, request.Namespace)
	core.PanicError(rErr, log.Error, "MongoDB Root user credentials secret reading failed")

	username := string(creds.Data[utils.Username])
	password := string(creds.Data[utils.Password])

	//create root user during clean install
	mongo.AddStep(&AddUserStep{
		Username: username,
		Password: password,
		Role:     string(creds.Data[utils.Role]),
		Sharded:  !singleSchema,
		customCondition: func(ctx core.ExecutionContext) (bool, error) {
			return core.GetCurrentDeployType(ctx) == core.CleanDeploy && spec.Spec.DisasterRecovery.Mode == utils.ActiveMode, nil
		},
	})

	mongo.AddStep(&UpdateContextAuthMongo{User: username, Password: password})
	// mongo.AddStep(&UpdateMongoDBCredentials{})

	mongo.AddStep(&SetdefaultWriteConcernStep{})
	mongo.AddStep(&SetFeatureCompatibilityVersionStep{})

	if spec.Spec.SchemaSettings.SchemaType == v1alpha1.DR {
		// Reconfigure replicas as left and right sides
		// mongo.AddStep(&DRConfigurationStep{})
		mongo.AddStep(&dr.AddCNFReplicas{})
		mongo.AddStep(&dr.AddDATAReplicas{})
	}

	return &mongo
}
