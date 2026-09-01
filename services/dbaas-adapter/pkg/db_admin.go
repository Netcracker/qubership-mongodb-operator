package pkg

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/Netcracker/qubership-dbaas-adapter-core/pkg/dao"
	"github.com/Netcracker/qubership-dbaas-adapter-core/pkg/service"
	"github.com/Netcracker/qubership-dbaas-adapter-core/pkg/utils"
	mUtils "github.com/Netcracker/qubership-dbaas-mongo/utils"
	"github.com/docker/distribution/uuid"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

var (
	dbaasMetadata            = "_dbaas_metadata"
	enableShardingCmd        = "enableSharding"
	readWrite                = "readWrite"
	dbOwner                  = "dbOwner"
	clusterAdmin             = "clusterAdmin"
	streaming                = "streaming"
	read                     = "read"
	dbNameRegexpExpression   = "^[_A-z0-9\\-]*$"
	dbNameRegexp, _          = regexp.Compile(dbNameRegexpExpression)
	prefixRegexpExpression   = "^[_A-z0-9\\-]*$"
	prefixRegexp, _          = regexp.Compile(prefixRegexpExpression)
	userNameRegexpExpression = "^[_A-z0-9\\-]*$"
	userNameRegexp, _        = regexp.Compile(userNameRegexpExpression)

	UserResourceKind = "user"
	DbResourceKind   = "database"

	metadataKey = "metadata"

	rolesMapping = map[string]string{
		"admin":     dbOwner,
		"rw":        readWrite,
		"ro":        read,
		"streaming": streaming,
	}
)

type MongoDbAdministration struct {
	logger            *zap.Logger
	mongodService     MongoService
	apiVersion        string
	supportedRoles    []string
	supportedFeatures map[string]bool
}

type metadataDocument struct {
	Metadata map[string]interface{}
}

func GetMongoDbAdministration(logger *zap.Logger, hostname string, port int, user string, pass string, authDb string, apiVersion string, supportedRoles []string, supportedFeatures map[string]bool) service.DbAdministration {
	return &MongoDbAdministration{
		logger: logger,
		mongodService: &MongoServiceImpl{
			logger: logger,
			configuration: &MongoConfigurationImpl{
				hostname: hostname,
				port:     port,
				user:     user,
				pass:     pass,
				authDb:   authDb,
			},
		},
		apiVersion:        apiVersion,
		supportedRoles:    supportedRoles,
		supportedFeatures: supportedFeatures,
	}
}

// todo identical
func (c *MongoDbAdministration) getConnectionProperties(dbName string, username string, password string, role string) dao.ConnectionProperties {
	authDbName := dbName
	connectionProps := dao.ConnectionProperties{
		"url": fmt.Sprintf("mongodb://%s:%d/%s",
			c.mongodService.GetConfiguration().GetHost(),
			c.mongodService.GetConfiguration().GetPort(),
			authDbName),
		"username":   username,
		"password":   password,
		"authDbName": authDbName,
		"dbName":     dbName,
	}

	if len(role) > 0 && c.GetVersion() != "v1" {
		connectionProps["role"] = role
	}

	return connectionProps

}

func (c *MongoDbAdministration) UpdateMongodbSettingsHandler(ctx *fiber.Ctx) error {

	logger := utils.AddLoggerContext(c.logger, ctx.Context())
	logger.Info("UpdateMongodbSettings request accepted")
	log.Println("UpdateMongodbSettings request accepted")
	dbName := ctx.Params("dbName")

	var settings map[string]interface{}
	if err := ctx.BodyParser(&settings); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	shardSettings, err := c.parseShardingSettings(settings)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := c.mongodService.EnableShardingAndCreateCollection(ctx.Context(), dbName, shardSettings); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.SendStatus(fiber.StatusOK)
}

