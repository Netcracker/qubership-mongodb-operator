package pkg

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Netcracker/qubership-dbaas-adapter-core/pkg/utils"
	mUtils "github.com/Netcracker/qubership-dbaas-mongo/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type MongoService interface {
	RunWithGrants(ctx context.Context, dbName string, ff func(service MongoService) error) error
	RunWithClusterGrants(ctx context.Context, dbName string, ff func(service MongoService) error) error
	EnableShardingAndCreateCollection(ctx context.Context, dbName string, settings *mUtils.Settings) error
	EnableSharding(ctx context.Context, dbName string) error
	EnsureDBOnShard(ctx context.Context, dbName string, primaryShard string) error
	MovePrimary(ctx context.Context, dbName string, targetShard string) error
	IsValidShard(ctx context.Context, shardName string) (bool, error)
	ShardCollection(ctx context.Context, dbName string, settings mUtils.ShardingSettings) error
	GetConfiguration() MongoConfiguration
	CreateOrUpdateUser(ctx context.Context, username string, pass string, database string, authDb string, role string, force bool) error
	UpdateUserPassword(ctx context.Context, username string, password string, dbName string) error
	IsUserExist(ctx context.Context, conf MongoConfiguration, username string, dbName string) (bool, error)
	DropUser(ctx context.Context, username string) error
	DropDatabase(ctx context.Context, dbName string) error
	ListDatabases(ctx context.Context) ([]string, error)
	IsDatabaseAvailable(ctx context.Context, dbName string) (bool, error)
	InsertOrUpdate(ctx context.Context, dbName string, collection string, data bson.D) (*mongo.UpdateResult, error)
	InsertOne(ctx context.Context, dbName string, collection string, data bson.D) (*mongo.InsertOneResult, error)
	GetOne(ctx context.Context, dbName string, collection string) (*mongo.SingleResult, error)
	GetRoles(ctx context.Context, dbName string) ([]string, error)
	GetDbOwner(ctx context.Context, dbName string) (string, error)
	GetReplicaSetHosts() ([]string, error)
	CreateCollection(ctx context.Context, dbName string, collection string) error
}

type MongoServiceImpl struct {
	logger        *zap.Logger
	configuration MongoConfiguration
}

var _ MongoService = &MongoServiceImpl{}

func (r *MongoServiceImpl) RunWithGrants(ctx context.Context, dbName string, ff func(service MongoService) error) (err error) {
	result, err := r.GrantRole(ctx, dbName, dbOwner)
	if err != nil {
		return err
	} else if result != nil && result.Err() != nil {
		return result.Err()
	}

	defer func() {
		result, err = r.RevokeRole(ctx, dbName, dbOwner)
		if result != nil && result.Err() != nil {
			err = result.Err()
		}
	}()

	return ff(r)
}

func (r *MongoServiceImpl) RunWithClusterGrants(ctx context.Context, dbName string, ff func(service MongoService) error) (err error) {
	result, err := r.GrantRole(ctx, "admin", clusterAdmin)
	if err != nil {
		return err
	} else if result != nil && result.Err() != nil {
		return result.Err()
	}

	defer func() {
		result, err = r.RevokeRole(ctx, "admin", clusterAdmin)
		if result != nil && result.Err() != nil {
			err = result.Err()
		}
	}()

	return ff(r)
}

func (r *MongoServiceImpl) GetConfiguration() MongoConfiguration {
	return r.configuration
}

func (r *MongoServiceImpl) GrantRole(ctx context.Context, dbName, role string) (*mongo.SingleResult, error) {
	client, err := GetMongoClient(r.configuration)
	if err != nil {
		return nil, err
	}
	return client.Database(r.configuration.GetAuthDb()).RunCommand(ctx,
		bson.D{primitive.E{Key: "grantRolesToUser", Value: r.configuration.GetUser()},
			primitive.E{Key: "roles", Value: []bson.M{{"role": role, "db": dbName}}}}), nil

}

func (r *MongoServiceImpl) RevokeRole(ctx context.Context, dbName, role string) (*mongo.SingleResult, error) {
	client, err := GetMongoClient(r.configuration)
	if err != nil {
		return nil, err
	}

	return client.Database(r.configuration.GetAuthDb()).RunCommand(ctx,
		bson.D{primitive.E{Key: "revokeRolesFromUser", Value: r.configuration.GetUser()},
			primitive.E{Key: "roles", Value: []bson.M{{"role": role, "db": dbName}}}}), nil

}

