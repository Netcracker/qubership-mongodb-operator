package pkg

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"sync"

	mUtils "github.com/Netcracker/qubership-dbaas-mongo/utils"
	"github.com/Netcracker/qubership-dbaas-adapter-core/pkg/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoClients struct {
	configuration MongoConfiguration
	client        *mongo.Client
}

var mu sync.Mutex
var clients []MongoClients

func GetMongoClient(conf MongoConfiguration) (*mongo.Client, error) {
	mu.Lock()
	logger := utils.GetLogger(true)
	defer func() {
		logger.Debug(fmt.Sprintf("active connections: %d", len(clients)))
		mu.Unlock()
	}()

	for _, client := range clients {
		if reflect.DeepEqual(client.configuration, conf) {
			return client.client, nil
		}
	}
	client, err := createMongoClient(conf)
	if err != nil {
		return nil, err
	}

	clients = append(clients, MongoClients{configuration: conf, client: client})

	return client, err
}

func DisconnectMongoClient(conf MongoConfiguration) error {
	mu.Lock()
	defer mu.Unlock()
	index := -1
	for i, client := range clients {
		if reflect.DeepEqual(client.configuration, conf) {
			client.client.Disconnect(context.TODO())
			index = i
		}
	}

	if index != -1 {
		newClients := make([]MongoClients, 0)
		newClients = append(newClients, clients[:index]...)
		clients = append(newClients, clients[index+1:]...)
	}

	return nil
}

func createMongoClient(conf MongoConfiguration) (*mongo.Client, error) {
	var mainURI string
	var URIOptions string

	if conf.GetAuthDb() != "" {
		mainURI = fmt.Sprintf("mongodb://%s:%s@%s/", url.QueryEscape(conf.GetUser()), url.QueryEscape(conf.GetPassword()), conf.GetHost())
		URIOptions = fmt.Sprintf("?authSource=%s", conf.GetAuthDb())
	} else {
		mainURI = fmt.Sprintf("mongodb://%s/", conf.GetHost())
	}

	if conf.IsTLSEnabled() {
		tlsOptions := fmt.Sprintf("tls=true&tlsCAFile=%s", mUtils.GetCACert())

		// remove before using production certs
		tlsOptions = fmt.Sprint(tlsOptions, "&tlsInsecure=true")

		if len(URIOptions) > 0 {
			URIOptions = fmt.Sprint(URIOptions, "&", tlsOptions)
		} else {
			URIOptions = fmt.Sprint("?", tlsOptions)
		}
	}

	return mongo.Connect(context.TODO(), options.Client().ApplyURI(mainURI+URIOptions))
}