func (c *MongoDbAdministration) validateRequestParamsAndGetLogicalDbName(ctx context.Context, requestOnCreateDb dao.DbCreateRequest) (string, *mUtils.Settings, bool, bool, error) {
	var dbName string
	var moveShard, shardCollection bool
	s := &mUtils.Settings{}

	if requestOnCreateDb.DbName != "" {
		dbName = requestOnCreateDb.DbName
	} else if requestOnCreateDb.NamePrefix != nil {
		if len(*requestOnCreateDb.NamePrefix) > 0 {
			if !prefixRegexp.MatchString(*requestOnCreateDb.NamePrefix) {
				return "", s, moveShard, shardCollection, &utils.ExecutionError{Msg: "namePrefix must not contain the following characters :/'\\\"%?.,@#&*()"}
			} else {
				dbName = utils.RegenerateDbName(*requestOnCreateDb.NamePrefix, 63)
			}
		}
	} else {
		if classifier, ok := requestOnCreateDb.Metadata["classifier"].(map[string]interface{}); ok {
			if namespace, ok := classifier["namespace"].(string); ok {
				if microserviceName, ok := classifier["microserviceName"].(string); ok {
					shrinkedName, err := utils.PrepareDatabaseName(namespace, microserviceName, 63)
					if err != nil {
						panic(err)
					}
					dbName = shrinkedName
				}
			}
		}
	}

	if !dbNameRegexp.MatchString(dbName) {
		return "", s, moveShard, shardCollection, &utils.ExecutionError{Msg: "dbName must not contain the following characters :/'\\\"%?.,@#&*()"}
	}

	if !userNameRegexp.MatchString(requestOnCreateDb.Username) {
		return "", s, moveShard, shardCollection, &utils.ExecutionError{Msg: "userName must not contain the following characters :/'\\\"%?.,@#&*()"}
	}

	available, err := c.mongodService.IsDatabaseAvailable(ctx, dbName)
	if err != nil {
		return "", s, moveShard, shardCollection, &utils.ExecutionError{Msg: fmt.Sprintf("error getting database list %v", err)}
	}

	if !available {
		return "", s, moveShard, shardCollection, &utils.ExecutionError{Msg: fmt.Sprintf("Database already exists")}
	}

	settings := requestOnCreateDb.Settings
	targetShard, _ := settings["targetShard"].(string)
	if targetShard != "" {
		valid, err := c.mongodService.IsValidShard(ctx, targetShard)
		if err != nil {
			return "", s, moveShard, shardCollection, &utils.ExecutionError{Msg: fmt.Sprintf("error getting shard details %v", err)}
		}
		if !valid {
			return "", s, moveShard, shardCollection, &utils.ExecutionError{Msg: fmt.Sprintf("Shard [%v] is not a valid shard", targetShard)}
		}
		s.TargetShard = targetShard
		moveShard = true
	}

	parsedShardSettings, parseErr := c.parseShardingSettings(settings)
	if parseErr != nil {
		return "", s, moveShard, shardCollection, parseErr
	}
	s.ShardingSettings = parsedShardSettings.ShardingSettings
	s.Extra = parsedShardSettings.Extra
	shardCollection = len(s.ShardingSettings) > 0

	return dbName, s, moveShard, shardCollection, nil
}

