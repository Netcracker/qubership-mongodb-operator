package mongodb

import (
	"fmt"
	"strings"

	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"go.uber.org/zap"
	v12 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type CreateMongosServiceStep struct {
	core.DefaultExecutable
}

func (r *CreateMongosServiceStep) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	return spec.Spec.SchemaSettings.Sharded, nil
}

func (r *CreateMongosServiceStep) Execute(ctx core.ExecutionContext) error {
	client := ctx.Get(constants.ContextClient).(client.Client)
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)

	log.Info("Mongos Services creation step started")

	templates := []*v12.Service{
		utils.MongosServiceTemplate(utils.Mongos, utils.Mongos, request.Namespace),
		utils.MongosServiceTemplate(utils.MongosPrivate, utils.Mongos, request.Namespace),
	}

	for _, template := range templates {
		// Kubernetes api causes "invalid resourceVersion error" on update. So remove it.
		core.DeleteRuntimeObject(client, &v12.Service{
			ObjectMeta: template.ObjectMeta,
		})

		err := utils.CreateRuntimeObjectContextWrapper(ctx, template, template.ObjectMeta)

		if err != nil {
			return err
		}
	}

	log.Debug("Mongos Services has been created")

	return nil
}

func commonMongosExecute(ctx core.ExecutionContext, template *v12.ReplicationController) error {
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	var spec *v1alpha1.MongodbDeployment = ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)

	helperImpl := ctx.Get(utils.KubernetesHelperImpl).(core.KubernetesHelper)

	err := helperImpl.DeleteRCAndPods(template.Name, request.Namespace, spec.Spec.WaitSeconds)

	core.PanicError(err, log.Error, "Mongos rc "+template.Name+" deletion failed")

	err = utils.CreateRuntimeObjectContextWrapper(ctx, template, template.ObjectMeta)

	if err != nil {
		return &core.ExecutionError{Msg: "Error while creating Mongos rc. Error: " + err.Error()}
	}

	if spec.Spec.DisasterRecovery.Mode == utils.ActiveMode {
		err = helperImpl.WaitForPodsReady(
			template.Labels,
			request.Namespace,
			int(*template.Spec.Replicas),
			spec.Spec.WaitSeconds)

		if err != nil {
			return &core.ExecutionError{Msg: "Error happened while waiting " + utils.Mongos + " pod is ready. Error: " + err.Error()}
		}
	}

	log.Info("Mongos successfully created")

	return nil
}

type CreateHAMongosStep struct {
	core.DefaultExecutable
}

func (r *CreateHAMongosStep) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	return spec.Spec.SchemaSettings.Sharded, nil
}

func (r *CreateHAMongosStep) Execute(ctx core.ExecutionContext) error {
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	var spec *v1alpha1.MongodbDeployment = ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	mongoDbSpec := &spec.Spec.MongoDB
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)

	log.Info("HA Mongos Creation step started")
	keyfileAuth := true
	if spec.Spec.MongoDB.ClusterAuthMode == "x509" {
		spec.Spec.TLS.CertificateSecretName = "mongos-x509-tls"
		keyfileAuth = false
	}

	cnfSize := spec.Spec.SchemaSettings.CnfReplicaSize
	var configNodes []string
	if spec.Spec.SchemaSettings.SchemaType == v1alpha1.DR {
		configNodes = utils.GetDRCNFReplicaSetHostNames(cnfSize, spec.Spec.SchemaSettings.ThisDomainName, spec.Spec.SchemaSettings.OtherDomainName, request.Namespace)
	} else {
		configNodes = utils.GetCNFReplicaSetHostNames(cnfSize, spec.Spec.SchemaSettings.ThisDomainName, request.Namespace)
	}

	log.Debug(fmt.Sprintf("Config nodes for mongos initialization: %s", configNodes))

	args := utils.HAMongosContainerArgs(utils.Tmp, utils.Data, utils.MongoSecret,
		utils.MongoSecretKeyFile, fmt.Sprintf("cnfrs/%s", strings.Join(configNodes, ",")), spec.Spec.IpV6, &spec.Spec.TLS, keyfileAuth)

	log.Debug("Mongos container args: " + args)

	var tolerations []v12.Toleration
	if spec.Spec.Policies != nil {
		tolerations = spec.Spec.Policies.Tolerations
	}

	var numberOfReplicas int
	if spec.Spec.DisasterRecovery.Mode == utils.ActiveMode {
		numberOfReplicas = spec.Spec.SchemaSettings.MongosReplicas
	} else {
		numberOfReplicas = 0
	}

	template := utils.MongosRCTemplate(
		mongoDbSpec.DockerImage,
		*mongoDbSpec.MongosResources,
		request.Namespace,
		args,
		spec.Spec.PodSecurityContext,
		tolerations,
		spec.Spec.ImagePullPolicy,
		numberOfReplicas,
		spec.Spec.TLS,
		spec.Spec.MongoDB.PriorityClassName,
		spec.Spec.MongoDB.Affinity)

	return commonMongosExecute(ctx, template)
}

