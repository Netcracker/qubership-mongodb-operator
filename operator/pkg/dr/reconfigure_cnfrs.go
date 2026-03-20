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

type ReconfigureCnfrsStep struct {
	core.DefaultExecutable
}

func (r *ReconfigureCnfrsStep) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	return !(spec.Spec.DisasterRecovery.Mode == utils.StandbyMode && core.GetCurrentDeployType(ctx) == core.CleanDeploy) && spec.Spec.SchemaSettings.Sharded, nil
}

func (r *ReconfigureCnfrsStep) Execute(ctx core.ExecutionContext) error {
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)

	log.Info("Cnfrs Reconfiguration step started")

	//we pass hostnames of replicas that must become hidden
	var domain string
	if spec.Spec.DisasterRecovery.Mode == utils.ActiveMode {
		domain = spec.Spec.SchemaSettings.OtherDomainName
	} else {
		domain = spec.Spec.SchemaSettings.ThisDomainName
	}

	rs := utils.GetCNFReplicaSetHostNames(spec.Spec.SchemaSettings.CnfReplicaSize, domain, request.Namespace)

	// we need to panic explicitly to catch dr error in right section
	err := mongoImpl.ReconfigureRS(map[string]string{utils.Microservice: utils.CnfNameKey}, rs)
	if err != nil {
		panic(&core.DRExecutionError{Msg: fmt.Sprintf("Failed to reconfigure CNFRS, err: %s", err.Error())})
	}

	return nil
}
