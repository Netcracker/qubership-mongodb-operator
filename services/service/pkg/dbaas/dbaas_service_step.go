package dbaas

import (
	"os"

	"github.com/Netcracker/qubership-mongodb-supplementary/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-supplementary/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	cUtils "github.com/Netcracker/qubership-nosqldb-operator-core/pkg/utils"
	"go.uber.org/zap"
	v12 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type DbaasService struct {
	core.DefaultExecutable
}

func (r *DbaasService) Execute(ctx core.ExecutionContext) error {
	client := ctx.Get(constants.ContextClient).(client.Client)
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbSupplService)

	log.Info("Dbaas Service creation step started")

	template := cUtils.SimpleServiceTemplate(
		utils.DbaasName,
		map[string]string{
			constants.App:              utils.MongoCluster,
			constants.Microservice:     utils.DbaasName,
			utils.Name:                 utils.DbaasName,
			utils.AppName:              utils.DbaasName,
			utils.AppTechnology:        "go",
			utils.AppComponent:         "backend",
			utils.AppInstance:          os.Getenv("RELEASE_NAME"),
			utils.AppManagedBy:         "operator",
			utils.AppPartOf:            "mongodb-services",
			utils.AppManagedByOperator: "mongodb-services-operator",
		},
		map[string]string{
			utils.Name: utils.DbaasName,
		},
		map[string]int32{"http": cUtils.GetHTTPPort(cUtils.IsTLSEnableForDBAAS(spec.Spec.Dbaas.DbaasAggregatorRegistrationAddress, spec.Spec.TLS.Enabled))},
		request.Namespace)

	// Kubernetes api causes "invalid resourceVersion error" on update. So remove it.
	core.DeleteRuntimeObject(client, &v12.Service{
		ObjectMeta: template.ObjectMeta,
	})

	err := utils.CreateRuntimeObjectContextWrapper(ctx, template, template.ObjectMeta)

	if err != nil {
		return err
	}

	log.Debug("Dbaas Service has been created")

	return nil
}