func (r *MongoServiceImpl) CreateOrUpdateUser(ctx context.Context, username string, pass string, database string, authDb string, role string, force bool) error {
	logger := utils.AddLoggerContext(r.logger, ctx)
	logger.Debug(fmt.Sprintf("check user %v exists", username))
	exists, err := r.IsUserExist(ctx, r.configuration, username, authDb)
	if err != nil {
		return err
	}
	var callFunc func(client *mongo.Client) (interface{}, error)
	if !exists || force {
		callFunc = r.createUser(ctx, username, pass, database, authDb, role)

	} else {
		callFunc = r.updateUser(ctx, username, pass, database, authDb, role)
	}

	logger.Debug(fmt.Sprintf("create/update user %v started", username))
	client, err := GetMongoClient(r.configuration)
	if err != nil {
		return err
	}

	_, err = callFunc(client)
	if err != nil {
		return err
	}
	logger.Debug(fmt.Sprintf("create/update user %v finished", username))

	//if streaming - we create users in shards as well
	if role == streaming {
		rsHosts, err := r.GetReplicaSetHosts()

		if err != nil {
			logger.Error(fmt.Sprintf("Failed to get replicasets. This error can be ignored in case of single Mongos schema. Err: %v. ", err))
			return nil
		}

		for _, host := range rsHosts {
			conf := MongoConfigurationImpl{
				hostname: host,
				authDb:   r.configuration.GetAuthDb(),
				port:     r.configuration.GetPort(),
				user:     r.configuration.GetUser(),
				pass:     r.configuration.GetPassword(),
			}

			logger.Debug(fmt.Sprintf("check streaming user %v exists started", username))
			exists, err = r.IsUserExist(ctx, &conf, username, database)
			if err != nil {
				return err
			}
			logger.Debug(fmt.Sprintf("check streaming user %v exists finished", username))

			if exists {
				callFunc = r.updateUser(ctx, username, pass, database, authDb, role)
			} else {
				callFunc = r.createUser(ctx, username, pass, database, authDb, role)
			}

			logger.Debug(fmt.Sprintf("create/update streaming user %v started", username))
			client, err := GetMongoClient(&conf)
			if err != nil {
				return err
			}

			_, funcErr := callFunc(client)
			if funcErr != nil {
				return funcErr
			}
			logger.Debug(fmt.Sprintf("create/update streaming user %v finished", username))
		}
	}

	return err
}

func (r *MongoServiceImpl) createUser(ctx context.Context, username string, pass string, database string, authDb string, role string) func(client *mongo.Client) (interface{}, error) {
	logger := utils.AddLoggerContext(r.logger, ctx)
	var roles []bson.M
	if role == "" || role == dbOwner {
		roles = []bson.M{{"role": dbOwner, "db": database}}
	} else if role == "streaming" {
		roles = []bson.M{{"role": read, "db": database}, {"role": "streaming", "db": "admin"}}
	} else {
		roles = []bson.M{{"role": role, "db": database}}
	}

	if authDb != "" {
		database = authDb
	}
	return func(client *mongo.Client) (interface{}, error) {

		command := bson.D{primitive.E{Key: "createUser", Value: username},
			primitive.E{Key: "pwd", Value: pass},
			primitive.E{Key: "roles", Value: roles},
		}
		logger.Info(fmt.Sprintf("Creating user %s with role %v", username, roles))
		commandResult := client.Database(authDb).RunCommand(ctx, command)
		return nil, commandResult.Err()
	}
}

func (r *MongoServiceImpl) updateUser(ctx context.Context, username string, pass string, database string, authDb string, role string) func(client *mongo.Client) (interface{}, error) {
	logger := utils.AddLoggerContext(r.logger, ctx)
	var roles []bson.M
	if role == "" || role == dbOwner {
		roles = []bson.M{{"role": dbOwner, "db": database}}
	} else if role == "streaming" {
		roles = []bson.M{{"role": read, "db": database}, {"role": "streaming", "db": "admin"}}
	} else {
		roles = []bson.M{{"role": role, "db": database}}
	}

	var update bson.D
	if pass == "" {
		update = bson.D{{"updateUser", username}, {"roles", roles}}
	} else {
		update = bson.D{{"updateUser", username}, {"roles", roles}, {"pwd", pass}}
	}

	if authDb != "" {
		database = authDb
	}
	return func(client *mongo.Client) (interface{}, error) {
		logger.Info(fmt.Sprintf("Updating user %s role", username))
		commandResult := client.Database(database).RunCommand(ctx, update)

		return commandResult, commandResult.Err()
	}
}

