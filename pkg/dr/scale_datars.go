package dr

import (
	"fmt"

	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
)

type ScaleDATARSStep struct {
	core.DefaultExecutable
	Replicas int
}

func (s *ScaleDATARSStep) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	return spec.Spec.SchemaSettings.Sharded, nil
}

func (s *ScaleDATARSStep) Execute(ctx core.ExecutionContext) error {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)

	err := mongoImpl.ScaleDATARS(spec.Spec.SchemaSettings.ShardCount, s.Replicas)

	if err != nil {
		panic(&core.DRExecutionError{Msg: fmt.Sprintf("Failed to scale datars, err: %s", err.Error())})
	}

	return nil
}
