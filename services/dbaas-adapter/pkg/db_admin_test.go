package pkg

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/Netcracker/qubership-dbaas-adapter-core/pkg/dao"
	"github.com/Netcracker/qubership-dbaas-adapter-core/pkg/utils"
	mUtils "github.com/Netcracker/qubership-dbaas-mongo/utils"
	"github.com/docker/distribution/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type MongoConfigurationTest struct {
	mock.Mock
}

var _ MongoConfiguration = &MongoConfigurationTest{}

func (r *MongoConfigurationTest) GetUser() string {
	args := r.Called()
	return args.String(0)
}

func (r *MongoConfigurationTest) GetHost() string {
	args := r.Called()
	return args.String(0)
}

func (r *MongoConfigurationTest) GetPort() int {
	args := r.Called()
	return args.Int(0)
}

func (r *MongoConfigurationTest) GetPassword() string {
	args := r.Called()
	return args.String(0)
}

func (r *MongoConfigurationTest) GetAuthDb() string {
	args := r.Called()
	return args.String(0)
}

func (r *MongoConfigurationTest) GetMongoClient() (*mongo.Client, error) {
	args := r.Called()
	return args.Get(0).(*mongo.Client), args.Error(1)
}

func (r *MongoConfigurationTest) IsTLSEnabled() bool {
	args := r.Called()
	return args.Bool(0)
}

// MongoServiceTest is a mocked object that implements an interface
// that describes an object that the code I am testing relies on.
type MongoServiceTest struct {
	mock.Mock
	configuration MongoConfiguration
}

func (m *MongoServiceTest) RunWithGrants(ctx context.Context, dbName string, ff func(service MongoService) error) (err error) {
	result, err := m.GrantRole(context.Background(), dbName)
	if err != nil {
		return err
	} else if result != nil && result.Err() != nil {
		return result.Err()
	}

	defer func() {
		result, err2 := m.RevokeRole(context.Background(), dbName)
		if err2 != nil {
			err = err2
		} else if result != nil && result.Err() != nil {
			err = result.Err()
		}
	}()

	return ff(m)
}

func (m *MongoServiceTest) RunWithClusterGrants(ctx context.Context, dbName string, ff func(service MongoService) error) (err error) {
	result, err := m.GrantRole(context.Background(), dbName)
	if err != nil {
		return err
	} else if result != nil && result.Err() != nil {
		return result.Err()
	}

	defer func() {
		result, err2 := m.RevokeRole(context.Background(), dbName)
		if err2 != nil {
			err = err2
		} else if result != nil && result.Err() != nil {
			err = result.Err()
		}
	}()

	return ff(m)
}

func (m *MongoServiceTest) GrantRole(ctx context.Context, dbName string) (*mongo.SingleResult, error) {
	args := m.Called(dbName)
	return nil, args.Error(1)
}

func (m *MongoServiceTest) RevokeRole(ctx context.Context, dbName string) (*mongo.SingleResult, error) {
	args := m.Called(dbName)
	return nil, args.Error(1)
}

func (m *MongoServiceTest) CreateOrUpdateUser(ctx context.Context, username, pass, database, authDb string, role string, force bool) error {
	args := m.Called(username, pass, database, authDb, role)
	return args.Error(0)
}

func (m *MongoServiceTest) UpdateUser(ctx context.Context, username string, pass string, database string, authDb string, role string) error {
	args := m.Called(username, pass, database, authDb, role)
	return args.Error(0)
}

func (m *MongoServiceTest) IsUserExist(ctx context.Context, conf MongoConfiguration, username string, dbName string) (bool, error) {
	args := m.Called(username)
	return args.Bool(0), args.Error(1)
}

func (m *MongoServiceTest) UpdateUserPassword(ctx context.Context, username string, password string, dbName string) error {
	args := m.Called(username, password)
	return args.Error(0)
}

func (m *MongoServiceTest) DropUser(ctx context.Context, username string) error {
	args := m.Called(username)
	return args.Error(0)
}

func (m *MongoServiceTest) DropDatabase(ctx context.Context, dbName string) error {
	args := m.Called(dbName)
	return args.Error(0)
}

func (m *MongoServiceTest) ListDatabases(ctx context.Context) ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *MongoServiceTest) GetDbOwner(ctx context.Context, dbName string) (string, error) {
	args := m.Called(dbName)
	return args.Get(0).(string), args.Error(1)
}

