// Copyright 2024-2025 NetCracker Technology Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

	//there must be a better way to set all this vars like https://github.com/caarlos0/env
	// Defaults
	appPath := "/" + appName
	profiler := utils.GetEnvAsBool("PROFILER", false)
	namespace := utils.GetEnv("NAMESPACE", "")
	port := utils.GetEnvAsInt("PORT", mUtils.DefaultPort)
	apiUser := mUtils.GetSecret(
		"/var/run/secrets/mongodb/dbaas-aggregator/username",
		"dbaas-adapter",
	)
	apiPass := mUtils.GetSecret(
		"/var/run/secrets/mongodb/dbaas-aggregator/password",
		"dbaas-adapter",
	)
	aggregatorAdapterAddress := utils.GetEnv("DBAAS_ADAPTER_ADDRESS", fmt.Sprintf("http://dbaas-%s-adapter.%s:8080", appName, namespace))
	aggregatorRegistrationAddress := utils.GetEnv("DBAAS_AGGREGATOR_REGISTRATION_ADDRESS", "http://dbaas-aggregator.dbaas:8080")
	aggregatorRegistrationIdentifier := utils.GetEnv("DBAAS_AGGREGATOR_PHYSICAL_DATABASE_IDENTIFIER", appName)

	aggregatorRegistrationUser := mUtils.GetSecret(
		"/var/run/secrets/mongodb/dbaas-registration/username",
		"cluster-dba",
	)

	aggregatorRegistrationPass := mUtils.GetSecret(
		"/var/run/secrets/mongodb/dbaas-registration/password",
		"Bnmq5567_PO",
	)
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
		"/var/run/secrets/mongodb/mongo-admin/username",
		"root",
	)

	mongoPass := mUtils.GetSecret(
		"/var/run/secrets/mongodb/mongo-admin/password",
		"root",
	)

	authDb := mUtils.GetSecret(
		"/var/run/secrets/mongodb/mongo-admin/auth-database",
		"admin",
	)

	apiVersion := utils.GetEnv("API_VERSION", "v1")
	multiUserEnabled := utils.GetEnvAsBool("MULTI_USERS_ENABLED", false)

	// Backup Daemon Administration
	backupAddress := utils.GetEnv("BACKUP_DAEMON_ADDRESS", fmt.Sprintf("http://%s-backup-daemon:8080", appName))
	credsVoid := "dbaas_bckp_nosql_void_covers_empty_env_var"

	backupDaemonApiUser := mUtils.GetSecret(
		"/var/run/secrets/mongodb/backup-api/username",
		credsVoid,
	)

	backupDaemonApiUPass := mUtils.GetSecret(
		"/var/run/secrets/mongodb/backup-api/password",
		credsVoid,
	)

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
		false,
		&utils.VaultClient{},
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

		prefix := fmt.Sprintf("/api/%s/dbaas/adapter/mongodb/databases/", apiVersion)
		app.Put(prefix+":dbName/settings", dbAdminImpl.(*pkg.MongoDbAdministration).UpdateMongodbSettingsHandler)
		app.Get("/api/version", func(c *fiber.Ctx) error {
			return c.SendString(apiVersion)
		})

		return nil
	}))
}
