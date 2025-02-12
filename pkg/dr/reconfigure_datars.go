package dr

import (
	"fmt"

	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ReconfigureDataRSStep struct {
	core.DefaultExecutable
}

func (r *ReconfigureDataRSStep) Execute(ctx core.ExecutionContext) error {
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)

	//Reconfigure datars
	datarsSize := spec.Spec.SchemaSettings.ShardCount

	//we pass hostnames of replicas that must become hidden
	var domain string
	if spec.Spec.DisasterRecovery.Mode == utils.ActiveMode {
		domain = spec.Spec.SchemaSettings.OtherDomainName
	} else {
		domain = spec.Spec.SchemaSettings.ThisDomainName
	}

	for i := 0; i < datarsSize; i++ {
		serviceName := fmt.Sprintf(utils.DataNameKey, i+1)
		rs := utils.GetDATAReplicaSetHostName(spec.Spec.SchemaSettings.DataReplicaSize, i, domain, request.Namespace)
		err := mongoImpl.ReconfigureRS(map[string]string{utils.Microservice: serviceName}, rs)
		if err != nil {
			panic(&core.DRExecutionError{Msg: fmt.Sprintf("Failed to reconfigure DATARS, err: %s", err.Error())})
		}
	}

	return nil
}

func (r *ReconfigureDataRSStep) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	return !(spec.Spec.DisasterRecovery.Mode == utils.StandbyMode && core.GetCurrentDeployType(ctx) == core.CleanDeploy), nil
}
