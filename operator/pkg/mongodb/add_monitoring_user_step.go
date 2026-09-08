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

type AddMonitoringUserStep struct {
	core.DefaultExecutable
}

func (r *AddMonitoringUserStep) Execute(ctx core.ExecutionContext) error {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)

	schema := spec.Spec.SchemaSettings.SchemaType
	sharded := spec.Spec.SchemaSettings.Sharded

	creds, err := utils.ReadSecret(ctx, utils.MonitoringSecretName, request.Namespace)
	if err != nil || creds == nil {
		return fmt.Errorf("secret %s not found, monitoring user bootstrap failed: %w", utils.MonitoringSecretName, err)
	}

	user := string(creds.Data[utils.Username])
	pass := string(creds.Data[utils.Password])
	role := string(creds.Data[utils.Role])

	if user == "" || pass == "" {
		return fmt.Errorf("secret %s missing username or password", utils.MonitoringSecretName)
	}

	log.Info(fmt.Sprintf("Monitoring user %q bootstrap started", user))

	return mongoImpl.CreateUser(
		spec.Spec.AuthDb,
		user,
		pass,
		role,
		false,
		sharded,
		schema != v1alpha1.Single,
		spec.Spec.SchemaSettings.ShardCount,
	)
}

func (r *AddMonitoringUserStep) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	return spec.Spec.DisasterRecovery.Mode == utils.ActiveMode, nil
}
