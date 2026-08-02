package dr

import (
	"fmt"

	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Add replicas from DR site to Main
type AddDATAReplicas struct {
	core.DefaultExecutable
}

func (r *AddDATAReplicas) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	return spec.Spec.DisasterRecovery.Mode == utils.ActiveMode, nil
}

func (r *AddDATAReplicas) Execute(ctx core.ExecutionContext) error {
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)
	shardSize := spec.Spec.SchemaSettings.ShardCount
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)

	for i := 0; i < shardSize; i++ {
		rs := utils.GetDATAReplicaSetHostName(spec.Spec.SchemaSettings.DataReplicaSize, i, spec.Spec.SchemaSettings.OtherDomainName, request.Namespace)
		serviceName := fmt.Sprintf(utils.DataNameKey, i+1)
		err := mongoImpl.AddDRReplicas(map[string]string{utils.Microservice: serviceName}, rs)
		core.PanicError(err, log.Error, "Failed to add DR replica")
	}

	return nil
}
