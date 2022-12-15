package mongodb

import (
	"context"
	"time"

	"github.com/Lukiya/logs/core"
	"github.com/Lukiya/logs/model"
	"github.com/syncfuture/go/sdto"
	"github.com/syncfuture/go/serr"
	"github.com/syncfuture/go/stask"
	"github.com/syncfuture/go/u"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	_nameOnly = true
	// _clients     map[string]*model.LogClient
	// _cacheLocker = new(sync.RWMutex)
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

type MongoDAL struct {
}

func Init() {
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

	// err = refreshCache()
	// u.LogFatal(err)
}

// ************************************************************************************************

// func refreshCache() error {
// 	c, err := _clientTable.Find(nil, bson.M{})
// 	if err != nil {
// 		return serr.WithStack(err)
// 	}

// 	var clients []*model.LogClient
// 	c.All(nil, &clients)

// 	_cacheLocker.Lock()
// 	defer _cacheLocker.Unlock()
// 	_clients = make(map[string]*model.LogClient, len(clients))
// 	for _, x := range clients {
// 		_clients[x.ID] = x
// 	}

// 	return nil
// }

// func (self *MongoDAL) InsertClient(client *model.LogClient) error {
// 	_, err := _clientTable.InsertOne(nil, client)
// 	if err != nil {
// 		return serr.WithStack(err)
// 	}

// 	_cacheLocker.Lock()
// 	defer _cacheLocker.Unlock()
// 	_clients[client.ID] = client

// 	return nil
// }
// func (self *MongoDAL) GetClient(id string) (r *model.LogClient, err error) {
// 	var ok bool
// 	_cacheLocker.RLock()
// 	r, ok = _clients[id]
// 	_cacheLocker.RUnlock()
// 	if ok {
// 		return r, nil
// 	}

// 	rs := _clientTable.FindOne(nil, bson.M{"id": id})
// 	err = rs.Err()
// 	if err != nil {
// 		if err == mongo.ErrNoDocuments {
// 			return nil, nil
// 		}
// 		return nil, serr.WithStack(err)
// 	}

// 	err = rs.Decode(&r)
// 	if err != nil {
// 		return nil, serr.WithStack(err)
// 	}

// 	_cacheLocker.Lock()
// 	_clients[id] = r
// 	_cacheLocker.Unlock()

// 	return r, nil
// }
// func (self *MongoDAL) UpdateClient(client *model.LogClient) error {
// 	_, err := _clientTable.UpdateOne(nil, bson.M{"id": client.ID}, bson.M{"$set": bson.M{
// 		"dbpolicy": client.DBPolicy,
// 	}})
// 	if err != nil {
// 		return serr.WithStack(err)
// 	}

// 	_cacheLocker.Lock()
// 	defer _cacheLocker.Unlock()
// 	_clients[client.ID] = client

// 	return nil
// }
// func (self *MongoDAL) DeleteClient(id string) error {
// 	_, err := _clientTable.DeleteOne(nil, bson.M{"id": id})
// 	if err != nil {
// 		return serr.WithStack(err)
// 	}

// 	_cacheLocker.Lock()
// 	defer _cacheLocker.Unlock()
// 	delete(_clients, id)

// 	return nil
// }
// func (self *MongoDAL) GetClients(query *model.LogClientsQuery) ([]*model.LogClient, error) {
// 	c, err := _clientTable.Find(nil, bson.M{})
// 	if err != nil {
// 		return nil, serr.WithStack(err)
// 	}

// 	var clients []*model.LogClient
// 	err = c.All(nil, &clients)
// 	if err != nil {
// 		return nil, serr.WithStack(err)
// 	}

// 	return clients, nil
// }
// func (self *MongoDAL) RefreshCache() error {
// 	return refreshCache()
// }

// ************************************************************************************************

func (self *MongoDAL) GetDatabases(clientID string) ([]string, error) {
	return _client.ListDatabaseNames(
		nil,
		// bson.M{"name": bson.M{"$regex": "LOG_" + clientID, "$options": "i"}},
		bson.M{"name": bson.M{"$regex": core.LOG_DB_PREFIX + clientID}},
		&options.ListDatabasesOptions{NameOnly: &_nameOnly},
	)
}

func (self *MongoDAL) GetTables(database string) ([]string, error) {
	r, err := _client.Database(database).ListCollectionNames(context.Background(), bson.M{})
	if err != nil {
		return nil, serr.WithStack(err)
	}

	return r, nil
}

func (self *MongoDAL) InsertLogEntry(dbName, tableName string, logEntry *model.LogEntry) error {
	table := _client.Database(dbName).Collection(tableName)
	_, err := table.InsertOne(nil, logEntry)
	if err != nil {
		return serr.WithStack(err)
	}

	return nil
}

func (self *MongoDAL) GetLogEntry(query *model.LogEntryQuery) (*model.LogEntry, error) {
	table := _client.Database(query.DBName).Collection(query.TableName)

	rs := table.FindOne(context.Background(), bson.M{"id": query.ID})
	err := rs.Err()
	if err != nil {
		return nil, serr.WithStack(err)
	}

	var r *model.LogEntry
	err = rs.Decode(&r)
	if err != nil {
		return nil, serr.WithStack(err)
	}

	return r, nil
}

func (self *MongoDAL) GetLogEntries(query *model.LogEntriesQuery) ([]*model.LogEntry, int64, error) {
	table := _client.Database(query.DBName).Collection(query.TableName)
	// Find
	match := bson.M{"$match": bson.M{}}
	matchExp := match["$match"].(bson.M)

	// if query.Keyword != "" {
	// 	matchExp["$or"] = []bson.M{
	// 		{"message": bson.M{"$regex": query.Keyword, "$options": "i"}},
	// 		{"error": bson.M{"$regex": query.Keyword, "$options": "i"}},
	// 	}
	// }

	if query.StartTime != "" {
		t, err := time.ParseInLocation(time.RFC3339, query.StartTime, time.UTC)
		if err != nil {
			return nil, 0, serr.WithStack(err)
		}
		matchExp["createdonutc"] = bson.M{"$gte": t.UnixMilli()}
	}
	if query.EndTime != "" {
		t, err := time.ParseInLocation(time.RFC3339, query.EndTime, time.UTC)
		if err != nil {
			return nil, 0, serr.WithStack(err)
		}
		matchExp["createdonutc"] = bson.M{"$lte": t.UnixMilli()}
	}
	if query.Level >= 0 {
		matchExp["level"] = bson.M{"$eq": query.Level}
	}

	if query.User != "" {
		matchExp["user"] = bson.M{"$regex": query.User, "$options": "i"}
	}
	if query.TraceNo != "" {
		matchExp["traceno"] = bson.M{"$regex": query.TraceNo, "$options": "i"}
	}
	if query.Message != "" {
		matchExp["message"] = bson.M{"$regex": query.Message, "$options": "i"}
	}

	// TotalCount
	count := bson.M{"$count": "totalcount"}

	// Sort
	sortDir := -1
	sort := bson.M{"$sort": bson.M{"createdonutc": sortDir}}

	// Paginate
	limit := bson.M{"$limit": query.PageSize}
	// Skip
	skip := bson.M{"$skip": (query.PageIndex - 1) * query.PageSize}

	// Parallel run
	chrs := _parallel.Invoke(
		func(ch chan *sdto.ChannelResultDTO) {
			chr := &sdto.ChannelResultDTO{Result: 0}
			defer func() {
				ch <- chr
			}()

			countMapPtr := make(map[string]int64)
			var rs *mongo.Cursor
			rs, chr.Error = table.Aggregate(context.Background(), []bson.M{match, count})
			if chr.Error != nil {
				chr.Error = serr.WithStack(chr.Error)
				return
			}

			if rs.TryNext(context.Background()) {
				chr.Error = rs.Decode(&countMapPtr)
				if chr.Error != nil {
					chr.Error = serr.WithStack(chr.Error)
					return
				}
			}

			// if chr.Error != nil && chr.Error != mgo.ErrNotFound {
			totalCount := countMapPtr["totalcount"]
			chr.Result = totalCount
		},
		func(ch chan *sdto.ChannelResultDTO) {
			chr := new(sdto.ChannelResultDTO)
			defer func() {
				ch <- chr
			}()

			r := make([]*model.LogEntry, 0, query.PageSize)
			var rs *mongo.Cursor
			rs, chr.Error = table.Aggregate(context.Background(), []bson.M{match, sort, skip, limit})
			if chr.Error != nil {
				chr.Error = serr.WithStack(chr.Error)
				return
			}
			chr.Error = rs.All(context.Background(), &r)
			if chr.Error != nil {
				chr.Error = serr.WithStack(chr.Error)
				return
			}
			chr.Result = r
		},
	)

	err := u.JointErrors(chrs[0].Error, chrs[1].Error)
	if err != nil {
		return nil, 0, err
	}

	// Get results
	totalCount := chrs[0].Result.(int64)
	list := chrs[1].Result.([]*model.LogEntry)

	return list, totalCount, nil
}
