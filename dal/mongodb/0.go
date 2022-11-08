package mongodb

import (
	"context"

	"github.com/Lukiya/logs/core"
	"github.com/syncfuture/go/stask"
	"github.com/syncfuture/go/u"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	_CLIENT_DB    = "LogClients"
	_CLIENT_TABLE = "clients"
)

var (
	_client      *mongo.Client
	_parallel    = stask.NewParallel()
	_clientTable *mongo.Collection
)

func init() {
	connStr := core.GrpcCP.GetString("ConnectionStrings.Mongo")
	ctx := context.Background()
	// Create a new client and connect to the server
	var err error
	_client, err = mongo.Connect(ctx, options.Client().ApplyURI(connStr))
	u.LogFatal(err)

	_clientTable = _client.Database(_CLIENT_DB).Collection(_CLIENT_TABLE)
	unique := true

	_, err = _clientTable.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: &options.IndexOptions{Unique: &unique},
		},
	})
	u.LogFatal(err)
}
