package mongodb

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type roleWrapper struct {
	Roles []json.RawMessage `json:"roles"`
}

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
		log.Info(fmt.Sprintf("secret %s not found, monitoring user bootstrap skipped", utils.MonitoringSecretName))
		return nil
	}

	user := string(creds.Data[utils.Username])
	pass := string(creds.Data[utils.Password])
	roleRaw := string(creds.Data[utils.Role])

	if user == "" || pass == "" {
		return fmt.Errorf("secret %s missing username or password", utils.MonitoringSecretName)
	}

	log.Info(fmt.Sprintf("Monitoring user %q bootstrap started", user))

	role, err := flattenRoleDocs(roleRaw)
	if err != nil {
		return fmt.Errorf("secret %s has invalid role format: %w", utils.MonitoringSecretName, err)
	}

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

func flattenRoleDocs(raw string) (string, error) {
	var w roleWrapper
	if err := json.Unmarshal([]byte(raw), &w); err != nil {
		return "", err
	}
	if len(w.Roles) == 0 {
		return "", fmt.Errorf("no roles found in %q", raw)
	}
	parts := make([]string, len(w.Roles))
	for i, r := range w.Roles {
		parts[i] = string(r)
	}
	return strings.Join(parts, ","), nil
}

func (r *AddMonitoringUserStep) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	return spec.Spec.DisasterRecovery.Mode == utils.ActiveMode, nil
}