func (r *MongoServiceImpl) IsUserExist(ctx context.Context, conf MongoConfiguration, username string, dbName string) (bool, error) {
	client, err := GetMongoClient(conf)
	if err != nil {
		return false, nil
	}
	userInfo := client.Database(dbName).RunCommand(ctx,
		bson.D{primitive.E{Key: "usersInfo", Value: username}})

	if userInfo.Err() != nil {
		return false, userInfo.Err()
	}

	var user bson.M
	userInfo.Decode(&user)

	users := []interface{}(user["users"].(primitive.A))
	return len(users) > 0, nil
}

func (r *MongoServiceImpl) GetReplicaSetHosts() ([]string, error) {
	client, err := GetMongoClient(r.configuration)
	if err != nil {
		return nil, err
	}
	listShards := client.Database("admin").RunCommand(context.TODO(),
		bson.D{primitive.E{Key: "listShards", Value: 1}})

	if listShards.Err() != nil {
		return nil, listShards.Err()
	}

	var shardsResponse bson.M
	listShards.Decode(&shardsResponse)

	shards := []interface{}(shardsResponse["shards"].(primitive.A))

	if len(shards) == 0 {
		return nil, fmt.Errorf("Shard list is empty")
	}

	var hosts []string = make([]string, len(shards))
	for i, shard := range shards {
		s := shard.(primitive.M)
		hosts[i] = strings.Split(s["host"].(string), "/")[1]
	}

	return hosts, nil
}

func (r *MongoServiceImpl) UpdateUserPassword(ctx context.Context, username string, password string, dbName string) error {
	client, err := GetMongoClient(r.configuration)
	logger := utils.AddLoggerContext(r.logger, ctx)
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("Updating user %s password", username))
	commandResult := client.Database(dbName).RunCommand(ctx, bson.D{{"updateUser", username}, {"pwd", password}})

	return commandResult.Err()

}

func (r *MongoServiceImpl) DropUser(ctx context.Context, username string) error {
	logger := utils.AddLoggerContext(r.logger, ctx)

	// Expecting format "dbName:username"
	entities := strings.Split(username, ":")
	if len(entities) != 2 {
		return errors.New("The format of the incoming data is incorrect. Name must be \"dbName:username\"")
	}
	dbName := entities[0]
	username = entities[1]

	client, err := GetMongoClient(r.configuration)
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("Checking if user %s exists in db %s", username, dbName))

	// Step 1: Check if the user exists
	var result bson.M
	err = client.Database(dbName).RunCommand(ctx, bson.D{
		{Key: "usersInfo", Value: username},
	}).Decode(&result)

	if err != nil {
		return fmt.Errorf("error checking user existence: %w", err)
	}

	users, ok := result["users"].(primitive.A)
	if !ok || len(users) == 0 {
		// User doesn't exist – no need to drop
		logger.Info(fmt.Sprintf("User %s does not exist in db %s. Skipping drop.", username, dbName))
		return nil
	}

	// Step 2: Drop the user
	logger.Info(fmt.Sprintf("Dropping user %s from db %s", username, dbName))
	err = client.Database(dbName).RunCommand(ctx, bson.D{
		{Key: "dropUser", Value: username},
	}).Err()

	if err != nil {
		return fmt.Errorf("failed to drop user %s: %w", username, err)
	}

	logger.Info(fmt.Sprintf("User %s dropped successfully from db %s", username, dbName))
	return nil
}

func (r *MongoServiceImpl) DropDatabase(ctx context.Context, dbName string) error {
	logger := utils.AddLoggerContext(r.logger, ctx)
	logger.Info(fmt.Sprintf("Dropping database %s", dbName))
	client, err := GetMongoClient(r.configuration)
	if err != nil {
		return err
	}
	return client.Database(dbName).Drop(context.TODO())

}

func (r *MongoServiceImpl) ListDatabases(ctx context.Context) ([]string, error) {
	logger := utils.AddLoggerContext(r.logger, ctx)
	logger.Info("ListDatabases databases")
	client, err := GetMongoClient(r.configuration)
	if err != nil {
		return nil, err
	}
	return client.ListDatabaseNames(context.TODO(), bson.D{{}})

}

