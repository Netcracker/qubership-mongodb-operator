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

const (
	defaultMonitoringUser = "monitoring"
	defaultMonitoringPass = "monitoring"
	// Roles required by mongodb-exporter
	monitoringRoles = `{role: 'clusterMonitor', db: 'admin'}, {role: 'read', db: 'local'}`
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

	user := defaultMonitoringUser
	pass := defaultMonitoringPass

	creds, err := utils.ReadSecret(ctx, utils.MonitoringSecretName, request.Namespace)
	if err != nil {
		log.Info(fmt.Sprintf("Secret %s not found, using default monitoring credentials: %v", utils.MonitoringSecretName, err))
	} else if creds != nil {
		if u, ok := creds.Data[utils.Username]; ok && len(u) > 0 {
			user = string(u)
		}
		if p, ok := creds.Data[utils.Password]; ok && len(p) > 0 {
			pass = string(p)
		}
	}

	log.Info(fmt.Sprintf("Monitoring user %q bootstrap started", user))

	return mongoImpl.CreateUser(
		spec.Spec.AuthDb,
		user,
		pass,
		monitoringRoles,
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
