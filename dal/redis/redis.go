package redis

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/Lukiya/logs/core"
	"github.com/Lukiya/logs/model"
	goredis "github.com/go-redis/redis/v8"
	"github.com/syncfuture/go/sconfig"
	"github.com/syncfuture/go/serr"
	"github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/sredis"
	"github.com/syncfuture/go/u"
)

const (
	_KEY = "account:Logs"
)

var (
	_clientsMap  map[string]*model.LogClient
	_cacheLocker = new(sync.RWMutex)
)

type RedisDAL struct {
	client goredis.UniversalClient
}

func (self *RedisDAL) Init(config sconfig.IConfigProvider) {
	var rc *sredis.RedisConfig
	core.GrpcCP.GetStruct("Redis", &rc)
	self.client = sredis.NewClient(rc)

	err := self.refreshCache()
	u.LogFatal(err)

	go self.monitor()
}

func (self *RedisDAL) monitor() {
	// Subscribe key changes
	sub := self.client.Subscribe(context.Background(), "__keyspace@0__:"+_KEY)

	// refresh cache when there's a change
	for {
		msg := <-sub.Channel()
		self.refreshCache()
		slog.Debugf("Cache refreshed (Channel:'%s', Pattern:'%s', Payload:'%s', PayloadSlice:%v)", msg.Channel, msg.Pattern, msg.Payload, msg.PayloadSlice)
	}
}
func (self *RedisDAL) refreshCache() error {
	_cacheLocker.Lock()
	defer _cacheLocker.Unlock()

	var clients []*model.LogClient
	clients, err := self.GetClients(nil)
	if err != nil {
		return serr.WithStack(err)
	}

	_clientsMap = make(map[string]*model.LogClient, len(clients))
	for _, x := range clients {
		_clientsMap[x.ID] = x
	}

	return nil
}

func (self *RedisDAL) InsertClient(in *model.LogClient) error {
	return self.UpdateClient(in)
}
func (self *RedisDAL) GetClient(id string) (r *model.LogClient, err error) {
	var ok bool
	_cacheLocker.RLock()
	r, ok = _clientsMap[id]
	_cacheLocker.RUnlock()
	if ok {
		return r, nil
	}

	r = new(model.LogClient)
	rs, err := self.client.HGet(context.Background(), _KEY, id).Result()
	if err != nil {
		if err == goredis.Nil {
			return nil, nil
		}

		return nil, serr.WithStack(err)
	}

	err = json.Unmarshal(u.StrToBytes(rs), &r)
	if err != nil {
		return nil, serr.WithStack(err)
	}

	_cacheLocker.Lock()
	_clientsMap[id] = r
	_cacheLocker.Unlock()

	return r, nil
}
func (self *RedisDAL) UpdateClient(in *model.LogClient) error {
	jsonStr, err := json.Marshal(in)
	if err != nil {
		return serr.WithStack(err)
	}
	_, err = self.client.HSet(context.Background(), _KEY, in.ID, jsonStr).Result()
	if err != nil {
		return serr.WithStack(err)
	}

	_cacheLocker.Lock()
	defer _cacheLocker.Unlock()
	_clientsMap[in.ID] = in

	return nil
}
func (self *RedisDAL) DeleteClient(id string) error {
	_, err := self.client.HDel(context.Background(), _KEY, id).Result()
	if err != nil {
		return serr.WithStack(err)
	}

	_cacheLocker.Lock()
	defer _cacheLocker.Unlock()
	delete(_clientsMap, id)

	return nil
}
func (self *RedisDAL) GetClients(*model.LogClientsQuery) ([]*model.LogClient, error) {
	rs, err := self.client.HGetAll(context.Background(), _KEY).Result()
	if err != nil {
		return nil, serr.WithStack(err)
	}

	r := make([]*model.LogClient, 0, len(rs))
	for _, x := range rs {
		var client *model.LogClient
		err = json.Unmarshal(u.StrToBytes(x), &client)
		if err != nil {
			return nil, serr.WithStack(err)
		}
		r = append(r, client)
	}

	return r, nil
}
