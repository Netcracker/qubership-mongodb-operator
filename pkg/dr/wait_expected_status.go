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
)

type WaitExpectedClusterStatusStep struct {
	core.DefaultExecutable
	Status string
}

func (u *WaitExpectedClusterStatusStep) Condition(ctx core.ExecutionContext) (bool, error) {
	return ctx.Get(utils.MongoDBDeploymentType) == nil || ctx.Get(utils.MongoDBDeploymentType).(core.MicroServiceDeployType) != core.CleanDeploy, nil
}

func (u *WaitExpectedClusterStatusStep) Execute(ctx core.ExecutionContext) error {
	mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment).Spec
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)

	err := wait.Poll(2*time.Second, time.Duration(spec.WaitSeconds)*time.Second, func() (done bool, err error) {
		status, err := mongoImpl.GetClusterStatus(spec.DisasterRecovery.Mode, spec.SchemaSettings.ThisDomainName, spec.SchemaSettings.CnfReplicaSize,
			spec.SchemaSettings.DataReplicaSize, spec.SchemaSettings.ShardCount, spec.SchemaSettings.Sharded)

		if err != nil {
			log.Warn(fmt.Sprintf("Failed to get cluster status, err is %v", err))
		}
		return status == u.Status, nil
	})

	log.Debug("Cluster is up")

	return err
}
