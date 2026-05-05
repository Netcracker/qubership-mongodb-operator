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

type CreateCNFServiceStep struct {
	core.DefaultExecutable
}

func (r *CreateCNFServiceStep) Execute(ctx core.ExecutionContext) error {
	client := ctx.Get(constants.ContextClient).(client.Client)
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)

	log.Info("CNF Service Creation step started")

	template := utils.ShardServiceTemplate(utils.CnfNameKey, request.Namespace)

	// Kubernetes api causes "invalid resourceVersion error" on update. So remove it.
	core.DeleteRuntimeObject(client, &v12.Service{
		ObjectMeta: template.ObjectMeta,
	})

	err := utils.CreateRuntimeObjectContextWrapper(ctx, template, template.ObjectMeta)

	if err != nil {
		return err
	}

	log.Info("CNF Service has been created")

	return nil
}

func (r *CreateCNFServiceStep) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	return spec.Spec.SchemaSettings.Sharded, nil
}

type CreateCNFStep struct {
	core.DefaultExecutable
}

func (r *CreateCNFStep) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	return spec.Spec.SchemaSettings.Sharded, nil
}

func (r *CreateCNFStep) Execute(ctx core.ExecutionContext) error {
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	var spec *v1alpha1.MongodbDeployment = ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	mongoDbSpec := &spec.Spec.MongoDB
	helperImpl := ctx.Get(utils.KubernetesHelperImpl).(core.KubernetesHelper)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)

	log.Info("CNF Replicas Creation step started")

	pvcNames := ctx.Get(utils.PvcNames).([]string)
	nodeLabels := ctx.Get(utils.PVNodes).([]map[string]string)

	cnfSize := spec.Spec.SchemaSettings.CnfReplicaSize

	microservice := utils.Microservice
	nameKey := utils.CnfNameKey

	wiredCacheGb := utils.CalcProperMongoWiredCacheSize(mongoDbSpec.CnfResources.Limits.Memory().Value(), mongoDbSpec.CnfWiredTigerCacheGb, 0.25)
	log.Debug(fmt.Sprintf("Wired tiger cache size = %v", wiredCacheGb))

	for i := 0; i < cnfSize; i++ {
		nameWithIndex := fmt.Sprintf(utils.CnfNameWithIndexFormat, i)
		_, pvcName := utils.ElementsDistribution(utils.GetDummyPvcMap(pvcNames), cnfSize, i, 1, 0)
		nodeSelector := core.ConcatMaps(
			mongoDbSpec.AdditionalNodeLabels,
			utils.MongoReplicaNodeSelector(nodeLabels, cnfSize, i, 1, 0))

		containerArgs := utils.MongoReplicaContainerArgs("--configsvr", utils.Data, nameKey, nameWithIndex, utils.MongoSecret,
			utils.MongoSecretKeyFile, wiredCacheGb, spec.Spec.IpV6, mongoDbSpec.CustomDataRSParameters, &spec.Spec.TLS, mongoDbSpec.CnfOpLogSizeMb)

		log.Debug("Cnfrs container args: " + containerArgs)

		var tolerations []v12.Toleration
		if spec.Spec.Policies != nil {
			tolerations = spec.Spec.Policies.Tolerations
		}

		defaultAffinity := &v12.Affinity{
			PodAntiAffinity: &v12.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []v12.PodAffinityTerm{
					{
						LabelSelector: &v13.LabelSelector{
							MatchExpressions: []v13.LabelSelectorRequirement{
								{
									Key:      microservice,
									Operator: "In",
									Values: []string{
										nameKey,
									},
								},
							},
						},
						TopologyKey: utils.KubeHostName,
					},
				},
			},
		}

		// Use user-defined affinity if available; otherwise use default
		var affinity *v12.Affinity
		if spec.Spec.MongoDB.Affinity != nil {
			affinity = spec.Spec.MongoDB.Affinity
		} else {
			affinity = defaultAffinity
		}

		statefulSet := utils.MongoSSCommonTemplate(
			mongoDbSpec.DockerImage,
			*mongoDbSpec.CnfResources,
			*affinity,
			nodeSelector,
			request.Namespace,
			pvcName,
			utils.MongoSecret,
			nameKey,
			nameWithIndex,
			containerArgs,
			spec.Spec.PodSecurityContext,
			tolerations,
			spec.Spec.ImagePullPolicy,
			mongoDbSpec.ContainerTimeoutSeconds,
			mongoDbSpec.ContainerPeriodSeconds,
			spec.Spec.TLS,
			spec.Spec.MongoDB.PriorityClassName)

		err := helperImpl.DeleteStatefulsetAndPods(statefulSet.Name, request.Namespace, spec.Spec.WaitSeconds)

		core.PanicError(err, log.Error, "CNFRS statefulset "+statefulSet.Name+" deletion failed")

		err = utils.CreateRuntimeObjectContextWrapper(ctx, statefulSet, statefulSet.ObjectMeta)
		core.PanicError(err, log.Error, "CNFRS"+strconv.Itoa(i)+" statefulset failed")

		log.Debug(fmt.Sprintf("Statefulset %s has been created", statefulSet.Name))

		err = helperImpl.WaitForPodsReady(
			map[string]string{
				utils.MongoNode: nameWithIndex,
			},
			request.Namespace,
			1,
			spec.Spec.WaitSeconds)
		core.PanicError(err, log.Error, "Pods waiting failed")
	}

	log.Info("CNF Pods are ready")

	return nil
}

