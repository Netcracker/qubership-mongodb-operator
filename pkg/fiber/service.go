package fiber

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/dr"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/vault"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type MongoFiberService interface {
	GetKeyFile() string
	Health() Status
	RotateRoles() error
	AddDRReplicas() error
	FlushInMemoryData() error
	Compact(dbName string, collectionName string) error
	CompactAll(dbName string) error
	GetRSStatus() []RSStatus
	UpdateCtx(ctx core.ExecutionContext)
}

type Status struct {
	Status string `json:"status,omitempty"`
}

type RSStatus struct {
	Name     string `json:"name,omitempty"`
	StateStr string `json:"stateStr,omitempty"`
}
type MongoFiberServiceImpl struct {
	ctx core.ExecutionContext
}

var newMongoFiberServiceOnce sync.Once
var mongofiberService MongoFiberService

func NewMongoFiberService(ctx core.ExecutionContext) MongoFiberService {
	newMongoFiberServiceOnce.Do(func() {
		mongofiberService = &MongoFiberServiceImpl{ctx: ctx}
	})

	return mongofiberService
}

func (m *MongoFiberServiceImpl) UpdateCtx(ctx core.ExecutionContext) {
	m.ctx = ctx
}

func (m *MongoFiberServiceImpl) GetKeyFile() string {
	request := m.ctx.Get(constants.ContextRequest).(reconcile.Request)
	log := m.ctx.Get(constants.ContextLogger).(*zap.Logger)

	secret, err := utils.ReadSecret(m.ctx, utils.MongoSecret, request.Namespace)
	core.PanicError(err, log.Error, fmt.Sprintf("Could not recieve secret %s", utils.MongoSecret))

	return string(secret.Data[utils.MongoSecretKeyFile])
}

func (m *MongoFiberServiceImpl) Health() Status {
	log := m.ctx.Get(constants.ContextLogger).(*zap.Logger)
	mongoImpl := m.ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)

	spec := m.ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)

	log.Debug(fmt.Sprintf("spec.Spec.DisasterRecovery.Mode is %s", spec.Spec.DisasterRecovery.Mode))
	status, err := mongoImpl.GetClusterStatus(spec.Spec.DisasterRecovery.Mode, spec.Spec.SchemaSettings.ThisDomainName,
		spec.Spec.SchemaSettings.CnfReplicaSize, spec.Spec.SchemaSettings.DataReplicaSize, spec.Spec.SchemaSettings.ShardCount, spec.Spec.SchemaSettings.Sharded)

	log.Debug(fmt.Sprintf("Final Status is %s", status))
	core.PanicError(err, log.Error, "Could not get cluster status")

	return Status{string(status)}

}
func (m *MongoFiberServiceImpl) GetRSStatus() []RSStatus {
	log := m.ctx.Get(constants.ContextLogger).(*zap.Logger)
	mongoImpl := m.ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)

	spec := m.ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	clsuterReplicas := mongoImpl.GetClusterRSStatus(spec.Spec.SchemaSettings.CnfReplicaSize, spec.Spec.SchemaSettings.ShardCount, spec.Spec.SchemaSettings.Sharded)

	var rsstatus = make([]RSStatus, 0)
	for _, item := range clsuterReplicas {
		var temprsstatus = make([]RSStatus, 0)
		err := json.Unmarshal([]byte(item), &temprsstatus)
		if err != nil {
			core.PanicError(err, log.Error, "Could not get cluster status")
		}
		rsstatus = append(rsstatus, temprsstatus...)
	}
	return rsstatus

}

func (m *MongoFiberServiceImpl) AddDRReplicas() error {
	err := (&dr.AddCNFReplicas{}).Execute(m.ctx)
	if err != nil {
		return err
	}
	err = (&dr.AddDATAReplicas{}).Execute(m.ctx)
	if err != nil {
		return err
	}

	return nil
}