func (r *MongoServiceImpl) GetOne(ctx context.Context, dbName string, collection string) (*mongo.SingleResult, error) {
	client, err := GetMongoClient(r.configuration)
	if err != nil {
		return nil, err
	}
	return client.Database(dbName).Collection(collection).FindOne(ctx, bson.M{}), nil
}

func (r *MongoServiceImpl) CreateCollection(ctx context.Context, dbName string, collection string) error {
	client, err := GetMongoClient(r.configuration)
	if err != nil {
		return err
	}

	return client.Database(dbName).CreateCollection(ctx, collection)
}

func (r *MongoServiceImpl) MovePrimary(ctx context.Context, dbName string, targetShard string) error {
	client, err := GetMongoClient(r.configuration)
	if err != nil {
		return err
	}

	adminDB := client.Database("admin")
	var dbInfo struct {
		Primary string `bson:"primary"`
		Db      string `bson:"db"`
	}

	err = client.Database("config").
		Collection("databases").
		FindOne(ctx, bson.M{"_id": dbName}).
		Decode(&dbInfo)

	if err == mongo.ErrNoDocuments {
		dbInfo.Primary = ""
	} else if err != nil {
		return fmt.Errorf("failed to read config.databases: %w", err)
	}

	if dbInfo.Primary == targetShard {
		return nil
	}

	cmd := bson.D{
		{Key: "movePrimary", Value: dbName},
		{Key: "to", Value: targetShard},
	}

	var result bson.M
	if err := adminDB.RunCommand(ctx, cmd).Decode(&result); err != nil {
		return fmt.Errorf("movePrimary failed: %w", err)
	}
	return nil
}

func (r *MongoServiceImpl) EnsureDBOnShard(ctx context.Context, dbName string, primaryName string) error {

	err := r.RunWithClusterGrants(ctx, dbName, func(service MongoService) error {
		err := service.MovePrimary(ctx, dbName, primaryName)
		if err != nil && !strings.Contains(err.Error(), "already enabled") {
			return fmt.Errorf("failed to move db [%s] on shard [%s] error : ", dbName, primaryName, err)
		}
		return nil
	})

	return err
}

func (r *MongoServiceImpl) IsValidShard(ctx context.Context, shardName string) (bool, error) {
	client, err := GetMongoClient(r.configuration)
	if err != nil {
		return false, err
	}

	cur, err := client.Database("config").Collection("shards").Find(ctx, bson.M{})
	if err != nil {
		return false, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var shard struct {
			ID string `bson:"_id"`
		}
		if err := cur.Decode(&shard); err != nil {
			return false, err
		}
		if shard.ID == shardName {
			return true, nil
		}
	}

	return false, nil
}

func (r *MongoServiceImpl) IsDatabaseAvailable(ctx context.Context, dbName string) (bool, error) {
	dbs, err := r.ListDatabases(ctx)
	if err != nil {
		return false, err
	}

	for _, name := range dbs {
		if name == dbName {
			// Database already exists → NOT valid to create
			return false, nil
		}
	}

	// Database does not exist → valid to create
	return true, nil
}

func (r *MongoServiceImpl) InsertOrUpdate(ctx context.Context, dbName string, collection string, data bson.D) (*mongo.UpdateResult, error) {
	client, err := GetMongoClient(r.configuration)
	if err != nil {
		return nil, err
	}
	opts := options.Update().SetUpsert(true)
	return client.Database(dbName).Collection(collection).UpdateMany(ctx, bson.M{}, bson.D{{"$set", data}}, opts)

}

func (r *MongoServiceImpl) InsertOne(ctx context.Context, dbName string, collection string, data bson.D) (*mongo.InsertOneResult, error) {
	client, err := GetMongoClient(r.configuration)
	if err != nil {
		return nil, err
	}
	coll := client.Database(dbName).Collection(collection)
	return coll.InsertOne(ctx, data)

}

