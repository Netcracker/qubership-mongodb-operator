package mongodb

import (
	"strconv"
	"strings"

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

	var needsAnyResize bool
	cmd := `db.getSiblingDB("local").oplog.rs.stats().maxSize`
	sizeOutputs, err := mongoImpl.RunOnShards(cmd, shardCount)
	if err != nil {
		return err
	}

	log.Info("Size outputs are like this : ===========")
	log.Sugar().Infof("%s", sizeOutputs)

	log.Info("Size outputs are like this : ===========")

	return nil

	for _, size := range sizeOutputs {
		currentBytes, err := strconv.ParseInt(strings.TrimSpace(size), 10, 64)
		if err != nil {
			//return fmt.Errorf("failed to parse oplog size for shard %s: %v", dKey, err)
		}

		currentMB := currentBytes / (1024 * 1024)

		if currentMB < u.desiredMB {
			needsAnyResize = true
		}
	}

	// 4. Skip Execute if nothing to do
	if !needsAnyResize {

	}

	return nil
}