func (c *MongoDbAdministration) CreateDatabase(ctx context.Context, requestOnCreateDb dao.DbCreateRequest) (string, *dao.LogicalDatabaseDescribed, error) {
	logger := utils.AddLoggerContext(c.logger, ctx)
	logger.Debug("create db request accepted")
	logicalDatabaseName, shardSettings, moveShard, shardCollection, validationError := c.validateRequestParamsAndGetLogicalDbName(ctx, requestOnCreateDb)
	if validationError != nil {
		return "", nil, validationError
	}

	// Flag to track if db creation succeeded
	dbCreated := false
	defer func() {
		if !dbCreated {
			logger.Warn(fmt.Sprintf("sharding or setup failed, dropping database %s", logicalDatabaseName))
			err := c.mongodService.DropDatabase(ctx, logicalDatabaseName)
			if err != nil {
				logger.Error(fmt.Sprintf("failed to drop database %s: %v", logicalDatabaseName, err))
			}
		}
		logger.Debug("create db request finished")
	}()

	var connectionProps []dao.ConnectionProperties
	var resources []dao.DbResource

	//create user for database
	authDb := logicalDatabaseName

	if c.GetVersion() == "v2" {
		logger.Debug("v2 starting role creating")
		for _, role := range c.GetSupportedRoles() {
			var username string
			var password string
			if role == "admin" {
				if requestOnCreateDb.Username != "" {
					username = requestOnCreateDb.Username
				}
				if requestOnCreateDb.Password != "" {
					password = requestOnCreateDb.Password
				}
			} else {
				username = fmt.Sprintf("dbaas_%s", c.generateUUID())
				password = c.generateUUID()
			}
			logger.Debug(fmt.Sprintf("CreateOrUpdateUser with role: %s started", role))
			err := c.mongodService.CreateOrUpdateUser(ctx, username, password, logicalDatabaseName, authDb, rolesMapping[role], false)
			if err != nil {
				return "", nil, err
			}
			logger.Debug(fmt.Sprintf("CreateOrUpdateUser with role: %s finished", role))

			connectionProps = append(connectionProps, c.getConnectionProperties(logicalDatabaseName, username, password, role))
			resources = append(resources, dao.DbResource{
				Kind: UserResourceKind,
				Name: fmt.Sprintf("%s:%s", authDb, username),
			})
		}
	} else {
		logger.Debug("v1 starting role creating")
		err := c.mongodService.CreateOrUpdateUser(ctx, requestOnCreateDb.Username, requestOnCreateDb.Password, logicalDatabaseName, authDb, "", false)
		logger.Debug("v1 role created")
		if err != nil {
			return "", nil, err
		}

		connectionProps = append(connectionProps, c.getConnectionProperties(logicalDatabaseName, requestOnCreateDb.Username, requestOnCreateDb.Password, ""))
		resources = append(resources, dao.DbResource{
			Kind: UserResourceKind,
			Name: fmt.Sprintf("%s:%s", authDb, requestOnCreateDb.Username),
		})
	}

	//fill metadata
	logger.Debug("filling metadata started")
	err := c.mongodService.RunWithGrants(ctx, logicalDatabaseName, func(service MongoService) error {
		_, err := service.InsertOne(ctx, logicalDatabaseName, dbaasMetadata, bson.D{{"metadata", requestOnCreateDb.Metadata}})
		return err
	})

	if err != nil {
		return "", nil, err
	}
	logger.Debug("filling metadata finished")

	if shardCollection {
		logger.Debug("sharding setup started")
		err = c.mongodService.EnableShardingAndCreateCollection(
			ctx,
			logicalDatabaseName,
			shardSettings,
		)
		if err != nil {
			return "", nil, err
		}
		logger.Debug("sharding setup completed")
	}
	dbCreated = true
	if moveShard {
		logger.Debug("moving database to target shard")
		err = c.mongodService.EnsureDBOnShard(ctx, logicalDatabaseName, shardSettings.TargetShard)
		if err != nil {
			logger.Warn(fmt.Sprintf("Failed to move database [%v] to shard [%v]", logicalDatabaseName, shardSettings.TargetShard))
		}
		logger.Debug("moved database to target shard")
	}

	return logicalDatabaseName,
		&dao.LogicalDatabaseDescribed{
			ConnectionProperties: connectionProps,
			Resources: append(resources, dao.DbResource{
				Kind: DbResourceKind,
				Name: logicalDatabaseName,
			}),
		},
		nil
}

func (c *MongoDbAdministration) DescribeDatabases(ctx context.Context, logicalDatabases []string, showResources bool, showConnections bool) map[string]dao.LogicalDatabaseDescribed {
	dbsDescription := make(map[string]dao.LogicalDatabaseDescribed)
	for _, logicalDatabaseName := range logicalDatabases {
		dbResourcesList := []dao.DbResource{
			{
				Kind: DbResourceKind,
				Name: logicalDatabaseName,
			},
		}

		roles, getRolesErr := c.mongodService.GetRoles(ctx, logicalDatabaseName)
		if getRolesErr != nil {
			panic(getRolesErr)
		}
		for _, role := range roles {
			dbResourcesList = append(dbResourcesList, dao.DbResource{
				Kind: UserResourceKind,
				Name: logicalDatabaseName + ":" + role,
			})
		}

		user, err := c.mongodService.GetDbOwner(ctx, logicalDatabaseName)
		if err != nil {
			panic(err)
		}
		var connectionProps []dao.ConnectionProperties
		connectionProps = append(connectionProps, c.getConnectionProperties(
			logicalDatabaseName,
			user,
			"",
			""))
		dbsDescription[logicalDatabaseName] = dao.LogicalDatabaseDescribed{
			ConnectionProperties: connectionProps,
			Resources:            dbResourcesList,
		}
	}
	return dbsDescription
}

func (c *MongoDbAdministration) GetDatabases(ctx context.Context) []string {
	dbs, err := c.mongodService.ListDatabases(ctx)
	if err != nil {
		panic(err)
	}
	return dbs
}

func (c *MongoDbAdministration) DropResources(ctx context.Context, resources []dao.DbResource) []dao.DbResource {
	var dropStatuses []dao.DbResource
	for _, resource := range resources {
		var dropError error
		if resource.Kind == UserResourceKind {
			dropError = c.mongodService.DropUser(ctx, resource.Name)
		} else if resource.Kind == DbResourceKind {
			// to drop db dbaas user must have dbOwner role
			dropError = c.mongodService.RunWithGrants(ctx, resource.Name, func(serive MongoService) error {
				return c.mongodService.DropDatabase(ctx, resource.Name)
			})
		}

		if dropError != nil {
			resource.Status = dao.DELETE_FAILED
			resource.ErrorMessage = fmt.Sprintf("%v", dropError)
		} else {
			resource.Status = dao.DELETED
		}

		dropStatuses = append(dropStatuses, resource)
	}

	return dropStatuses
}

