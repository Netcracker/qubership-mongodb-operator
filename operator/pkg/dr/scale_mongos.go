package dr

import (
	"fmt"

	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
)

type ScaleMongosStep struct {
	core.DefaultExecutable
	Replicas int
}

func (s *ScaleMongosStep) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	return spec.Spec.SchemaSettings.Sharded, nil
}

func (s *ScaleMongosStep) Execute(ctx core.ExecutionContext) error {
	mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)

	err := mongoImpl.ScaleMongos(s.Replicas)

	if err != nil {
		panic(&core.DRExecutionError{Msg: fmt.Sprintf("Failed to scale mongos, err: %s", err.Error())})
	}

	return nil
}
