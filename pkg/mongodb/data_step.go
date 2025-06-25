package mongodb

import (
	"fmt"
	"strconv"

	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"go.uber.org/zap"
	v12 "k8s.io/api/core/v1"
	v13 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type CreateDataStep struct {
	core.DefaultExecutable
}

const DatarsAffinityFunc = "contextDataAffinityFunc"

func (r *CreateDataStep) Validate(ctx core.ExecutionContext) error {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)

	resultFunc := func(dataShardIndex int, replicaIndex int) v12.Affinity {
		// If user defined Affinity, use it
		if spec.Spec.MongoDB.Affinity != nil {
			return *spec.Spec.MongoDB.Affinity
		}

		// Fallback to default operator logic
		affinity := v12.Affinity{
			PodAntiAffinity: &v12.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []v12.PodAffinityTerm{
					{
						LabelSelector: &v13.LabelSelector{
							MatchExpressions: []v13.LabelSelectorRequirement{
								{
									Key:      utils.ServiceName,
									Operator: "In",
									Values: []string{
										fmt.Sprintf(utils.DataNameKey, dataShardIndex+1),
									},
								},
							},
						},
						TopologyKey: utils.KubeHostName,
					},
				},
			},
		}
		if spec.Spec.SchemaSettings.Sharded {
			affinity.PodAffinity = &v12.PodAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []v12.PodAffinityTerm{
					{
						LabelSelector: &v13.LabelSelector{
							MatchExpressions: []v13.LabelSelectorRequirement{
								{
									Key:      utils.MongoNode,
									Operator: "In",
									Values: []string{
										fmt.Sprintf(utils.CnfNameWithIndexFormat, replicaIndex),
									},
								},
							},
						},
						TopologyKey: utils.KubeHostName,
					},
				},
			}
		}
		return affinity
	}

	ctx.Set(DatarsAffinityFunc, resultFunc)

	return nil
}

func (r *CreateDataStep) Execute(ctx core.ExecutionContext) error {
	client := ctx.Get(constants.ContextClient).(client.Client)
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	var spec *v1alpha1.MongodbDeployment = ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	helperImpl := ctx.Get(utils.KubernetesHelperImpl).(core.KubernetesHelper)
	mongoDbSpec := &spec.Spec.MongoDB
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)

	log.Info("Data Shards Creation step started")

	pvcNames := ctx.Get(utils.PvcNames).([]string)
	nodeLabels := ctx.Get(utils.PVNodes).([]map[string]string)
	affinityFunc := ctx.Get(DatarsAffinityFunc).(func(dataShardIndex int, replicaIndex int) v12.Affinity)

	dataReplicaSize := spec.Spec.SchemaSettings.DataReplicaSize
	shardCount := spec.Spec.SchemaSettings.ShardCount
	replicaType := "--shardsvr"

	for s := 0; s < shardCount; s++ {
		nameKey := fmt.Sprintf(utils.DataNameKey, s+1)

		template := utils.ShardServiceTemplate(nameKey, request.Namespace)

		// Kubernetes api causes "invalid resourceVersion error" on update. So remove it.
		core.DeleteRuntimeObject(client, &v12.Service{
			ObjectMeta: template.ObjectMeta,
		})

		err := utils.CreateRuntimeObjectContextWrapper(ctx, template, template.ObjectMeta)

		log.Info(fmt.Sprintf("%s Service has been created", nameKey))

		if err != nil {
			return err
		}

		for i := 0; i < dataReplicaSize; i++ {
			dataWiredCacheGb := utils.CalcProperMongoWiredCacheSize(mongoDbSpec.DataResources.Limits.Memory().Value(), mongoDbSpec.DataWiredTigerCacheGb, 0.25)
			log.Debug(fmt.Sprintf("Wired tiger cache size = %v", dataWiredCacheGb))

			nameWithIndexes := fmt.Sprintf(utils.DataNameWithIndexesFormat, s+1, i)
			_, pvcName := utils.ElementsDistribution(utils.GetDummyPvcMap(pvcNames), dataReplicaSize, i, shardCount, s)

			nodeSelector := core.ConcatMaps(
				mongoDbSpec.AdditionalNodeLabels,
				utils.MongoReplicaNodeSelector(nodeLabels, dataReplicaSize, i, 1, 0))

			dataOrArbiterResource := *mongoDbSpec.DataResources
			if mongoDbSpec.ArbiterResources != nil && spec.Spec.SchemaSettings.SchemaType == v1alpha1.Arbiter {
				arbFunc := ctx.Get(utils.ArbiterIndexSelectorFunc).(func(int) int)
				arbiterNodeIndex := arbFunc(ctx.Get(utils.MaxReplicaSize).(int))
				if i == arbiterNodeIndex {
					dataWiredCacheGb = utils.CalcProperMongoWiredCacheSize(mongoDbSpec.ArbiterResources.Limits.Memory().Value(), "0", 0.25)
					dataOrArbiterResource = *mongoDbSpec.ArbiterResources
				}
			}

			if !spec.Spec.SchemaSettings.Sharded {
				replicaType = ""
			}

			containerArgs := utils.MongoReplicaContainerArgs(replicaType, utils.Data, nameKey, nameWithIndexes, utils.MongoSecret, utils.MongoSecretKeyFile, dataWiredCacheGb, spec.Spec.IpV6, mongoDbSpec.CustomDataRSParameters, &spec.Spec.TLS)
			log.Debug("Datars container args: " + containerArgs)

			var tolerations []v12.Toleration
			if spec.Spec.Policies != nil {
				tolerations = spec.Spec.Policies.Tolerations
			}

			statefulSet := utils.MongoSSCommonTemplate(
				mongoDbSpec.DockerImage,
				dataOrArbiterResource,
				affinityFunc(s, i),
				nodeSelector,
				request.Namespace,
				pvcName,
				utils.MongoSecret,
				nameKey,
				nameWithIndexes,
				containerArgs,
				spec.Spec.PodSecurityContext,
				tolerations,
				spec.Spec.ImagePullPolicy,
				mongoDbSpec.ContainerTimeoutSeconds,
				mongoDbSpec.ContainerPeriodSeconds,
				spec.Spec.TLS,
				spec.Spec.MongoDB.PriorityClassName)

			err := helperImpl.DeleteStatefulsetAndPods(statefulSet.Name, request.Namespace, spec.Spec.WaitSeconds)

			core.PanicError(err, log.Error, "DATARS statefulset "+statefulSet.Name+" deletion failed")

			err = utils.CreateRuntimeObjectContextWrapper(ctx, statefulSet, statefulSet.ObjectMeta)
			core.PanicError(err, log.Error, "DATARS"+strconv.Itoa(i)+" statefulset failed")

			log.Debug(fmt.Sprintf("Statefulset %s has been created", statefulSet.Name))

			helperImpl := ctx.Get(utils.KubernetesHelperImpl).(core.KubernetesHelper)
			err = helperImpl.WaitForPodsReady(
				map[string]string{
					utils.MongoNode: nameWithIndexes,
				},
				request.Namespace,
				1,
				spec.Spec.WaitSeconds)

			core.PanicError(err, log.Error, "Pods waiting failed")
		}

	}

	return nil
}

