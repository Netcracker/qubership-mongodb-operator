package mongodb

import (
	"fmt"

	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"go.uber.org/zap"
)

type CalculateMaxReplicaSizeStep struct {
	core.DefaultExecutable
}

func (r *CalculateMaxReplicaSizeStep) Validate(ctx core.ExecutionContext) error {
	var spec *v1alpha1.MongodbDeployment = ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)

	if spec.Spec.SchemaSettings.SchemaType != v1alpha1.Single {

		if spec.Spec.SchemaSettings.CnfReplicaSize == 0 ||
			spec.Spec.SchemaSettings.DataReplicaSize == 0 ||
			spec.Spec.SchemaSettings.ShardCount == 0 {
			return &core.ExecutionError{Msg: utils.SchemaSettingsValidationZeros}
		}

		if spec.Spec.SchemaSettings.SchemaType != v1alpha1.DR {
			if spec.Spec.SchemaSettings.CnfReplicaSize%2 == 0 ||
				spec.Spec.SchemaSettings.DataReplicaSize%2 == 0 {
				return &core.ExecutionError{Msg: utils.SchemaSettingsValidationHAandArbiter}
			}
		}
	}

	return nil
}

func (r *CalculateMaxReplicaSizeStep) Execute(ctx core.ExecutionContext) error {
	var spec *v1alpha1.MongodbDeployment = ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)

	var res int
	if spec.Spec.SchemaSettings.SchemaType == v1alpha1.Single {
		res = 1
	} else {
		res = core.MaxInt(spec.Spec.SchemaSettings.DataReplicaSize, spec.Spec.SchemaSettings.CnfReplicaSize)
	}

	ctx.Set(utils.MaxReplicaSize, res)
	ctx.Set(utils.MaxPVCCountForService, res)

	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	log.Debug(fmt.Sprintf("Max replica size = %v", res))

	return nil
}
