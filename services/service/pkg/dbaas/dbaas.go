package dbaas

import (
	"github.com/Netcracker/qubership-mongodb-supplementary/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-supplementary/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type DbaasCompound struct {
	core.DefaultCompound
}

type DbaasBuilder struct {
	core.ExecutableBuilder
}

func (r *DbaasBuilder) Build(ctx core.ExecutionContext) core.Executable {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbSupplService)
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	client := ctx.Get(constants.ContextClient).(client.Client)

	creds, rErr := core.ReadSecret(client, spec.Spec.Dbaas.DbaasAdminSecretName, request.Namespace)
	core.PanicError(rErr, log.Error, "Dbaas credentials secret reading failed")

	dbaasAdminRole, dbARErr := core.ReadSecret(client, spec.Spec.Dbaas.DbaasAdminRoleSecretName, request.Namespace)
	core.PanicError(dbARErr, log.Error, "Dbaas Admin Role secret reading failed")

	userToAdd := utils.UserToAdd{
		User:       string(creds.Data[utils.Username]),
		Role:       string(creds.Data[utils.Role]),
		ShardLocal: true,
	}
	userToAdd.Pass = func() string { return string(creds.Data[utils.Password]) }

	users := []utils.UserToAdd{userToAdd}

	for _, user := range users {
		utils.AddServicesUsersToContext(
			ctx,
			user,
		)
	}

	utils.AddServicesRolesToContext(ctx, utils.RoleToAdd{
		Role:       string(dbaasAdminRole.Data["name"]),
		Privileges: string(dbaasAdminRole.Data["privileges"]),
		Roles:      string(dbaasAdminRole.Data["roles"]),
		ShardLocal: true,
	})

	dbaas := DbaasCompound{}

	dbaas.AddStep(&DbaasService{})

	// dbaas.AddStep(&DbaasConfigMaps{})
	dbaas.AddStep(&DbaasDeployment{})

	return &dbaas
}

func (r *DbaasCompound) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbSupplService)
	microServiceCheck, microserviceCheckErr := core.CheckSpecChange(ctx, spec.Spec.Dbaas, utils.DbaasName)
	commonCheck := ctx.Get(utils.IsAnyCommonParameterChanged).(bool)

	if microserviceCheckErr != nil {
		return microServiceCheck, microserviceCheckErr
	} else {
		return microServiceCheck || commonCheck, nil
	}
}
