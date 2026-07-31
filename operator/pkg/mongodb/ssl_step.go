package mongodb

import (
	"fmt"

	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type CreateSSLSecretStep struct {
	core.DefaultExecutable
}

func (r *CreateSSLSecretStep) Execute(ctx core.ExecutionContext) error {
	var request reconcile.Request = ctx.Get(constants.ContextRequest).(reconcile.Request)
	helperImpl := ctx.Get(utils.KubernetesHelperImpl).(core.KubernetesHelper)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)

	log.Info("Mongo SSL step started")
	var sslString string

	if spec.Spec.SchemaSettings.SchemaType == v1alpha1.DR {

		if core.GetCurrentDeployType(ctx) != core.CleanDeploy {
			// Rolling update on an already-deployed DR-enabled cluster (e.g.
			// Running -> Disabled, or Disabled -> Primary). The keyfile secret
			// already exists and MUST NOT change — internal auth requires it
			// stay byte-identical across all already-running members. Fetch it
			// from our own operator rather than generating a new one or
			// contacting the other cluster, which may not be reachable yet.
			url := fmt.Sprintf("http://%s.%s.svc.%s:8069", utils.OperatorServiceName,
				request.Namespace, spec.Spec.SchemaSettings.ThisDomainName)

			client := utils.NewOperatorClinet(url)
			result, err := client.GetKeyFile()
			core.PanicError(err, log.Error, "Failed to receive existing keyfile from own operator during rolling update")
			sslString = result

		} else {
			url := fmt.Sprintf("http://%s.%s.svc.%s:8069", utils.OperatorServiceName,
				request.Namespace, spec.Spec.SchemaSettings.OtherDomainName)

			client := utils.NewOperatorClinet(url)
			result, err := client.GetKeyFile()
			if err == nil {
				sslString = result
			} else if spec.Spec.DisasterRecovery.Mode == utils.StandbyMode {
				sslResult, sslErr := helperImpl.OpensslCommand([]string{"rand", "-base64", "755"})
				core.PanicError(sslErr, log.Error, "Can't create mongo-secret")
				sslString = string(sslResult)
			} else {
				core.PanicError(err, log.Error, fmt.Sprintf("Failed to recieve keyfile from %s. Make sure that other operator is running ", url))
			}
		}

	} else {
		result, err := helperImpl.OpensslCommand([]string{"rand", "-base64", "755"})
		core.PanicError(err, log.Error, "Can't create mongo-secret")
		sslString = string(result)
	}

	secretTemplate := utils.SecretTemplate(
		utils.MongoSecret,
		map[string]string{
			utils.MongoSecretKeyFile: string(sslString),
		},
		request.Namespace)

	err := utils.CreateRuntimeObjectContextWrapper(ctx, secretTemplate, secretTemplate.ObjectMeta)
	core.PanicError(err, log.Error, "Can't create mongo-secret")

	log.Info("Mongo SSL has been successfully created")
	return nil
}