func (c *MongoDbAdministration) GetMetadata(ctx context.Context, logicalDatabase string) map[string]interface{} {
	singleResult, err := c.mongodService.GetOne(ctx, logicalDatabase, dbaasMetadata)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil
		} else {
			panic(err)
		}
	}
	metadataDoc := &metadataDocument{}
	err = singleResult.Decode(metadataDoc)
	if err != nil {
		panic(err)
	}
	return metadataDoc.Metadata
}

func (c *MongoDbAdministration) UpdateMetadata(ctx context.Context, newMetadata map[string]interface{}, serviceName string) {
	err := c.mongodService.RunWithGrants(ctx, serviceName, func(service MongoService) error {
		_, err := service.InsertOrUpdate(ctx, serviceName, dbaasMetadata, bson.D{{"metadata", newMetadata}})
		return err
	})
	if err != nil {
		panic(err)
	}

}

func (c *MongoDbAdministration) generateUUID() string {
	return uuid.Generate().String()
}

func (c *MongoDbAdministration) generateUserName() string {
	return c.generateUUID()
}

func (c *MongoDbAdministration) GetDefaultCreateRequest() dao.DbCreateRequest {
	return dao.DbCreateRequest{
		Metadata: map[string]interface{}{},
		Password: c.generateUUID(),
		Username: c.generateUserName(),
	}
}

func (c *MongoDbAdministration) GetDefaultUserCreateRequest() dao.UserCreateRequest {
	return dao.UserCreateRequest{
		DbName:   "",
		Password: c.generateUUID(),
	}
}

func (c *MongoDbAdministration) PreStart() {}

func (c *MongoDbAdministration) CreateUser(ctx context.Context, userName string, requestOnCreateUser dao.UserCreateRequest) (*dao.CreatedUser, error) {
	var logicalDbName string
	var authDb string

	if requestOnCreateUser.DbName == "" {
		logicalDbName = c.mongodService.GetConfiguration().GetAuthDb()
	} else {
		logicalDbName = requestOnCreateUser.DbName
	}

	authDb = logicalDbName
	if userName == "" {
		userName = c.generateUserName()
	}
	if requestOnCreateUser.Password == "" {
		requestOnCreateUser.Password = c.generateUUID()
	}

	if !userNameRegexp.MatchString(userName) {
		return nil, &utils.ExecutionError{Msg: "userName must not contain the following characters :/'\\\"%?.,@#&*()"}
	}

	resources := []dao.DbResource{
		{
			Kind: UserResourceKind,
			Name: fmt.Sprintf("%s:%s", authDb, userName),
		},
	}

	err := c.mongodService.CreateOrUpdateUser(ctx, userName, requestOnCreateUser.Password, logicalDbName, authDb, rolesMapping[requestOnCreateUser.Role], false)

	if err != nil {
		return nil, err
	}

	connProp := c.getConnectionProperties(
		logicalDbName,
		userName,
		requestOnCreateUser.Password, "")

	return &dao.CreatedUser{
		ConnectionProperties: connProp,
		Resources:            resources,
		Name:                 logicalDbName,
		Role:                 requestOnCreateUser.Role,
	}, nil
}

func (c *MongoDbAdministration) MigrateToVault(ctx context.Context, dbName, userName string) error {
	authDb := c.mongodService.GetConfiguration().GetAuthDb()
	userMigrated, err := c.mongodService.IsUserExist(ctx, c.mongodService.GetConfiguration(), userName, authDb)
	if err != nil {
		return err
	}
	if !userMigrated {
		err = c.mongodService.CreateOrUpdateUser(ctx, userName, userName, dbName, authDb, "", true)
		if err != nil {
			return err
		}
	}
	oldUserCleanupNeeded, err := c.mongodService.IsUserExist(ctx, c.mongodService.GetConfiguration(), userName, dbName)
	if oldUserCleanupNeeded {
		if err != nil {
			return err
		}
		return c.mongodService.DropUser(ctx, fmt.Sprintf("%s:%s", dbName, userName))
	}
	return nil
}

func (c *MongoDbAdministration) GetDBPrefix() string {
	return "dbaas"
}