func (m *MongoFiberServiceImpl) RotateRoles() error {
	spec := m.ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	logger := m.ctx.Get(constants.ContextLogger).(*zap.Logger)
	request := m.ctx.Get(constants.ContextRequest).(reconcile.Request)
	vaultHelper := m.ctx.Get(constants.ContextVault).(vault.VaultHelper)
	utilsHelper := m.ctx.Get(utils.KubernetesHelperImpl).(core.KubernetesHelper)
	cloudPublicHost := spec.Spec.CloudPublicHost
	roles := make([]string, 0)
	podSelectors := make([]map[string]string, 0)

	registerRoleForRotation(m.ctx, logger, &roles, spec.Spec.MongoDB.MongoRootSecretName, cloudPublicHost, request.Namespace, spec.Spec.ServiceAccountName)

	// TODO services
	// if spec.Spec.Backup.Install {
	// 	podSelectors = append(podSelectors, map[string]string{
	// 		utils.Name: utils.BackupDaemon,
	// 	})
	// 	registerRoleForRotation(m.ctx, logger, &roles, utils.BackupBackupCreds, cloudPublicHost, request.Namespace, spec.Spec.ServiceAccountName)
	// 	registerRoleForRotation(m.ctx, logger, &roles, utils.BackupRestoreCreds, cloudPublicHost, request.Namespace, spec.Spec.ServiceAccountName)
	// }

	// if spec.Spec.Dbaas.Install {
	// 	podSelectors = append(podSelectors, map[string]string{
	// 		utils.Name: utils.DbaasName,
	// 	})
	// 	registerRoleForRotation(m.ctx, logger, &roles, utils.DbaasAdminCreds, cloudPublicHost, request.Namespace, spec.Spec.ServiceAccountName)
	// }

	// if spec.Spec.PrometheusExporter.Install {
	// 	podSelectors = append(podSelectors, map[string]string{
	// 		utils.Name: utils.MongoPrometheusExporter,
	// 	})
	// 	registerRoleForRotation(m.ctx, logger, &roles, utils.MonitoringCreds, cloudPublicHost, request.Namespace, spec.Spec.ServiceAccountName)
	// }

	for _, role := range roles {
		if err := vaultHelper.RotateRole(role); err != nil {
			return err
		}
	}

	for _, selector := range podSelectors {

		podList, err := utilsHelper.ListPods(request.Namespace, selector)
		if err != nil {
			return err
		}

		for podIdx := 0; podIdx < len(podList.Items); podIdx++ {
			podForRestart := podList.Items[podIdx]
			logger.Info(fmt.Sprintf("Restarting %v", podForRestart.ObjectMeta.Name))
			if err = utilsHelper.RestartPod(&podForRestart, request.Namespace, spec.Spec.WaitSeconds); err != nil {
				return err
			}
		}
	}

	return nil
}

func registerRoleForRotation(ctx core.ExecutionContext, logger *zap.Logger, roles *[]string,
	secretName, cloudPublicHost, namespace, serviceAccount string) {
	creds, err := utils.ReadSecret(ctx, secretName, namespace)
	core.PanicError(err, logger.Error, secretName+" secret reading failed")
	*roles = append(*roles, vault.GetVaultRoleName(cloudPublicHost, namespace, serviceAccount, string(creds.Data[utils.Username])))
}

func (m *MongoFiberServiceImpl) FlushInMemoryData() error {
	spec := m.ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	log := m.ctx.Get(constants.ContextLogger).(*zap.Logger)
	mongoImpl := m.ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)

	err := mongoImpl.ExecuteFlushData(m.ctx, spec.Spec.SchemaSettings.ShardCount)
	if err != nil {
		core.PanicError(err, log.Error, "Mongodb shards flush command failed.")
		return err
	}

	return nil
}

func (m *MongoFiberServiceImpl) Compact(dbName string, collectionName string) error {
	spec := m.ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	log := m.ctx.Get(constants.ContextLogger).(*zap.Logger)
	mongoImpl := m.ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)

	err := mongoImpl.Compact(spec.Spec.SchemaSettings.ShardCount, dbName, collectionName)
	if err != nil {
		core.PanicError(err, log.Error, "Mongodb disk space free command failed.")
		return err
	}

	return nil
}

func (m *MongoFiberServiceImpl) CompactAll(dbName string) error {
	spec := m.ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	log := m.ctx.Get(constants.ContextLogger).(*zap.Logger)
	mongoImpl := m.ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)

	err := mongoImpl.CompactAll(spec.Spec.SchemaSettings.ShardCount, dbName)
	if err != nil {
		core.PanicError(err, log.Error, "Compact command failed while running for all collection.")
		return err
	}

	return nil
}