type InitCNFStep struct {
	core.DefaultExecutable
}

func (r *InitCNFStep) Execute(ctx core.ExecutionContext) error {
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)

	log.Info("CNF Initialization step started")

	domain := "cluster.local"
	if spec.Spec.SchemaSettings.SchemaType == v1alpha1.DR {
		domain = spec.Spec.SchemaSettings.ThisDomainName
	}

	err := mongoImpl.MongoReplicaSetInit(
		utils.CnfNameKey,
		utils.CnfNameWithIndexFormat,
		spec.Spec.SchemaSettings.CnfReplicaSize,
		request.Namespace,
		domain,
		true,
		func(i int) string {
			return ""
		})

	if err == nil {
		log.Info("CNF Successfully initialized")
	}

	return err
}

func (r *InitCNFStep) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)

	return core.GetCurrentDeployType(ctx) == core.CleanDeploy &&
		spec.Spec.SchemaSettings.Sharded && spec.Spec.DisasterRecovery.Mode == utils.ActiveMode, nil
}

type UpdateCnfOplogStep struct {
	core.DefaultExecutable
	desiredMB   int64
	needsResize bool
	oplogReport *utils.OplogSizeReport
}

func (u *UpdateCnfOplogStep) Condition(ctx core.ExecutionContext) (bool, error) {
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	shardCount := spec.Spec.SchemaSettings.ShardCount

	if core.GetCurrentDeployType(ctx) == core.Update && spec.Spec.MongoDB.CnfOpLogSizeMb != 0 {
		status, err := mongoImpl.GetClusterStatus(spec.Spec.DisasterRecovery.Mode, spec.Spec.SchemaSettings.ThisDomainName,
			spec.Spec.SchemaSettings.CnfReplicaSize, spec.Spec.SchemaSettings.DataReplicaSize, spec.Spec.SchemaSettings.ShardCount, spec.Spec.SchemaSettings.Sharded)
		if err != nil {
			return false, err
		}

		creds, rErr := utils.ReadSecret(ctx, spec.Spec.MongoDB.MongoRootSecretName, request.Namespace)
		core.PanicError(rErr, log.Error, "MongoDB Root user credentials secret reading failed")

		u.desiredMB = spec.Spec.MongoDB.CnfOpLogSizeMb
		oplogReport, err := mongoImpl.GetOplogSizes(utils.CnfNameWithIndexFormat, shardCount, creds, request.Namespace, spec.Spec.SchemaSettings.ThisDomainName, spec.Spec.DockerImage, spec.Spec.AuthDb)
		if err != nil {
			return false, err
		}

		needsResize := false
		for _, replicaSetInfo := range oplogReport.Items {
			currentSizeMb := replicaSetInfo.MaxSizeMB
			if currentSizeMb != u.desiredMB {
				needsResize = true
			}
		}

		u.needsResize = needsResize
		u.oplogReport = oplogReport

		return u.needsResize && status == utils.Up, nil
	}

	return false, nil
}

func (u *UpdateCnfOplogStep) Validate(ctx core.ExecutionContext) error {
	return nil
}

func (u *UpdateCnfOplogStep) Execute(ctx core.ExecutionContext) error {
	mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	err := mongoImpl.UpdateOplogSize(u.desiredMB, *u.oplogReport)
	if err != nil {
		log.Debug(fmt.Sprintf("Update oplog error [%s]", err))
		return err
	}
	return nil
}

type CNFStepBuilder struct {
	core.ExecutableBuilder
}

func (r *CNFStepBuilder) Build(ctx core.ExecutionContext) core.Executable {
	step := &core.DefaultCompound{}

	step.AddStep(&CreateCNFServiceStep{})
	step.AddStep(&CreateCNFStep{})
	step.AddStep(&InitCNFStep{}) //DR
	step.AddStep(&UpdateCnfOplogStep{})
	return step
}
