package dr

import (
	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
)

type DeleteConfigStatefulsetsStep struct {
	core.DefaultExecutable
}

func (r *DeleteConfigStatefulsetsStep) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	return core.GetCurrentDeployType(ctx) != core.CleanDeploy &&
		spec.Spec.SchemaSettings.Sharded && spec.Spec.DisasterRecovery.Mode != utils.ActiveMode, nil
}

func (r *DeleteConfigStatefulsetsStep) Execute(ctx core.ExecutionContext) error {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)
	replicas := spec.Spec.SchemaSettings.CnfReplicaSize

	return mongoImpl.RemoveConfigRS(replicas)
}