func (r *MongoServiceImpl) GetRoles(ctx context.Context, dbName string) ([]string, error) {
	client, err := GetMongoClient(r.configuration)
	if err != nil {
		return nil, err
	}
	commandResult := client.Database(dbName).RunCommand(ctx, bson.D{{"usersInfo", 1}})

	var bRoles bson.M
	var roles []string
	err = commandResult.Decode(&bRoles)
	if err != nil {
		return nil, err
	}

	rolesMap := map[string]interface{}(bRoles)
	if paRolesMap, ok := rolesMap["users"].(primitive.A); ok {
		for _, dataMap := range []interface{}(paRolesMap) {
			dataMap2 := map[string]interface{}(dataMap.(primitive.M))
			roles = append(roles, dataMap2["user"].(string))
		}
	}
	return roles, commandResult.Err()

}

func (r *MongoServiceImpl) GetDbOwner(ctx context.Context, dbName string) (string, error) {
	client, err := GetMongoClient(r.configuration)
	if err != nil {
		return "", err
	}
	tries := 3
	for i := 0; i < tries; i++ {
		user, err := findDbOwner(ctx, client, dbName, dbName)
		if err != nil {
			user, err = findDbOwner(ctx, client, dbName, "admin")
			if err == nil {
				return user, nil
			}
		} else {
			return user, nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return "", errors.New("can't find user")

}

func findDbOwner(ctx context.Context, client *mongo.Client, dbName, authDb string) (string, error) {
	findCommand := bson.D{{"usersInfo", 1}, {"filter", bson.M{"roles": bson.M{"role": "dbOwner", "db": dbName}}}}
	commandResult := client.Database(authDb).RunCommand(context.TODO(), findCommand)
	var usersResponse bson.M
	if commandResult.Err() == nil {
		commandResult.Decode(&usersResponse)
		responseMap := map[string]interface{}(usersResponse)
		users := responseMap["users"].(bson.A)
		if len(users) != 0 {
			user := users[0].(bson.M)
			username := user["user"].(string)
			return username, nil
		} else {
			return "", mongo.ErrNoDocuments
		}
	}
	return "", commandResult.Err()
}

type MongoConfiguration interface {
	GetHost() string
	GetPort() int
	GetUser() string
	GetPassword() string
	GetAuthDb() string
	// GetMongoClient() (*mongo.Client, error)
	IsTLSEnabled() bool
}

type MongoConfigurationImpl struct {
	hostname string
	port     int
	user     string
	pass     string
	authDb   string
	client   *mongo.Client
}

var _ MongoConfiguration = &MongoConfigurationImpl{}

func (r *MongoConfigurationImpl) GetUser() string {
	return r.user
}

func (r *MongoConfigurationImpl) GetHost() string {
	return r.hostname
}

func (r *MongoConfigurationImpl) GetPort() int {
	return r.port
}

func (r *MongoConfigurationImpl) GetPassword() string {
	return r.pass
}

func (r *MongoConfigurationImpl) GetAuthDb() string {
	return r.authDb
}

func (r *MongoConfigurationImpl) IsTLSEnabled() bool {
	return utils.IsTLSEnabledForMainService()
}

func (r *MongoServiceImpl) EnableShardingAndCreateCollection(ctx context.Context, dbName string, settings *mUtils.Settings) error {
	err := r.RunWithClusterGrants(ctx, dbName, func(service MongoService) error {
		err := service.EnableSharding(ctx, dbName)
		if err != nil && !strings.Contains(err.Error(), "already enabled") {
			return fmt.Errorf("failed to enable sharding on db %s: %w", dbName, err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	for _, record := range settings.ShardingSettings {
		err = r.RunWithClusterGrants(ctx, dbName, func(service MongoService) error {
			err := service.CreateCollection(ctx, dbName, record.CollectionName)
			if err != nil && !strings.Contains(err.Error(), "already exists") {
				return fmt.Errorf("failed to create collection %s: %w", record.CollectionName, err)
			}
			return nil
		})
		if err != nil {
			return err
		}

		if err = r.ShardNonShardedCollection(ctx, dbName, record); err != nil {
			return err
		}
	}

	return nil
}

func (r *MongoServiceImpl) EnableSharding(ctx context.Context, dbName string) error {
	client, err := GetMongoClient(r.configuration)
	if err != nil {
		return err
	}

	admin := client.Database("admin")

	enableCmd := bson.D{
		{enableShardingCmd, dbName},
	}

	if err := admin.RunCommand(ctx, enableCmd).Err(); err != nil && !strings.Contains(err.Error(), "already enabled") {
		return fmt.Errorf("failed to enable sharding on db %s: %w", dbName, err)
	}

	return nil
}

func (r *MongoServiceImpl) ShardCollection(ctx context.Context, dbName string, settings mUtils.ShardingSettings) error {
	client, err := GetMongoClient(r.configuration)
	if err != nil {
		return err
	}

	admin := client.Database("admin")
	shardCmd := bson.D{
		{"shardCollection", fmt.Sprintf("%s.%s", dbName, settings.CollectionName)},
		{"key", bson.D{{settings.ShardKey, func() any {
			if settings.Strategy == "hashed" {
				return "hashed"
			}
			return 1
		}()}}},
	}

	if err := admin.RunCommand(ctx, shardCmd).Err(); err != nil && !strings.Contains(err.Error(), "already sharded") {
		return fmt.Errorf("failed to shard collection %s.%s: %w", dbName, settings.CollectionName, err)
	}

	return nil
}

// IsCollectionSharded checks config.collections for the current shard key of a namespace.
func (r *MongoServiceImpl) IsCollectionSharded(ctx context.Context, dbName, collectionName string) (bool, bson.D, error) {
	client, err := GetMongoClient(r.configuration)
	if err != nil {
		return false, nil, err
	}

	ns := fmt.Sprintf("%s.%s", dbName, collectionName)
	var result struct {
		Key bson.D `bson:"key"`
	}

	err = client.Database("config").Collection("collections").
		FindOne(ctx, bson.D{{"_id", ns}}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil, nil
		}
		return false, nil, fmt.Errorf("failed to check sharding status for %s: %w", ns, err)
	}
	return true, result.Key, nil
}

// shardKeyMatches compares an existing config.collections key to the requested shard settings.
func shardKeyMatches(currentKey bson.D, settings mUtils.ShardingSettings) bool {
	if len(currentKey) != 1 || currentKey[0].Key != settings.ShardKey {
		return false
	}
	if settings.Strategy == "hashed" {
		v, ok := currentKey[0].Value.(string)
		return ok && v == "hashed"
	}
	switch v := currentKey[0].Value.(type) {
	case int32:
		return v == 1
	case int64:
		return v == 1
	case float64:
		return v == 1
	}
	return false
}

// CreateShardKeyIndex creates the index required before sharding on the given key/strategy.
func (r *MongoServiceImpl) CreateShardKeyIndex(ctx context.Context, dbName string, settings mUtils.ShardingSettings) error {
	client, err := GetMongoClient(r.configuration)
	if err != nil {
		return err
	}

	var keyValue interface{} = 1
	if settings.Strategy == "hashed" {
		keyValue = "hashed"
	}

	_, err = client.Database(dbName).Collection(settings.CollectionName).Indexes().
		CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{settings.ShardKey, keyValue}}})
	if err != nil {
		return fmt.Errorf("failed to create index on %s.%s for shard key %s: %w",
			dbName, settings.CollectionName, settings.ShardKey, err)
	}
	return nil
}

// ShardNonShardedCollection creates the required index and shards the collection, unless it's
// already sharded with the requested key (no-op) or sharded with a different key (error).
func (r *MongoServiceImpl) ShardNonShardedCollection(ctx context.Context, dbName string, settings mUtils.ShardingSettings) error {
	logger := utils.AddLoggerContext(r.logger, ctx)

	sharded, currentKey, err := r.IsCollectionSharded(ctx, dbName, settings.CollectionName)
	if err != nil {
		return err
	}
	if sharded {
		if shardKeyMatches(currentKey, settings) {
			logger.Info(fmt.Sprintf("collection %s.%s already sharded with key %s (%s), skipping",
				dbName, settings.CollectionName, settings.ShardKey, settings.Strategy))
			return nil
		}
		return fmt.Errorf("collection %s.%s is already sharded with a different key %v",
			dbName, settings.CollectionName, currentKey)
	}

	err = r.RunWithClusterGrants(ctx, dbName, func(service MongoService) error {
		if err := r.CreateShardKeyIndex(ctx, dbName, settings); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	logger.Debug(fmt.Sprintf("sharding collection %s.%s on key %s (%s)",
		dbName, settings.CollectionName, settings.ShardKey, settings.Strategy))

	return r.RunWithClusterGrants(ctx, dbName, func(service MongoService) error {
		if err := service.ShardCollection(ctx, dbName, settings); err != nil {
			return fmt.Errorf("failed to shard collection %s: %w", settings.CollectionName, err)
		}
		return nil
	})
}
