package mongodb

import (
	"fmt"
	"reflect"
	"time"

	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/dr"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/steps"
	"go.uber.org/zap"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
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
	mongoWaitSeconds := spec.Spec.WaitSeconds
	mongo.AddStep(&steps.WaitForPVCExpansionStep{
		WaitTimeout: mongoWaitSeconds,
		PVCNamesVar: utils.PvcNames,
		OnNeedsRestart: func(ctx core.ExecutionContext) error {
			helperImpl := ctx.Get(utils.KubernetesHelperImpl).(core.KubernetesHelper)
			mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)
			req := ctx.Get(constants.ContextRequest).(reconcile.Request)
			log := ctx.Get(constants.ContextLogger).(*zap.Logger)
			pvcNames, _ := ctx.Get(utils.PvcNames).([]string)
			schema := spec.Spec.SchemaSettings

			for N, pvcName := range pvcNames {
				// Derive StatefulSet names for this PVC index from schema.
				// PVC-N is shared by: cnfrs{N} (config server replica N)
				// and datars{shard}{N} for each shard (data replica N of each shard).
				var ssNames []string
				if !singleSchema {
					if N < schema.CnfReplicaSize {
						ssNames = append(ssNames, fmt.Sprintf("cnfrs%d", N))
					}
					for shard := 1; shard <= schema.ShardCount; shard++ {
						if N < schema.DataReplicaSize {
							ssNames = append(ssNames, fmt.Sprintf("datars%d%d", shard, N))
						}
					}
				} else {
					// Single schema: all mongo pods share the same PVCs; collect SS names from pods.
					pods, err := helperImpl.ListPods(req.Namespace, map[string]string{utils.Microservice: utils.MongoCluster})
					if err != nil {
						return fmt.Errorf("listing MongoDB pods: %w", err)
					}
					seen := map[string]bool{}
					for _, pod := range pods.Items {
						for _, ref := range pod.OwnerReferences {
							if ref.Kind == "StatefulSet" && !seen[ref.Name] {
								ssNames = append(ssNames, ref.Name)
								seen[ref.Name] = true
							}
						}
					}
				}

				if len(ssNames) == 0 {
					log.Info(fmt.Sprintf("No StatefulSets for PVC %s, skipping", pvcName))
					continue
				}

				log.Info(fmt.Sprintf("PVC %s group: %v", pvcName, ssNames))

				// Save original replicas and scale down the whole group.
				originalReplicas := map[string]int32{}
				for _, ssName := range ssNames {
					ss, err := helperImpl.GetStatefulSetByName(ssName, req.Namespace)
					if err != nil {
						return fmt.Errorf("getting StatefulSet %s: %w", ssName, err)
					}
					replicas := int32(1)
					if ss.Spec.Replicas != nil {
						replicas = *ss.Spec.Replicas
					}
					originalReplicas[ssName] = replicas
					log.Info(fmt.Sprintf("Scaling down %s", ssName))
					if err := helperImpl.ScaleStatefulSetByName(ssName, req.Namespace, 0, mongoWaitSeconds); err != nil {
						return fmt.Errorf("scaling down %s: %w", ssName, err)
					}
				}

				// All pods gone; wait for Cinder to detach the volume and complete resize.
				// VolumeAttachment is cluster-scoped and may not be accessible, so use a fixed wait.
				log.Info(fmt.Sprintf("All pods for PVC %s down, waiting 60s for volume detach and resize", pvcName))
				time.Sleep(60 * time.Second)

				// Scale back up.
				for _, ssName := range ssNames {
					replicas := originalReplicas[ssName]
					log.Info(fmt.Sprintf("Scaling up %s to %d", ssName, replicas))
					if err := helperImpl.ScaleStatefulSetByName(ssName, req.Namespace, int(replicas), mongoWaitSeconds); err != nil {
						return fmt.Errorf("scaling up %s: %w", ssName, err)
					}
				}

				// Gate on MongoDB cluster health before moving to the next PVC.
				log.Info(fmt.Sprintf("Waiting for MongoDB health after PVC %s expansion", pvcName))
				(&UpdateContextAuthMongo{}).Execute(ctx)
				if err := wait.PollImmediate(10*time.Second, time.Duration(mongoWaitSeconds)*time.Second,
					func() (bool, error) {
						status, err := mongoImpl.GetClusterStatus(
							spec.Spec.DisasterRecovery.Mode,
							schema.ThisDomainName,
							schema.CnfReplicaSize,
							schema.DataReplicaSize,
							schema.ShardCount,
							schema.Sharded,
						)
						if err != nil {
							log.Warn(fmt.Sprintf("MongoDB health check error: %v", err))
							return false, nil
						}
						return status == utils.Up, nil
					},
				); err != nil {
					return fmt.Errorf("MongoDB not healthy after PVC %s expansion: %w", pvcName, err)
				}
			}
			return nil
		},
	})
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
		mongo.AddStep(&dr.RenameMemberDomainStep{})
		mongo.AddStep(&dr.AddCNFReplicas{})
		mongo.AddStep(&dr.AddDATAReplicas{})
	}

	return &mongo
}