type InitDataStep struct {
	core.DefaultExecutable
}

func (r *InitDataStep) Execute(ctx core.ExecutionContext) error {
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)

	log.Info("Data Shards Initialization step started")

	shardCount := spec.Spec.SchemaSettings.ShardCount

	arbFunc := ctx.Get(utils.ArbiterIndexSelectorFunc).(func(int) int)
	arbiterNodeIndex := arbFunc(ctx.Get(utils.MaxReplicaSize).(int))
	domain := "cluster.local"
	if spec.Spec.SchemaSettings.SchemaType == v1alpha1.DR {
		domain = spec.Spec.SchemaSettings.ThisDomainName
	}

	for s := 0; s < shardCount; s++ {
		nameKey := fmt.Sprintf(utils.DataNameKey, s+1)
		err := mongoImpl.MongoReplicaSetInit(
			nameKey,
			nameKey+"%v",
			spec.Spec.SchemaSettings.DataReplicaSize,
			request.Namespace,
			domain,
			false,
			func(i int) string {
				result := ""
				if i != arbiterNodeIndex {
					if i == spec.Spec.SchemaSettings.DataReplicaSize-1 {
						result += ", priority:2"
					} else {
						result += ", priority:1"
					}
				} else {
					result += ", arbiterOnly:true"
				}

				return result
			})

		if err != nil {
			return err
		}

		log.Info(fmt.Sprintf("%s replicaset has been initialized", nameKey))
	}

	//TODO need to check that PRIMARY was elected
	return nil
}

func (r *InitDataStep) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	return core.GetCurrentDeployType(ctx) == core.CleanDeploy && spec.Spec.DisasterRecovery.Mode == utils.ActiveMode, nil
}

type DataStepBuilder struct {
	core.ExecutableBuilder
}

func (r *DataStepBuilder) Build(ctx core.ExecutionContext) core.Executable {
	step := &core.DefaultCompound{}

	step.AddStep(&CreateDataStep{})
	step.AddStep(&InitDataStep{}) //DR

	return step
}
