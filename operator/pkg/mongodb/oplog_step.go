package mongodb

import (
	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"go.uber.org/zap"
)

type UpdateDataOplogStep struct {
	core.DefaultExecutable
	desiredMB int64
}

func (u *UpdateDataOplogStep) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	log.Info("======== RUNNING CONDITION ===========")
	return core.GetCurrentDeployType(ctx) == core.Update && spec.Spec.MongoDB.DataOpLogSizeMb != "", nil
}

func (u *UpdateDataOplogStep) Validate(ctx core.ExecutionContext) error {
	var err error
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)
	shardCount := spec.Spec.SchemaSettings.ShardCount
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	log.Info("======== RUNNING VALIDATE ===========")

	u.desiredMB, err = utils.ParseOplogSizeMB(spec.Spec.MongoDB.DataOpLogSizeMb)
	if err != nil {
		return err
	}

	oplogReport, err := mongoImpl.GetOplogSizes(shardCount)
	if err != nil {
		log.Sugar().Infof("error for oplog is : %s", err)
		return err
	}

	log.Sugar().Infof("Oplog Report : %v ", oplogReport)
	log.Sugar().Infof("Desire MB : %v", u.desiredMB)

	return nil
}