func (c *MongoDbAdministration) GetDBPrefixDelimiter() string {
	return "-"
}

func (c *MongoDbAdministration) GetVersion() dao.ApiVersion {
	return dao.ApiVersion(c.apiVersion)
}

func (c *MongoDbAdministration) GetSupportedRoles() []string {
	if c.supportedFeatures[mUtils.FeatureMultiUsers] {
		return c.supportedRoles
	}
	return []string{"admin"}
}

func (c *MongoDbAdministration) GetFeatures() map[string]bool {
	return c.supportedFeatures
}

func (c *MongoDbAdministration) CreateRoles(ctx context.Context, roles []dao.AdditionalRole) (success []dao.Success, failure *dao.Failure) {
	logger := utils.AddLoggerContext(c.logger, ctx)
	logger.Debug("CreateRoles was called")

	var additionalRoleIdInProcess string

	defer func() {
		if r := recover(); r != nil {
			logger.Error(fmt.Sprintf("error during additional roles creation %s", r))
			if failure == nil {
				failure = &dao.Failure{
					Id:      additionalRoleIdInProcess,
					Message: fmt.Sprintf("%s", r),
				}
			}
		}
	}()

	for _, additionalRole := range roles {
		additionalRoleIdInProcess = additionalRole.Id
		existingRoles := make([]string, 0)
		dbName := additionalRole.DbName

		newConProps := make([]dao.ConnectionProperties, 0)
		newResources := make([]dao.DbResource, 0)

		for _, connectionProperties := range additionalRole.ConnectionProperties {
			roleForCheck := connectionProperties["role"].(string)
			existingRoles = append(existingRoles, roleForCheck)
		}

		for _, role := range c.GetSupportedRoles() {
			if !Contains(existingRoles, role) {
				userCreateRequest := dao.UserCreateRequest{
					DbName: dbName,
					Role:   rolesMapping[role],
				}

				createdUser, err := c.CreateUser(ctx, "", userCreateRequest)
				if err != nil {
					failure = &dao.Failure{
						Id:      additionalRoleIdInProcess,
						Message: err.Error(),
					}
					break
				}

				createdUser.ConnectionProperties["role"] = role
				newConProps = append(newConProps, createdUser.ConnectionProperties)
				newResources = append(newResources, createdUser.Resources...)
			}
		}

		success = append(success, dao.Success{
			Id:                   additionalRole.Id,
			ConnectionProperties: newConProps,
			Resources:            newResources,
			DbName:               dbName,
		})
	}

	return success, failure
}

func Contains(slice []string, element string) bool {
	for _, v := range slice {
		if v == element {
			return true
		}
	}
	return false
}

func (c *MongoDbAdministration) parseShardingSettings(settings map[string]interface{}) (*mUtils.Settings, error) {
	s := &mUtils.Settings{}

	rawShardSettings, ok := settings["shardingSettings"]
	if !ok {
		return s, nil
	}

	shardSettingsSlice, ok := rawShardSettings.([]interface{})
	if !ok {
		return nil, &utils.ExecutionError{Msg: "shardingSettings must be a list of objects"}
	}

	for _, item := range shardSettingsSlice {
		var record mUtils.ShardingSettings
		collectionSetting, ok := item.(map[string]interface{})
		if !ok {
			return nil, &utils.ExecutionError{Msg: "each entry in shardingSettings must be an object"}
		}

		if v, ok := collectionSetting["collectionName"].(string); ok && v != "" {
			record.CollectionName = v
		} else {
			return nil, &utils.ExecutionError{Msg: "collectionName is required and must be a non-empty string"}
		}

		if v, ok := collectionSetting["shardKey"].(string); ok && v != "" {
			record.ShardKey = v
		} else {
			return nil, &utils.ExecutionError{Msg: fmt.Sprintf("shardKey is required and must be a non-empty string for collection %s", record.CollectionName)}
		}

		if v, ok := collectionSetting["strategy"].(string); ok {
			record.Strategy = strings.ToLower(v)
			if record.Strategy != "hashed" && record.Strategy != "ranged" {
				return nil, &utils.ExecutionError{Msg: fmt.Sprintf("strategy must be 'Hashed' or 'Ranged' for collection %s", record.CollectionName)}
			}
		} else {
			return nil, &utils.ExecutionError{Msg: fmt.Sprintf("strategy is required and must be a string for collection %s", record.CollectionName)}
		}

		s.ShardingSettings = append(s.ShardingSettings, record)
	}

	s.Extra = settings
	return s, nil
}