type ShardsRegistrationStep struct {
	core.DefaultExecutable
}

func (r *ShardsRegistrationStep) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	return core.GetCurrentDeployType(ctx) == core.CleanDeploy &&
		spec.Spec.SchemaSettings.Sharded && spec.Spec.DisasterRecovery.Mode == utils.ActiveMode, nil
}

func (r *ShardsRegistrationStep) Execute(ctx core.ExecutionContext) error {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)

	datarsSize := spec.Spec.SchemaSettings.DataReplicaSize
	shards := spec.Spec.SchemaSettings.ShardCount

	return mongoImpl.AddShards(spec.Spec.SchemaSettings.ThisDomainName, datarsSize, shards)
}

type CreateSingleMongosStep struct {
	core.DefaultExecutable
}

func (r *CreateSingleMongosStep) Execute(ctx core.ExecutionContext) error {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	mongoDbSpec := &spec.Spec.MongoDB
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)

	log.Info("Single (Non-HA) Mongos Creation step started")

	wiredCacheGb := utils.CalcProperMongoWiredCacheSize(mongoDbSpec.MongosResources.Limits.Memory().Value(), mongoDbSpec.SingleWiredTigerCacheGb, 0.25)
	log.Debug(fmt.Sprintf("Wired tiger cache size = %v", wiredCacheGb))

	args := utils.SingleMongosContainerArgs(wiredCacheGb, spec.Spec.IpV6, &spec.Spec.TLS)

	log.Debug("Mongos container args: " + args)

	pvcNames := ctx.Get(utils.PvcNames).([]string)

	nodeLabels := ctx.Get(utils.PVNodes).([]map[string]string)
	nodeSelector := map[string]string{}
	if nodeLabels != nil &&
		len(nodeLabels) > 0 {
		nodeSelector = nodeLabels[0%len(nodeLabels)]
	}

	var tolerations []v12.Toleration
	if spec.Spec.Policies != nil {
		tolerations = spec.Spec.Policies.Tolerations
	}

	template := utils.SingleMongosRCTemplate(
		mongoDbSpec.DockerImage,
		*mongoDbSpec.MongosResources,
		request.Namespace,
		args,
		pvcNames[0],
		nodeSelector,
		spec.Spec.PodSecurityContext,
		tolerations,
		spec.Spec.ImagePullPolicy,
		spec.Spec.TLS,
		spec.Spec.MongoDB.PriorityClassName,
		spec.Spec.Affinity)

	return commonMongosExecute(ctx, template)
}

type HAMongosStepBuilder struct {
	core.ExecutableBuilder
}

func (r *HAMongosStepBuilder) Build(ctx core.ExecutionContext) core.Executable {
	step := &core.DefaultCompound{}

	step.AddStep(&CreateMongosServiceStep{})
	step.AddStep(&CreateHAMongosStep{})
	step.AddStep(&ShardsRegistrationStep{})

	return step
}

type SingleMongosStepBuilder struct {
	core.ExecutableBuilder
}

func (r *SingleMongosStepBuilder) Build(ctx core.ExecutionContext) core.Executable {
	step := &core.DefaultCompound{}

	step.AddStep(&CreateMongosServiceStep{})
	step.AddStep(&CreateSingleMongosStep{})

	return step
}

type SetdefaultWriteConcernStep struct {
	core.DefaultExecutable
}

func (r *SetdefaultWriteConcernStep) Execute(ctx core.ExecutionContext) error {
	mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)
	_, err := mongoImpl.RunOnCnfrs(utils.WriteConcernChangeCommand)
	return err
}

func (r *SetdefaultWriteConcernStep) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	return spec.Spec.SchemaSettings.SchemaType == v1alpha1.Arbiter, nil
}
