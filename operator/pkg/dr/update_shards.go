package dr

import (
	"fmt"

	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
)

type UpdateShardsStep struct {
	core.DefaultExecutable
}

func (u *UpdateShardsStep) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	return spec.Spec.DisasterRecovery.Mode == utils.ActiveMode && spec.Spec.SchemaSettings.Sharded, nil
}

func (u *UpdateShardsStep) Execute(ctx core.ExecutionContext) error {
	mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment).Spec

	err := mongoImpl.UpdateShardsInConfigDB(spec.SchemaSettings.ThisDomainName,
		spec.SchemaSettings.DataReplicaSize, spec.SchemaSettings.ShardCount)

	if err != nil {
		panic(&core.DRExecutionError{Msg: fmt.Sprintf("Failed to update shards in config DB, err: %s", err.Error())})
	}

	return nil
}
