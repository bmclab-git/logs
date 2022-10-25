package mongodb

import (
	"context"
	"sync"

	"github.com/Lukiya/logs/core"
	"github.com/Lukiya/logs/model"
	"github.com/syncfuture/go/sdto"
	"github.com/syncfuture/go/serr"
	"github.com/syncfuture/go/u"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	_nameOnly    = true
	_clients     map[string]*model.LogClient
	_cacheLocker = new(sync.RWMutex)
)

func init() {
	err := refreshCache()
	u.LogFatal(err)
}

type MongoDAL struct {
}

// ************************************************************************************************

func (self *MongoDAL) InsertClient(client *model.LogClient) error {
	_, err := _clientTable.InsertOne(nil, client)
	if err != nil {
		return serr.WithStack(err)
	}

	_cacheLocker.Lock()
	defer _cacheLocker.Unlock()
	_clients[client.ID] = client

	return nil
}
func (self *MongoDAL) GetClient(id string) (r *model.LogClient, err error) {
	var ok bool
	_cacheLocker.RLock()
	r, ok = _clients[id]
	_cacheLocker.RUnlock()
	if ok {
		return r, nil
	}

	rs := _clientTable.FindOne(nil, bson.M{"id": id})
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

	_cacheLocker.Lock()
	_clients[id] = r
	_cacheLocker.Unlock()

	return r, nil
}
func (self *MongoDAL) UpdateClient(client *model.LogClient) error {
	_, err := _clientTable.UpdateOne(nil, bson.M{"id": client.ID}, bson.M{"$set": bson.M{
		"dbpolicy": client.DBPolicy,
	}})
	if err != nil {
		return serr.WithStack(err)
	}

	_cacheLocker.Lock()
	defer _cacheLocker.Unlock()
	_clients[client.ID] = client

	return nil
}
func (self *MongoDAL) DeleteClient(id string) error {
	_, err := _clientTable.DeleteOne(nil, bson.M{"id": id})
	if err != nil {
		return serr.WithStack(err)
	}

	_cacheLocker.Lock()
	defer _cacheLocker.Unlock()
	delete(_clients, id)

	return nil
}
func (self *MongoDAL) GetClients(query *model.LogClientsQuery) (*model.LogClientsResult, error) {
	// 聚集-查找
	match := bson.M{"$match": bson.M{}}
	matchExp := match["$match"].(bson.M)

	if query.Keyword != "" {
		matchExp["$or"] = []bson.M{
			{"id": bson.M{"$regex": query.Keyword, "$options": "i"}},
		}
	}

	// 聚集-总计
	count := bson.M{"$count": "totalcount"}

	// 聚集-排序
	sortDir := -1
	if query.OrderDir == "asc" {
		sortDir = 1
	}
	sort := bson.M{"$sort": bson.M{"id": sortDir}}
	switch query.OrderBy {
	case 1:
		sort["$sort"] = bson.M{"dbpolicy": sortDir}
		break
	}
	// 聚集-限制数量
	limit := bson.M{"$limit": query.PageSize}
	// 聚集-跳过
	skip := bson.M{"$skip": (query.PageIndex - 1) * query.PageSize}

	// 获取结果
	chrs := _parallel.Invoke(
		func(ch chan *sdto.ChannelResultDTO) {
			chr := &sdto.ChannelResultDTO{Result: 0}
			defer func() {
				ch <- chr
			}()

			countMapPtr := make(map[string]int64)
			var rs *mongo.Cursor
			rs, chr.Error = _clientTable.Aggregate(nil, []bson.M{match, count})
			if chr.Error != nil {
				chr.Error = serr.WithStack(chr.Error)
				return
			}

			if rs.TryNext(nil) {
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

			r := make([]*model.LogClient, 0, query.PageSize)
			var rs *mongo.Cursor
			rs, chr.Error = _clientTable.Aggregate(nil, []bson.M{match, sort, skip, limit})
			if chr.Error != nil {
				chr.Error = serr.WithStack(chr.Error)
				return
			}
			chr.Error = rs.All(nil, &r)
			if chr.Error != nil {
				chr.Error = serr.WithStack(chr.Error)
				return
			}
			chr.Result = r
		},
	)

	err := u.JointErrors(chrs[0].Error, chrs[1].Error)
	if err != nil {
		return nil, err
	}

	// 返回结果
	r := new(model.LogClientsResult)
	r.TotalCount = chrs[0].Result.(int64)
	r.LogClients = chrs[1].Result.([]*model.LogClient)
	return r, nil
}
func (self *MongoDAL) RefreshCache() error {
	return refreshCache()
}
func refreshCache() error {
	c, err := _clientTable.Find(nil, bson.M{})
	if err != nil {
		return serr.WithStack(err)
	}

	var clients []*model.LogClient
	c.All(nil, &clients)

	_cacheLocker.Lock()
	defer _cacheLocker.Unlock()
	_clients = make(map[string]*model.LogClient, len(clients))
	for _, x := range clients {
		_clients[x.ID] = x
	}

	return nil
}

// ************************************************************************************************

func (self *MongoDAL) GetDatabases(clientID string) ([]string, error) {
	return _client.ListDatabaseNames(
		nil,
		// bson.M{"name": bson.M{"$regex": "LOG_" + clientID, "$options": "i"}},
		bson.M{"name": bson.M{"$regex": core.LOG_DB_PREFIX + clientID}},
		&options.ListDatabasesOptions{NameOnly: &_nameOnly},
	)
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

	if query.TraceNo != "" {
		matchExp["message"] = bson.M{"$regex": query.TraceNo, "$options": "i"}
	}
	if query.Message != "" {
		matchExp["message"] = bson.M{"$regex": query.Message, "$options": "i"}
	}
	if query.Error != "" {
		matchExp["message"] = bson.M{"$regex": query.Error, "$options": "i"}
	}
	if query.User != "" {
		matchExp["user"] = bson.M{"$regex": query.User, "$options": "i"}
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
