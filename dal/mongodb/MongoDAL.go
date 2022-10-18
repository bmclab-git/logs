package mongodb

import (
	"context"

	"github.com/Lukiya/logs/model"
	"github.com/syncfuture/go/serr"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var nameOnly = true

type MongoDAL struct {
}

// ************************************************************************************************

func (self *MongoDAL) InsertClient(client *model.LogClient) error {
	_, err := clientTable.InsertOne(context.Background(), client)
	if err != nil {
		return serr.WithStack(err)
	}
	return nil
}
func (self *MongoDAL) GetClient(id string) (r *model.LogClient, err error) {
	rs := clientTable.FindOne(context.Background(), bson.M{"id": id})
	err = rs.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, serr.WithStack(err)
	}

	err = rs.Decode(&r)
	if err != nil {
		return nil, serr.WithStack(err)
	}

	return r, nil
}
func (self *MongoDAL) UpdateClient(client *model.LogClient) error {
	_, err := clientTable.UpdateOne(context.Background(), bson.M{"id": client.ID}, bson.M{"$set": bson.M{
		"dbpolicy": client.DBPolicy,
	}})
	if err != nil {
		return serr.WithStack(err)
	}

	return nil
}
func (self *MongoDAL) DeleteClient(id string) error {
	_, err := clientTable.DeleteOne(context.Background(), bson.M{"id": id})
	if err != nil {
		return serr.WithStack(err)
	}

	return nil
}

// ************************************************************************************************

func (self *MongoDAL) GetDatabases(clientID string) ([]string, error) {
	return _client.ListDatabaseNames(
		context.Background(),
		// bson.M{"name": bson.M{"$regex": "LOG_" + clientID, "$options": "i"}},
		bson.M{"name": bson.M{"$regex": "LOG_" + clientID}},
		&options.ListDatabasesOptions{NameOnly: &nameOnly},
	)
}
