package dr

import (
	"fmt"
	"time"

	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const defaultMaxLagSeconds = 30

type WaitReplicationLagStep struct {
	core.DefaultExecutable
	MaxLagSeconds int
}

// Condition runs only on HA→DR switch (Active mode, not a clean install).
// Uses ctx.Get(utils.MongoDBDeploymentType) directly — same pattern as WaitExpectedClusterStatusStep —
// because the DR builder runs after MicroServiceCompound restores the previous deploy type in context.
func (w *WaitReplicationLagStep) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	if spec.Spec.DisasterRecovery.Mode != utils.ActiveMode {
		return false, nil
	}
	deployType := ctx.Get(utils.MongoDBDeploymentType)
	isUpdate := deployType != nil && deployType.(core.MicroServiceDeployType) != core.CleanDeploy
	return isUpdate, nil
}

func (w *WaitReplicationLagStep) Execute(ctx core.ExecutionContext) error {
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)

	maxLag := w.MaxLagSeconds
	if maxLag == 0 {
		maxLag = defaultMaxLagSeconds
	}

	timeout := time.Duration(spec.Spec.WaitSeconds) * time.Second
	otherDomain := spec.Spec.SchemaSettings.OtherDomainName

	if spec.Spec.SchemaSettings.Sharded {
		cnfMembers := utils.GetCNFReplicaSetHostNames(
			spec.Spec.SchemaSettings.CnfReplicaSize, otherDomain, request.Namespace)
		labels := map[string]string{utils.Microservice: utils.CnfNameKey}

		if err := pollLag(mongoImpl, labels, cnfMembers, maxLag, timeout, log, utils.CnfNameKey); err != nil {
			return err
		}
	}

	for i := 0; i < spec.Spec.SchemaSettings.ShardCount; i++ {
		dataMembers := utils.GetDATAReplicaSetHostName(
			spec.Spec.SchemaSettings.DataReplicaSize, i, otherDomain, request.Namespace)
		serviceName := fmt.Sprintf(utils.DataNameKey, i+1)
		labels := map[string]string{utils.Microservice: serviceName}

		if err := pollLag(mongoImpl, labels, dataMembers, maxLag, timeout, log, serviceName); err != nil {
			return err
		}
	}

	log.Info("All DR-site replicas caught up within replication lag threshold")
	return nil
}

func pollLag(mongoImpl utils.MongoHelper, labels map[string]string, members []string, maxLag int, timeout time.Duration, log *zap.Logger, rsName string) error {
	return wait.Poll(5*time.Second, timeout, func() (bool, error) {
		ok, err := mongoImpl.CheckReplicationLag(labels, members, maxLag)
		if err != nil {
			log.Warn(fmt.Sprintf("Failed to check replication lag for %s: %v", rsName, err))
			return false, nil
		}
		if !ok {
			log.Debug(fmt.Sprintf("Replication lag for %s still above threshold, waiting", rsName))
		}
		return ok, nil
	})
}
