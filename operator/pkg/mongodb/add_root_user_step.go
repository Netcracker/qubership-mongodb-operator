package mongodb

import (
	"fmt"
	"net/url"

	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type AddUserStep struct {
	core.DefaultExecutable
	Username        string
	Password        string
	Role            string
	Sharded         bool
	customCondition func(ctx core.ExecutionContext) (bool, error)
}

func (r *AddUserStep) Execute(ctx core.ExecutionContext) error {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	schema := spec.Spec.SchemaSettings.SchemaType
	sharded := spec.Spec.SchemaSettings.Sharded

	log.Info(fmt.Sprintf("%s User Creation step is started", r.Username))

	return mongoImpl.CreateUser(spec.Spec.AuthDb, r.Username, r.Password,
		r.Role, core.GetCurrentDeployType(ctx) == core.CleanDeploy, sharded, schema != v1alpha1.Single, spec.Spec.SchemaSettings.ShardCount)

}

func (r *AddUserStep) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	if r.customCondition != nil {
		return r.customCondition(ctx)
	}
	return spec.Spec.DisasterRecovery.Mode == utils.ActiveMode, nil
}

type UpdateContextAuthMongo struct {
	core.DefaultExecutable
	User     string
	Password string
}

func (r *UpdateContextAuthMongo) Execute(ctx core.ExecutionContext) error {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)

	var cmd string
	if r.User != "" && r.Password != "" {
		cmd = fmt.Sprintf(
			utils.MongoCMDAuthTemplate,
			utils.MongoBinary(spec.Spec.MongoDB.DockerImage),
			url.QueryEscape(r.User),
			url.QueryEscape(r.Password),
			spec.Spec.AuthDb,
		)

	} else {
		request := ctx.Get(constants.ContextRequest).(reconcile.Request)
		creds, rErr := utils.ReadSecret(ctx, spec.Spec.MongoDB.MongoRootSecretName, request.Namespace)
		core.PanicError(rErr, log.Error, "MongoDB Root user credentials secret reading failed")

		username := string(creds.Data[utils.Username])
		password := string(creds.Data[utils.Password])

		cmd = fmt.Sprintf(
			utils.MongoCMDAuthTemplate,
			utils.MongoBinary(spec.Spec.MongoDB.DockerImage),
			url.QueryEscape(username),
			url.QueryEscape(password),
			spec.Spec.AuthDb,
		)
	}

	if spec.Spec.TLS.Enabled {
		cmd = fmt.Sprint(cmd, fmt.Sprintf(" --tls --tlsCAFile %s%s --tlsAllowInvalidCertificates", utils.RootCertPath, spec.Spec.TLS.RootCAFileName))
	}
	mongoImpl.SetMongoCMD(cmd)

	return nil
}

type ResetNonVaultPassword struct {
	core.DefaultExecutable
}

func (r *ResetNonVaultPassword) Execute(ctx core.ExecutionContext) error {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	secret, err := utils.ReadSecret(ctx, spec.Spec.MongoDB.MongoRootSecretName, request.Namespace)
	if err != nil {
		return err
	}

	secret.Data[utils.NonVaultPassword] = []byte{}

	return utils.CreateRuntimeObjectContextWrapper(ctx, secret, secret.ObjectMeta)
}
