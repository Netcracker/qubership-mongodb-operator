package prometheus_exporter

import (
	"github.com/Netcracker/qubership-mongodb-supplementary/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	cUtils "github.com/Netcracker/qubership-nosqldb-operator-core/pkg/utils"
	"go.uber.org/zap"
	v12 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type PrometheusExporterService struct {
	core.DefaultExecutable
}

func (r *PrometheusExporterService) Execute(ctx core.ExecutionContext) error {
	client := ctx.Get(constants.ContextClient).(client.Client)
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)

	log.Info("PrometheusExporterService Service creation step started")

	ports := []v12.ServicePort{
		{
			Name: "mongos",
			Port: 9216,
		},
		{
			Name: "cnfrs",
			Port: 9217,
		},
		{
			Name: "datars1",
			Port: 9218,
		},
		{
			Name: "datars2",
			Port: 9219,
		},
		{
			Name: "datars3",
			Port: 9220,
		},
	}

	labels := map[string]string{
		utils.App:                               utils.MongoCluster,
		utils.Microservice:                      utils.MongoPrometheusExporter,
		"app.kubernetes.io/part-of":             "mongodb-services",
		"name":                                  utils.MongoPrometheusExporter,
		"app.kubernetes.io/name":                utils.MongoPrometheusExporter,
		"app.kubernetes.io/managed-by":          "operator",
		"app.kubernetes.io/managed-by-operator": "mongodb-services-operator",
	}

	selectors := map[string]string{
		utils.Name: utils.MongoPrometheusExporter,
	}

	template := cUtils.MultiportServiceTemplate(utils.MongoPrometheusExporter, labels, selectors, &ports, request.Namespace)

	// Kubernetes api causes "invalid resourceVersion error" on update. So remove it.
	core.DeleteRuntimeObject(client, &v12.Service{
		ObjectMeta: template.ObjectMeta,
	})

	err := utils.CreateRuntimeObjectContextWrapper(ctx, template, template.ObjectMeta)

	if err != nil {
		return err
	}

	log.Debug("PrometheusExporter Service has been created")

	return nil
}
