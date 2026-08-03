package dr

import (
	"fmt"

	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"go.uber.org/zap"
)

// RenameMemberDomainStep renames RS member hostnames to match thisDomainName.
// Handles the case where a cluster was originally deployed with a different domain suffix
// (e.g. cluster.local) and is now being switched to DR with thisDomainName=cluster-1.local.
// Idempotent: JS skips reconfig if all members already use the correct domain.
type RenameMemberDomainStep struct {
	core.DefaultExecutable
}

func (r *RenameMemberDomainStep) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	// Only relevant for DR schema with a configured domain name.
	return spec.Spec.SchemaSettings.SchemaType == v1alpha1.DR &&
		spec.Spec.SchemaSettings.ThisDomainName != "" && spec.Spec.DisasterRecovery.Mode == utils.ActiveMode, nil
}

func (r *RenameMemberDomainStep) Execute(ctx core.ExecutionContext) error {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)

	domain := spec.Spec.SchemaSettings.ThisDomainName

	if spec.Spec.SchemaSettings.Sharded {
		log.Info(fmt.Sprintf("RenameMemberDomain: renaming cnfrs members to domain %s", domain))
		if err := mongoImpl.RenameRSMemberDomain(map[string]string{utils.Microservice: utils.CnfNameKey}, domain); err != nil {
			panic(&core.DRExecutionError{Msg: fmt.Sprintf("Failed to rename cnfrs member domains: %s", err.Error())})
		}
	}

	for i := 0; i < spec.Spec.SchemaSettings.ShardCount; i++ {
		serviceName := fmt.Sprintf(utils.DataNameKey, i+1)
		log.Info(fmt.Sprintf("RenameMemberDomain: renaming %s members to domain %s", serviceName, domain))
		if err := mongoImpl.RenameRSMemberDomain(map[string]string{utils.Microservice: serviceName}, domain); err != nil {
			panic(&core.DRExecutionError{Msg: fmt.Sprintf("Failed to rename %s member domains: %s", serviceName, err.Error())})
		}
	}

	return nil
}
