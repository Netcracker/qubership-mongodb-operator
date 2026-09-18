package dr

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type UpdatePrometheusExporterStep struct {
	core.DefaultExecutable
	ExportMongos bool
}

func (s *UpdatePrometheusExporterStep) Execute(ctx core.ExecutionContext) error {
	helperImpl := ctx.Get(utils.KubernetesHelperImpl).(core.KubernetesHelper)
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)

	err := helperImpl.UpdateDeploymentByLabels(map[string]string{utils.Microservice: utils.MongoPrometheusExporter},
		request.Namespace,
		func(depl *v1.Deployment) {
			healthzPort := 9216
			if !spec.Spec.SchemaSettings.Sharded {
				healthzPort = 9218
			} else if !s.ExportMongos {
				healthzPort = 9217
			}
			container := depl.Spec.Template.Spec.Containers[0]
			envs := container.Env
			for i := 0; i < len(envs); i++ {
				if envs[i].Name == "EXPORT_MONGOS" {
					envs[i].Value = strconv.FormatBool(s.ExportMongos)
					break
				}
			}
			container.LivenessProbe.TCPSocket.Port = intstr.FromInt(healthzPort)
			container.ReadinessProbe.TCPSocket.Port = intstr.FromInt(healthzPort)
			depl.Spec.Template.Spec.Containers[0].Env = envs
		})

	var nfe *core.NotFoundError
	if errors.As(err, &nfe) {
		return nil
	}

	if err != nil {
		panic(&core.DRExecutionError{Msg: fmt.Sprintf("Failed to scale deployment %s, err: %s", utils.MongoPrometheusExporter, err.Error())})
	}

	return nil
}
