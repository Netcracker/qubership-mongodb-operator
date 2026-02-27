package dr

import (
	"fmt"

	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
)

type UpdateConfigRSInDATARSStep struct {
	core.DefaultExecutable
}

func (u *UpdateConfigRSInDATARSStep) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	return spec.Spec.DisasterRecovery.Mode == utils.ActiveMode && spec.Spec.SchemaSettings.Sharded, nil
}

func (u *UpdateConfigRSInDATARSStep) Execute(ctx core.ExecutionContext) error {
	mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment).Spec

	err := mongoImpl.UpdateConfigRSInDATARS(spec.SchemaSettings.ThisDomainName, spec.SchemaSettings.OtherDomainName, spec.SchemaSettings.CnfReplicaSize,
		spec.SchemaSettings.DataReplicaSize, spec.SchemaSettings.ShardCount)

	if err != nil {
		panic(&core.DRExecutionError{Msg: fmt.Sprintf("Failed to update config string in datars, err: %s", err.Error())})
	}

	return nil
}
