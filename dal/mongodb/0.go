package mongodb

import (
	"context"

	"github.com/Lukiya/logs/core"
	"github.com/syncfuture/go/u"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	_LOG_DB_PREFIX = "LOG_"
	_ENTRY_TABLE   = "entries"
	_CLIENT_DB     = "LogClients"
	_CLIENT_TABLE  = "clients"
)

var (
	_client     *mongo.Client
	clientTable *mongo.Collection
)

func init() {
	connStr := core.MainCP.GetString("ConnectionStrings.Mongo")

	// Create a new client and connect to the server
	var err error
	_client, err = mongo.Connect(context.Background(), options.Client().ApplyURI(connStr))
	u.LogFatal(err)

	clientTable = _client.Database(_CLIENT_DB).Collection(_CLIENT_TABLE)
	unique := true

	_, err = clientTable.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: &options.IndexOptions{Unique: &unique},
		},
	})
	u.LogFatal(err)
}

func getTable(dbName string) *mongo.Collection {
	return _client.Database(dbName).Collection(_ENTRY_TABLE)
}
