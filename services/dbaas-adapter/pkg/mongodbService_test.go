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

package pkg

import (
	"context"
	"fmt"
	"testing"

	"github.com/Netcracker/qubership-dbaas-adapter-core/pkg/utils"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var logger = utils.GetLogger(utils.GetEnvAsBool("DEBUG_LOG", true))
var serverHostname = utils.GetEnv("TEST_MONGO_HOST", "10.236.155.207")

func prepareMongoDB() {
	conf := MongoConfigurationImpl{hostname: serverHostname}
	ms := MongoServiceImpl{configuration: &conf, logger: logger}
	if exists, _ := ms.IsUserExist(context.Background(), ms.configuration, "dbaas", "admin"); !exists {
		client, err := GetMongoClient(&conf)
		if err != nil {
			panic(err)
		}
		commandResult := client.Database("admin").RunCommand(context.TODO(),
			bson.D{primitive.E{Key: "createUser", Value: "dbaas"},
				primitive.E{Key: "pwd", Value: "dbaas"},
				primitive.E{Key: "roles", Value: []bson.M{
					{"role": "clusterMonitor", "db": "admin"},
					{"role": "userAdminAnyDatabase", "db": "admin"},
					{"role": "readAnyDatabase", "db": "admin"},
				}},
			})

		if commandResult.Err() != nil {
			panic(commandResult.Err())
		}
	}
}

func TestMongoServiceImpl_ListDatabases(t *testing.T) {
	prepareMongoDB()
	conf := MongoConfigurationImpl{hostname: serverHostname, authDb: "admin", user: "dbaas", pass: "dbaas"}
	ms := MongoServiceImpl{configuration: &conf, logger: logger}

	dbs, err := ms.ListDatabases(context.Background())
	assert.NotEmpty(t, dbs)
	assert.NoError(t, err)
}

func TestMongoServiceImpl_Grant_role(t *testing.T) {
	prepareMongoDB()
	conf := MongoConfigurationImpl{hostname: serverHostname, authDb: "admin", user: "dbaas", pass: "dbaas"}
	ms := MongoServiceImpl{configuration: &conf, logger: logger}

	_, err := ms.GrantRole(context.Background(), "foo", dbOwner)
	assert.NoError(t, err)

	_, err = ms.RevokeRole(context.Background(), "foo", dbOwner)
	assert.NoError(t, err)
}

func TestMongoServiceImpl_DropDatabase(t *testing.T) {
	prepareMongoDB()
	conf := MongoConfigurationImpl{hostname: serverHostname, authDb: "admin", user: "dbaas", pass: "dbaas"}
	ms := MongoServiceImpl{configuration: &conf, logger: logger}

	dbName := "testDbaas"
	_, err := ms.InsertOne(context.Background(), dbName, "test", bson.D{{"foo", "bar"}})
	assert.NoError(t, err)

	dbs, listErr := ms.ListDatabases(context.Background())
	assert.NoError(t, listErr)
	assert.Contains(t, dbs, dbName)

	err = ms.DropDatabase(context.Background(), dbName)
	assert.NoError(t, err)

	dbs, listErr = ms.ListDatabases(context.Background())
	assert.NoError(t, listErr)
	assert.NotContains(t, dbs, dbName)
}

func TestMongoServiceImpl_CreateUser(t *testing.T) {
	prepareMongoDB()
	authDB := "admin"

	conf := MongoConfigurationImpl{hostname: serverHostname, authDb: authDB, user: "dbaas", pass: "dbaas"}
	ms := MongoServiceImpl{configuration: &conf, logger: logger}

	username := generateUUID()
	pass := generateUUID()
	newPass := generateUUID()

	//create user
	err := ms.CreateOrUpdateUser(context.Background(), username, pass, "foo", authDB, "", false)
	assert.NoError(t, err)

	//check user's created
	exists, err := ms.IsUserExist(context.Background(), ms.configuration, username, authDB)
	assert.NoError(t, err)
	assert.True(t, exists)

	//check user can login and execute
	assert.NoError(t, checkUserLogin(authDB, username, pass))

	//change user pass
	err = ms.UpdateUserPassword(context.Background(), username, newPass, authDB)
	assert.NoError(t, err)

	//check user can login and execute
	assert.NoError(t, checkUserLogin(authDB, username, newPass))

	//delete user
	err = ms.DropUser(context.Background(), fmt.Sprintf("%s:%s", authDB, username))
	assert.NoError(t, err)

	//check deleted
	exists, err = ms.IsUserExist(context.Background(), ms.configuration, username, authDB)
	assert.NoError(t, err)
	assert.False(t, exists)

	//check user cannot execute
	assert.Error(t, checkUserLogin(authDB, username, newPass))
}

func checkUserLogin(authDB, username, pass string) error {
	newUserConf := MongoConfigurationImpl{hostname: serverHostname, authDb: authDB, user: username, pass: pass}
	newUserMs := MongoServiceImpl{configuration: &newUserConf, logger: logger}
	_, err := newUserMs.ListDatabases(context.Background())
	DisconnectMongoClient(&newUserConf)
	return err
}
