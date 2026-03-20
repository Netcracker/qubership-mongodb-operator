package dbaas

import (
	"context"

	driver "github.com/Netcracker/qubership-mongodb-driver"
	"github.com/Netcracker/qubership-mongodb-supplementary/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-supplementary/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type UpdateDbaasUserStep struct {
	core.DefaultExecutable
}

func (u UpdateDbaasUserStep) Execute(ctx core.ExecutionContext) error {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbSupplService)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	client := ctx.Get(constants.ContextClient).(client.Client)
	creds, rErr := core.ReadSecret(client, spec.Spec.Dbaas.DbaasAdminSecretName, request.Namespace)
	core.PanicError(rErr, log.Error, "Dbaas credentials secret reading failed")
	mongoService := ctx.Get(utils.ContextMongoService).(driver.MongoService)

	_, err := mongoService.GrantRoleToUser(context.Background(), "admin", string(creds.Data[utils.Username]), "readAnyDatabase")
	return err
}

func (u UpdateDbaasUserStep) Condition(ctx core.ExecutionContext) (bool, error) {
	return core.GetCurrentDeployType(ctx) == core.Update, nil
}
