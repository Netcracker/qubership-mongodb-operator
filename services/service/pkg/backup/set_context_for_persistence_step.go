package backup

import (
	"github.com/Netcracker/qubership-mongodb-supplementary/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-supplementary/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
)

type PrepareContextForBackupService struct {
	core.DefaultExecutable
}

func (r *PrepareContextForBackupService) Execute(ctx core.ExecutionContext) error {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbSupplService)

	ctx.Set(utils.MaxPVCCountForService, 1)

	if spec.Spec.SchemaSettings.SchemaType == v1alpha1.Single {
		ctx.Set(utils.BackupConfigNodes, 0)
	} else {
		ctx.Set(utils.BackupConfigNodes, spec.Spec.SchemaSettings.CnfReplicaSize)
	}

	return nil
}
