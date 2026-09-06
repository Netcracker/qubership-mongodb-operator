package dr

import (
	"fmt"

	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
)

// UpdateConfigsvrConnectionStringStep updates configsvrConnectionString in every datars shard without
// dropping the local database. It is safe to run repeatedly and is idempotent.
// It runs before pod recreation (in DataStepBuilder) so that when pods restart they already
// carry the correct cnfrs hostnames — critical when TLS is enabled on an existing DR active cluster.
type UpdateConfigsvrConnectionStringStep struct {
	core.DefaultExecutable
}

func (u *UpdateConfigsvrConnectionStringStep) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	return spec.Spec.SchemaSettings.SchemaType == v1alpha1.DR &&
		spec.Spec.DisasterRecovery.Mode == utils.ActiveMode &&
		spec.Spec.SchemaSettings.Sharded, nil
}

func (u *UpdateConfigsvrConnectionStringStep) Execute(ctx core.ExecutionContext) error {
	mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment).Spec

	err := mongoImpl.UpdateConfigRSConnectionString(
		spec.SchemaSettings.ThisDomainName,
		spec.SchemaSettings.OtherDomainName,
		spec.SchemaSettings.CnfReplicaSize,
		spec.SchemaSettings.ShardCount,
	)
	if err != nil {
		return fmt.Errorf("failed to update configsvrConnectionString in datars: %w", err)
	}
	return nil
}
