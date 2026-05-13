package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/Netcracker/qubership-dbaas-adapter-core/pkg/dao"
	"github.com/Netcracker/qubership-dbaas-adapter-core/pkg/dbaas"
	fiber2 "github.com/Netcracker/qubership-dbaas-adapter-core/pkg/impl/fiber"
	"github.com/Netcracker/qubership-dbaas-adapter-core/pkg/service"
	"github.com/Netcracker/qubership-dbaas-adapter-core/pkg/utils"
	"github.com/Netcracker/qubership-dbaas-mongo/pkg"
	mUtils "github.com/Netcracker/qubership-dbaas-mongo/utils"
	"github.com/gofiber/fiber/v2"
)

func main() {
	logger := utils.GetLogger(mUtils.GetEnvBool("DEBUG_LOG", false))
	ctxLogger := utils.AddLoggerContext(logger, context.Background())
	appName := "mongodb"

	// Vault
	vaultEnabled := utils.GetEnvAsBool("VAULT_ENABLED", false)
	var vaultClient *utils.VaultClient
	if vaultEnabled {
		vaultConfig := utils.VaultConfig{
			IsVaultEnabled:  vaultEnabled,
			Address:         utils.GetEnv("VAULT_ADDR", ""),
			VaultRole:       utils.GetEnv("VAULT_ROLE", "mongo-sa"),
			VaultRotPeriod:  utils.GetEnv("VAULT_ROTATION_PERIOD", "86400"),
			VaultAuthMethod: utils.GetEnv("VAULT_AUTH_METHOD", ""),
			VaultDBName:     utils.GetEnv("VAULT_DB_ENGINE_NAME", "mongodb"),
		}
		vaultClient = utils.NewVaultClient(vaultConfig)
	}
	//there must be a better way to set all this vars like https://github.com/caarlos0/env
	// Defaults
	appPath := "/" + appName
	profiler := utils.GetEnvAsBool("PROFILER", false)
	namespace := utils.GetEnv("NAMESPACE", "")
	port := utils.GetEnvAsInt("PORT", mUtils.DefaultPort)
	apiUser := mUtils.GetSecret(
		"/var/run/secrets/dbaas-aggregator/username",
		"dbaas-adapter",
	)
	apiPass := mUtils.GetSecret(
		"/var/run/secrets/dbaas-aggregator/password",
		"dbaas-adapter",
	)

	apiPass = checkForVaultPassword(vaultEnabled, apiPass, vaultClient)
	aggregatorAdapterAddress := utils.GetEnv("DBAAS_ADAPTER_ADDRESS", fmt.Sprintf("http://dbaas-%s-adapter.%s:8080", appName, namespace))
	aggregatorRegistrationAddress := utils.GetEnv("DBAAS_AGGREGATOR_REGISTRATION_ADDRESS", "http://dbaas-aggregator.dbaas:8080")
	aggregatorRegistrationIdentifier := utils.GetEnv("DBAAS_AGGREGATOR_PHYSICAL_DATABASE_IDENTIFIER", appName)

	aggregatorRegistrationUser := mUtils.GetSecret(
		"/var/run/secrets/dbaas-registration/username",
		"cluster-dba",
	)

	aggregatorRegistrationPass := mUtils.GetSecret(
		"/var/run/secrets/dbaas-registration/password",
		"Bnmq5567_PO",
	)
	aggregatorRegistrationPass = checkForVaultPassword(vaultEnabled, aggregatorRegistrationPass, vaultClient)
	aggregatorRegistrationDelay := utils.GetEnvAsInt("DBAAS_AGGREGATOR_REGISTRATION_FIXED_DELAY_MS", 150000)
	aggregatorRegistrationRetryTime := utils.GetEnvAsInt("DBAAS_AGGREGATOR_REGISTRATION_RETRY_TIME_MS", 60000)
	aggregatorRegistrationRetryDelay := utils.GetEnvAsInt("DBAAS_AGGREGATOR_REGISTRATION_RETRY_DELAY_MS", 5000)
	var aggregatorRegistrationLabels map[string]string
	if regLabels := utils.GetEnv("DBAAS_AGGREGATOR_REGISTRATION_LABELS", ""); regLabels != "" {
		err := json.Unmarshal([]byte(regLabels), &aggregatorRegistrationLabels)
		if err != nil {
			panic(err)
		}
	} else {
		aggregatorRegistrationLabels = map[string]string{}
	}

	// DB Administration
	mognoHost := utils.GetEnv("MONGO_HOST", fmt.Sprintf("mongos.%s", namespace))
	mongoPort := utils.GetEnvAsInt("MONGO_PORT", 27017)

	mongoUser := mUtils.GetSecret(
		"/var/run/secrets/mongo-admin/username",
		"root",
	)

	mongoPass := mUtils.GetSecret(
		"/var/run/secrets/mongo-admin/password",
		"root",
	)

	mongoPass = checkForVaultPassword(vaultEnabled, mongoPass, vaultClient)
	authDb := mUtils.GetSecret(
		"/var/run/secrets/mongo-admin/auth-database",
		"admin",
	)

	apiVersion := utils.GetEnv("API_VERSION", "v1")
	multiUserEnabled := utils.GetEnvAsBool("MULTI_USERS_ENABLED", false)

	// Backup Daemon Administration
	backupAddress := utils.GetEnv("BACKUP_DAEMON_ADDRESS", fmt.Sprintf("http://%s-backup-daemon:8080", appName))
	credsVoid := "dbaas_bckp_nosql_void_covers_empty_env_var"

	backupDaemonApiUser := mUtils.GetSecret(
		"/var/run/secrets/backup-api/username",
		credsVoid,
	)

	backupDaemonApiUPass := mUtils.GetSecret(
		"/var/run/secrets/backup-api/password",
		credsVoid,
	)

	backupDaemonApiUPass = checkForVaultPassword(vaultEnabled, backupDaemonApiUPass, vaultClient)

	var backupAdminServiceImpl service.BackupAdministrationService
	if backupAddress == "" {
		ctxLogger.Warn("Backup address is not set. Backup API is not working.")
	} else if backupDaemonApiUser == credsVoid || backupDaemonApiUPass == credsVoid {
		ctxLogger.Warn("Backup credentials are not set. Backup API is not working.")
	} else {
		if backupDaemonApiUser == "" || backupDaemonApiUPass == "" {
			ctxLogger.Warn("Assuming Backup Daemon is available w/o credentials, because BACKUP_DAEMON_API_CREDENTIALS_USERNAME or BACKUP_DAEMON_API_CREDENTIALS_PASSWORD is empty.")
			backupDaemonApiUser = ""
			backupDaemonApiUPass = ""
		}

		client := &http.Client{}
		if strings.Contains(backupAddress, "https") {
			cert := mUtils.GetCACert()
			if cert == "" {
				utils.PanicError(errors.New(""), ctxLogger.Error, "CA Certificate is empty or not set")
			}
			if err := utils.ConfigureHttpsForClientWithCertificate(client, cert); err != nil {
				utils.PanicError(err, ctxLogger.Error, "Failed to set up https client")
			}
		}

		backupAdminServiceImpl = service.DefaultBackupAdministrationService(
			logger,
			backupAddress,
			backupDaemonApiUser,
			backupDaemonApiUPass,
			false,
			client,
			63, nil)
	}

	// Supports
	supports := dao.SupportsBase{
		Users:             true,
		Settings:          true,
		DescribeDatabases: true,
		AdditionalKeys: dao.Supports{
			"backupRestore": backupAdminServiceImpl != nil,
			"vault":         vaultEnabled,
		},
	}

	basicRegistrationAuth := dao.BasicAuth{
		Username: aggregatorRegistrationUser,
		Password: aggregatorRegistrationPass,
	}

	dbaasClient, err := dbaas.NewDbaasClient(aggregatorRegistrationAddress, &basicRegistrationAuth, nil)
	if dbaasClient == nil {
		panic(fmt.Errorf("Failed to establish connection to DBaaS aggregator, err: %v", err))
	}

	if err != nil {
		ctxLogger.Error(fmt.Sprintf("Failed to get DBaaS aggregator version, err %v. Setting default API version", err))
	}

	version, _ := dbaasClient.GetVersion() // if err != nil it will fail in the condition above
	if version == "v3" {
		apiVersion = "v2"
	} else {
		apiVersion = "v1"
	}

	var dbAdminImpl service.DbAdministration = pkg.GetMongoDbAdministration(
		logger,
		mognoHost, mongoPort, mongoUser, mongoPass, authDb,
		vaultEnabled,
		apiVersion,
		[]string{"admin", "rw", "ro", "streaming"},
		map[string]bool{
			mUtils.FeatureMultiUsers: multiUserEnabled,
			mUtils.FeatureTLS:        utils.IsTLSEnabledForMainService(),
		},
	)

	admService := service.NewCoreAdministrationService(
		namespace,
		port,
		dbAdminImpl,
		logger,
		vaultEnabled,
		vaultClient,
		"",
	)

	// Defaults
	log.Fatal(fiber2.RunFiberServer(port, func(app *fiber.App, ctx context.Context) error {
		fiber2.BuildFiberDBaaSAdapterHandlers(
			app,
			apiUser,
			apiPass,
			appPath,
			admService,
			service.NewPhysicalRegistrationService(
				appName,
				logger,
				aggregatorRegistrationIdentifier,
				aggregatorAdapterAddress,
				dao.BasicAuth{
					Username: apiUser,
					Password: apiPass,
				},
				aggregatorRegistrationLabels,
				dbaasClient,
				aggregatorRegistrationDelay,
				aggregatorRegistrationRetryTime,
				aggregatorRegistrationRetryDelay,
				admService,
				ctx,
			),
			backupAdminServiceImpl,
			supports.ToMap(),
			logger,
			profiler, "")
		return nil
	}))
}

func checkForVaultPassword(vaultEnabled bool, password string, vaultClient *utils.VaultClient) string {
	if vaultEnabled {
		if utils.IsVaultPassword(password) {
			passwordFromVault, err := vaultClient.ReadPasswordFromKv(utils.GetSecretPath(password))
			if err != nil {
				panic(err)
			}
			return passwordFromVault
		}
	}
	return password
}
