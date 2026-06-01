package common

import (
	"context"
	"fmt"

	driver "github.com/Netcracker/qubership-mongodb-driver"
	"github.com/Netcracker/qubership-mongodb-supplementary/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-supplementary/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"go.uber.org/zap"
)

type AddServicesUsers struct {
	core.DefaultExecutable
}

func (r *AddServicesUsers) Execute(ctx core.ExecutionContext) error {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbSupplService)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	mongoService := ctx.Get(utils.ContextMongoService).(driver.MongoService)
	sharded := spec.Spec.SchemaSettings.Sharded

	log.Info("Adding services users step is started")

	iroles := ctx.Get(utils.ServicesRolesContextList)

	if iroles != nil {
		log.Info("Adding services roles...")
		roles := iroles.(map[string]utils.RoleToAdd)

		for _, role := range roles {
			// err := mongoImpl.CreateRole(spec.Spec.AuthDb, role.Role, role.Privileges, role.Roles, sharded, notSingle, datarsSize)
			err := mongoService.CreateOrUpdateRole(context.Background(), "admin", role.Role, role.Roles, role.Privileges, sharded)
			if err != nil {
				log.Error(fmt.Sprintf("Failed to add role %s, error is %v", role.Role, err))
				return err
			}
		}
	}

	iusers := ctx.Get(utils.ServicesUsersContextList)

	if iusers != nil {
		log.Debug("Adding services users...")

		users := iusers.(map[string]utils.UserToAdd)
		for _, user := range users {
			err := mongoService.CreateOrUpdateUser(context.Background(), user.User, user.Pass(),
				spec.Spec.AuthDb, spec.Spec.AuthDb, user.Role, sharded, false)
			if err != nil {
				log.Warn(fmt.Sprintf("Failed to add user %s, error is %v", user.User, err))
			}
		}

	}

	return nil
}