func (m *MongoServiceTest) GetRoles(ctx context.Context, dbName string) ([]string, error) {
	args := m.Called(dbName)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MongoServiceTest) InsertOrUpdate(ctx context.Context, dbName string, collection string, data bson.D) (*mongo.UpdateResult, error) {
	args := m.Called(dbName, collection, data)
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

func (m *MongoServiceTest) InsertOne(ctx context.Context, dbName string, collection string, data bson.D) (*mongo.InsertOneResult, error) {
	args := m.Called(dbName, collection, data)
	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}

func (m *MongoServiceTest) GetOne(ctx context.Context, dbName string, collection string) (*mongo.SingleResult, error) {
	args := m.Called(dbName, collection)
	return args.Get(0).(*mongo.SingleResult), args.Error(1)
}

func (m *MongoServiceTest) GetConfiguration() MongoConfiguration {
	return m.configuration
}

func (m *MongoServiceTest) EnableShardingAndCreateCollection(
	ctx context.Context,
	dbName string,
	settings *mUtils.Settings,
) error {
	return nil // Always success for tests
}

func (m *MongoServiceTest) CreateCollection(ctx context.Context, dbName string, collection string) error {
	return nil // Always success for tests
}

func (m *MongoServiceTest) EnsureDBOnShard(ctx context.Context, dbName string, primaryShard string) error {
	return nil // Always success for tests
}

func (m *MongoServiceTest) EnableSharding(ctx context.Context, dbName string) error {
	return nil // Always success for tests
}
func (m *MongoServiceTest) ShardCollection(ctx context.Context, dbName string, settings mUtils.ShardingSettings) error {
	return nil // Always success for tests
}

func (m *MongoServiceTest) IsValidShard(ctx context.Context, shardName string) (bool, error) {
	return true, nil // Always success for tests
}

func (m *MongoServiceTest) MovePrimary(ctx context.Context, dbName string, targetShard string) error {
	return nil // Always success for tests
}

func (m *MongoServiceTest) IsDatabaseAvailable(ctx context.Context, dbName string) (bool, error) {
	return true, nil // Always success for tests
}

func (m *MongoServiceTest) GetReplicaSetHosts() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func TestDropUserSuccessfull(t *testing.T) {
	config := new(MongoConfigurationTest)
	config.On("GetMongoClient").Return(nil, nil)

	service := &MongoServiceTest{configuration: config}

	service.On("DropUser", mock.Anything).Return(nil)
	dbAdmin := &MongoDbAdministration{mongodService: service, logger: utils.GetLogger(true)}

	dropped := dbAdmin.DropResources(context.Background(), []dao.DbResource{{Kind: "user", Name: "any"}})

	// assert that the expectations were met
	service.AssertExpectations(t)

	assert.True(t, dropped[0].Status == dao.DELETED)
}

func TestDropUserError(t *testing.T) {
	config := new(MongoConfigurationTest)
	config.On("GetMongoClient").Return(nil, nil)

	service := &MongoServiceTest{configuration: config}
	service.On("DropUser", mock.Anything).Return(errors.New("error droping user"))
	dbAdmin := &MongoDbAdministration{mongodService: service, logger: utils.GetLogger(true)}

	dropped := dbAdmin.DropResources(context.Background(), []dao.DbResource{{Kind: "user", Name: "db:any"}})

	service.AssertExpectations(t)

	assert.True(t, dropped[0].Status == dao.DELETE_FAILED)
	assert.True(t, dropped[0].ErrorMessage == "error droping user")
}

func TestDropDbError(t *testing.T) {
	config := new(MongoConfigurationTest)
	config.On("GetMongoClient").Return(nil, nil)

	service := &MongoServiceTest{configuration: config}
	service.On("DropDatabase", mock.Anything).Return(errors.New("error droping db"))
	service.On("GrantRole", mock.Anything).Return(&mongo.SingleResult{}, nil)
	service.On("RevokeRole", mock.Anything).Return(&mongo.SingleResult{}, nil)

	dbAdmin := &MongoDbAdministration{mongodService: service, logger: utils.GetLogger(true)}
	dropped := dbAdmin.DropResources(context.Background(), []dao.DbResource{{Kind: "database", Name: "any"}})

	service.AssertExpectations(t)

	assert.True(t, dropped[0].Status == dao.DELETE_FAILED)
	assert.True(t, dropped[0].ErrorMessage == "error droping db")
}

func TestListDatabases(t *testing.T) {
	config := new(MongoConfigurationTest)
	config.On("GetMongoClient").Return(nil, nil)

	service := &MongoServiceTest{configuration: config}
	service.On("ListDatabases").Return([]string{"foo", "bar"}, nil)

	dbAdmin := &MongoDbAdministration{mongodService: service, logger: utils.GetLogger(true)}
	dbs := dbAdmin.GetDatabases(context.Background())

	service.AssertExpectations(t)

	assert.Equal(t, dbs, []string{"foo", "bar"})
}

func TestListDatabasesPanic(t *testing.T) {
	config := new(MongoConfigurationTest)
	config.On("GetMongoClient").Return(nil, nil)

	service := &MongoServiceTest{configuration: config}
	service.On("ListDatabases").Return([]string{"foo", "bar"}, errors.New("failed to get dbs"))

	defer func() { recover() }()
	dbAdmin := &MongoDbAdministration{mongodService: service, logger: utils.GetLogger(true)}
	dbAdmin.GetDatabases(context.Background())
	t.Errorf("Did not panic")
}

func TestMongoDbAdministration_DescribeDatabases(t *testing.T) {
	dbName := generateUUID()
	config := new(MongoConfigurationTest)
	config.On("GetMongoClient").Return(nil, nil)
	config.On("GetHost").Return("localhost")
	config.On("GetPort").Return(27017)
	config.On("IsTLSEnabled").Return(false)

	service := &MongoServiceTest{configuration: config}
	service.On("GetRoles", dbName).Return(make([]string, 0), nil)
	service.On("GetDbOwner", dbName).Return("bar", nil)

	dbAdmin := &MongoDbAdministration{mongodService: service, logger: utils.GetLogger(true)}
	dbAdmin.DescribeDatabases(context.Background(), []string{dbName}, false, false)
}

func generateUUID() string {
	return uuid.Generate().String()
}

func returnIfTrue(condition bool, returnValue string) string {
	if condition {
		return returnValue
	}
	return ""
}

func generateRequestOnCreateDb(pref bool, meta bool, settings bool) dao.DbCreateRequest {
	prefix := returnIfTrue(pref, "foo")
	return dao.DbCreateRequest{
		NamePrefix: &prefix,
		DbName:     generateUUID(),
		Username:   generateUUID(),
		Password:   generateUUID(),
		Metadata:   map[string]interface{}{returnIfTrue(meta, generateUUID()): returnIfTrue(meta, generateUUID())},
		Settings:   map[string]interface{}{returnIfTrue(settings, generateUUID()): returnIfTrue(settings, generateUUID())},
	}
}

func TestMongoDbAdministration_CreateDatabase_BadNames(t *testing.T) {
	config := new(MongoConfigurationTest)
	config.On("GetMongoClient").Return(nil, nil)
	config.On("GetHost").Return("mongos.mongo-cluster.svc")
	config.On("GetPort").Return(27017)

	randomRequestCreateDb := generateRequestOnCreateDb(true, true, true)
	randomRequestCreateDb.Username += "/"
	randomRequestCreateDb.DbName += "@"

	service := &MongoServiceTest{configuration: config}
	service.On("CreateOrUpdateUser", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	dbAdmin := &MongoDbAdministration{mongodService: service, logger: utils.GetLogger(true)}
	_, _, err := dbAdmin.CreateDatabase(context.Background(), randomRequestCreateDb)

	assert.NotNil(t, err)
}

func TestMongoDbAdministration_CreateUserNotExists(t *testing.T) {
	config := new(MongoConfigurationTest)
	config.On("GetMongoClient").Return(nil, nil)
	config.On("GetHost").Return("mongos.mongo-cluster.svc")
	config.On("GetPort").Return(27017)
	config.On("IsTLSEnabled").Return(false)

	service := &MongoServiceTest{configuration: config}
	service.On("CreateOrUpdateUser", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	service.On("IsUserExist", mock.Anything, mock.Anything).Return(false, nil)

	dbAdmin := &MongoDbAdministration{mongodService: service, logger: utils.GetLogger(true)}
	userResponse, err := dbAdmin.CreateUser(context.Background(), "", dao.UserCreateRequest{
		DbName:   generateUUID(),
		Password: generateUUID(),
	})

	url := userResponse.ConnectionProperties["url"]
	assert.Nil(t, err)
	assert.NotNil(t, userResponse)
	assert.NotEqual(t, url, "")
}

func getAllParamsDescriptionNoVault(randomRequestCreateDb dao.DbCreateRequest) *dao.LogicalDatabaseDescribed {
	var connectionProperties []dao.ConnectionProperties
	connectionProperties = append(connectionProperties, dao.ConnectionProperties{
		"dbName":     *randomRequestCreateDb.NamePrefix + "-" + randomRequestCreateDb.DbName,
		"authDbName": *randomRequestCreateDb.NamePrefix + "-" + randomRequestCreateDb.DbName,
		"password":   randomRequestCreateDb.Password,
		"url":        fmt.Sprintf("mongodb://%s:27017/", serverHostname) + *randomRequestCreateDb.NamePrefix + "-" + randomRequestCreateDb.DbName,
		"username":   randomRequestCreateDb.Username,
	})
	return &dao.LogicalDatabaseDescribed{
		ConnectionProperties: connectionProperties,
		Resources: []dao.DbResource{
			{
				Kind: "user",
				Name: *randomRequestCreateDb.NamePrefix + "-" + randomRequestCreateDb.DbName + ":" + randomRequestCreateDb.Username,
			},
			{
				Kind: "database",
				Name: *randomRequestCreateDb.NamePrefix + "-" + randomRequestCreateDb.DbName,
			},
		},
	}
}

func getAllParamsDescriptionWithVault(randomRequestCreateDb dao.DbCreateRequest, authDB string) *dao.LogicalDatabaseDescribed {
	var connectionProperties []dao.ConnectionProperties
	connectionProperties = append(connectionProperties, dao.ConnectionProperties{
		"dbName":     *randomRequestCreateDb.NamePrefix + "-" + randomRequestCreateDb.DbName,
		"authDbName": authDB,
		"password":   randomRequestCreateDb.Password,
		"url":        fmt.Sprintf("mongodb://%s:27017/", serverHostname) + authDB,
		"username":   randomRequestCreateDb.Username,
	})
	return &dao.LogicalDatabaseDescribed{
		ConnectionProperties: connectionProperties,
		Resources: []dao.DbResource{
			{
				Kind: "user",
				Name: authDB + ":" + randomRequestCreateDb.Username,
			},
			{
				Kind: "database",
				Name: *randomRequestCreateDb.NamePrefix + "-" + randomRequestCreateDb.DbName,
			},
		},
	}
}

func getAllParamsDescriptionNoPrefix(randomRequestCreateDb dao.DbCreateRequest) *dao.LogicalDatabaseDescribed {
	var connectionProperties []dao.ConnectionProperties
	connectionProperties = append(connectionProperties, dao.ConnectionProperties{
		"authDbName": randomRequestCreateDb.DbName,
		"dbName":     randomRequestCreateDb.DbName,
		"password":   randomRequestCreateDb.Password,
		"url":        fmt.Sprintf("mongodb://%s:27017/", serverHostname) + randomRequestCreateDb.DbName,
		"username":   randomRequestCreateDb.Username,
	})
	return &dao.LogicalDatabaseDescribed{
		ConnectionProperties: connectionProperties,
		Resources: []dao.DbResource{
			{
				Kind: "user",
				Name: randomRequestCreateDb.DbName + ":" + randomRequestCreateDb.Username,
			},
			{
				Kind: "database",
				Name: randomRequestCreateDb.DbName,
			},
		},
	}
}

func getAllParamsDescriptionPrefixOnly(randomRequestCreateDb dao.DbCreateRequest) *dao.LogicalDatabaseDescribed {
	var connectionProperties []dao.ConnectionProperties
	connectionProperties = append(connectionProperties, dao.ConnectionProperties{
		"authDbName": *randomRequestCreateDb.NamePrefix + "-" + randomRequestCreateDb.DbName,
		"dbName":     *randomRequestCreateDb.NamePrefix + "-" + randomRequestCreateDb.DbName,
		"password":   randomRequestCreateDb.Password,
		"url":        fmt.Sprintf("mongodb://%s:27017/", serverHostname) + *randomRequestCreateDb.NamePrefix + "-" + randomRequestCreateDb.DbName,
		"username":   randomRequestCreateDb.Username,
	})
	return &dao.LogicalDatabaseDescribed{
		ConnectionProperties: connectionProperties,
		Resources: []dao.DbResource{
			{
				Kind: "user",
				Name: *randomRequestCreateDb.NamePrefix + "-" + randomRequestCreateDb.DbName + ":" + randomRequestCreateDb.Username,
			},
			{
				Kind: "database",
				Name: *randomRequestCreateDb.NamePrefix + "-" + randomRequestCreateDb.DbName,
			},
		},
	}
}

func TestMongoDbAdministration_CreateDatabase(t *testing.T) {
	authDB := uuid.Generate().String()
	config := new(MongoConfigurationTest)
	config.On("GetMongoClient").Return(nil, nil)
	config.On("GetHost").Return(serverHostname)
	config.On("GetPort").Return(27017)
	config.On("GetAuthDb").Return(authDB)
	config.On("IsTLSEnabled").Return(false)

	service := &MongoServiceTest{configuration: config}
	service.On("CreateOrUpdateUser", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	service.On("InsertOne", mock.Anything, mock.Anything, mock.Anything).Return(&mongo.InsertOneResult{}, nil)
	service.On("GrantRole", mock.Anything).Return(&mongo.SingleResult{}, nil)
	service.On("RevokeRole", mock.Anything).Return(&mongo.SingleResult{}, nil)

	type fields struct {
		logger        *zap.Logger
		mongodService MongoService
		vaultEnabled  bool
	}
	type args struct {
		requestOnCreateDb dao.DbCreateRequest
	}
	tests := []struct {
		name              string
		fields            fields
		args              args
		wantDbName        func(randomRequestCreateDb dao.DbCreateRequest) string
		wantDbDescription func(randomRequestCreateDb dao.DbCreateRequest) *dao.LogicalDatabaseDescribed
		wantErr           bool
	}{
		{
			name: "All params no vault",
			fields: fields{
				mongodService: service,
				vaultEnabled:  false,
				logger:        utils.GetLogger(true),
			},
			args: args{
				requestOnCreateDb: generateRequestOnCreateDb(true, true, true),
			},
			wantDbName: func(randomRequestCreateDb dao.DbCreateRequest) string {
				return *randomRequestCreateDb.NamePrefix + "-" + randomRequestCreateDb.DbName
			},
			wantDbDescription: func(randomRequestCreateDb dao.DbCreateRequest) *dao.LogicalDatabaseDescribed {
				return getAllParamsDescriptionNoVault(randomRequestCreateDb)
			},
		},
		{
			name: "All params with vault",
			fields: fields{
				mongodService: service,
				vaultEnabled:  true,
				logger:        utils.GetLogger(true),
			},
			args: args{
				requestOnCreateDb: generateRequestOnCreateDb(true, true, true),
			},
			wantDbName: func(randomRequestCreateDb dao.DbCreateRequest) string {
				return *randomRequestCreateDb.NamePrefix + "-" + randomRequestCreateDb.DbName
			},
			wantDbDescription: func(randomRequestCreateDb dao.DbCreateRequest) *dao.LogicalDatabaseDescribed {
				return getAllParamsDescriptionWithVault(randomRequestCreateDb, authDB)
			},
		},
		{
			name: "No prefix",
			fields: fields{
				mongodService: service,
				vaultEnabled:  false,
				logger:        utils.GetLogger(true),
			},
			args: args{
				requestOnCreateDb: generateRequestOnCreateDb(false, true, true),
			},
			wantDbName: func(randomRequestCreateDb dao.DbCreateRequest) string {
				return randomRequestCreateDb.DbName
			},
			wantDbDescription: func(randomRequestCreateDb dao.DbCreateRequest) *dao.LogicalDatabaseDescribed {
				return getAllParamsDescriptionNoPrefix(randomRequestCreateDb)
			},
		},
		{
			name: "Prefix only",
			fields: fields{
				mongodService: service,
				vaultEnabled:  false,
				logger:        utils.GetLogger(true),
			},
			args: args{
				requestOnCreateDb: generateRequestOnCreateDb(true, false, false),
			},
			wantDbName: func(randomRequestCreateDb dao.DbCreateRequest) string {
				return *randomRequestCreateDb.NamePrefix + "-" + randomRequestCreateDb.DbName
			},
			wantDbDescription: func(randomRequestCreateDb dao.DbCreateRequest) *dao.LogicalDatabaseDescribed {
				return getAllParamsDescriptionPrefixOnly(randomRequestCreateDb)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &MongoDbAdministration{
				logger:        tt.fields.logger,
				mongodService: tt.fields.mongodService,
			}
			got, got1, err := c.CreateDatabase(context.Background(), tt.args.requestOnCreateDb)
			if (err != nil) != tt.wantErr {
				t.Errorf("MongoDbAdministration.CreateDatabase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantDbName(tt.args.requestOnCreateDb) {
				t.Errorf("MongoDbAdministration.CreateDatabase() got = %v, want %v", got, tt.wantDbName(tt.args.requestOnCreateDb))
			}
			if !reflect.DeepEqual(got1, tt.wantDbDescription(tt.args.requestOnCreateDb)) {
				t.Errorf("MongoDbAdministration.CreateDatabase() got1 = %v, want %v", got1, tt.wantDbDescription(tt.args.requestOnCreateDb))
			}
		})
	}
}
