package mongodb

import (
	"fmt"

	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type UpdateDataOplogStep struct {
	core.DefaultExecutable
	desiredMB   int64
	needsResize bool
	oplogReport *utils.OplogSizeReport
}

func (u *UpdateDataOplogStep) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)
	status, err := mongoImpl.GetClusterStatus(spec.Spec.DisasterRecovery.Mode, spec.Spec.SchemaSettings.ThisDomainName,
		spec.Spec.SchemaSettings.CnfReplicaSize, spec.Spec.SchemaSettings.DataReplicaSize, spec.Spec.SchemaSettings.ShardCount, spec.Spec.SchemaSettings.Sharded)

	if err != nil {
		return false, err
	}

	return core.GetCurrentDeployType(ctx) == core.Update && spec.Spec.MongoDB.DataOpLogSizeMb != "" && u.needsResize && status == utils.Up, nil
}

func (u *UpdateDataOplogStep) Validate(ctx core.ExecutionContext) error {
	var err error
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)
	shardCount := spec.Spec.SchemaSettings.ShardCount
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)

	creds, rErr := utils.ReadSecret(ctx, spec.Spec.MongoDB.MongoRootSecretName, request.Namespace)
	core.PanicError(rErr, log.Error, "MongoDB Root user credentials secret reading failed")

	u.desiredMB, err = utils.ParseOplogSizeMB(spec.Spec.MongoDB.DataOpLogSizeMb)
	if err != nil {
		return err
	}

	oplogReport, err := mongoImpl.GetOplogSizes(ctx, shardCount, creds)
	if err != nil {
		return err
	}

	needsResize := false
	for _, replicaSetInfo := range oplogReport.Items {
		currentSizeMb := replicaSetInfo.MaxSizeMB
		if currentSizeMb < u.desiredMB {
			needsResize = true
		}
	}

	u.needsResize = needsResize
	u.oplogReport = oplogReport
	return nil
}

func (u *UpdateDataOplogStep) Execute(ctx core.ExecutionContext) error {
	mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	err := mongoImpl.UpdateOplogSize(ctx, u.desiredMB, *u.oplogReport)
	if err != nil {
		log.Debug(fmt.Sprintf("Update oplog error [%s]", err))
		return err
	}
	return nil
}
