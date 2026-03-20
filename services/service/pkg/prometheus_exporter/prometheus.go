package prometheus_exporter

import (
	"github.com/Netcracker/qubership-mongodb-supplementary/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-supplementary/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/steps"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type PrometheusExporterCompound struct {
	core.DefaultCompound
}

type PrometheusExporterBuilder struct {
	core.ExecutableBuilder
}

func (r *PrometheusExporterBuilder) Build(ctx core.ExecutionContext) core.Executable {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbSupplService)
	client := ctx.Get(constants.ContextClient).(client.Client)
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)

	creds, rErr := core.ReadSecret(client, spec.Spec.PrometheusExporter.MonitoringSecretName, request.Namespace)
	core.PanicError(rErr, log.Error, "Prometheus credentials secret reading failed")

	exporter := PrometheusExporterCompound{}

	exporter.AddStep(&PrometheusExporterService{})

	users := []utils.UserToAdd{
		{
			User:       string(creds.Data[utils.Username]),
			Pass:       func() string { return string(creds.Data[utils.Password]) },
			Role:       string(creds.Data[utils.Role]),
			ShardLocal: true,
			AddToVault: false,
		},
	}

	if spec.Spec.VaultRegistration.Enabled {
		step := &steps.MoveSecretToVault{
			SecretName:            spec.Spec.PrometheusExporter.MonitoringSecretName,
			VaultRegistration:     &spec.Spec.VaultRegistration,
			CtxVarToStorePassword: utils.MonitoringContextVar,
		}
		if ok, _ := step.Condition(ctx); ok {
			step.Execute(ctx)
		}

		users[0].Pass = func() string { return ctx.Get(utils.MonitoringContextVar).(string) }
	}

	for _, user := range users {
		utils.AddServicesUsersToContext(
			ctx,
			user,
		)
	}

	exporter.AddStep(&PrometheusExporterDeployment{})

	return &exporter
}

func (r *PrometheusExporterCompound) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbSupplService)
	microServiceCheck, microserviceCheckErr := core.CheckSpecChange(ctx, spec.Spec.PrometheusExporter, utils.MongoPrometheusExporter)
	commonCheck := ctx.Get(utils.IsAnyCommonParameterChanged).(bool)

	if microserviceCheckErr != nil {
		return microServiceCheck, microserviceCheckErr
	} else {
		return microServiceCheck || commonCheck, nil
	}
}
