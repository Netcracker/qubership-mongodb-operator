package dr

import (
	"errors"
	"fmt"

	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ScaleBackupDaemonStep struct {
	core.DefaultExecutable
	Replicas int
}

func (s *ScaleBackupDaemonStep) Execute(ctx core.ExecutionContext) error {
	helperImpl := ctx.Get(utils.KubernetesHelperImpl).(core.KubernetesHelper)
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	err := helperImpl.ScaleDeploymentByLabels(map[string]string{utils.Microservice: utils.BackupDaemon}, request.Namespace, s.Replicas, spec.Spec.WaitSeconds)

	var nfe *core.NotFoundError
	if errors.As(err, &nfe) {
		return nil
	}

	if err != nil {
		panic(&core.DRExecutionError{Msg: fmt.Sprintf("Failed to scale backup daemon, err: %s", err.Error())})
	}

	return nil
}
