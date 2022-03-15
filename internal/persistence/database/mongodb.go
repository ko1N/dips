package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBConfig - config entry describing a database config
type MongoDBConfig struct {
	Hosts         []string `json:"hosts" toml:"hosts"`
	AuthMechanism string   `json:"auth_mechanism" toml:"auth_mechanism"`
	AuthSource    string   `json:"auth_source" toml:"auth_source"`
	Username      string   `json:"username" toml:"username"`
	Password      string   `json:"password" toml:"password"`
	Database      string   `json:"database" toml:"database"`
}

// MongoDBConnect - opens a connection to mongodb and returns the connection object
func MongoDBConnect(conf *MongoDBConfig) (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	credential := options.Credential{
		AuthMechanism: conf.AuthMechanism,
		AuthSource:    conf.AuthSource,
		Username:      conf.Username,
		Password:      conf.Password,
	}
	clientOpts := options.
		Client().
		ApplyURI(conf.Hosts[0]).
		SetAuth(credential)

	mongo, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}
	return mongo.Database(conf.Database), nil
}
