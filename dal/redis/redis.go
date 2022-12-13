package redis

import (
	"context"
	"encoding/json"

	"github.com/Lukiya/logs/model"
	r "github.com/go-redis/redis/v8"
	"github.com/syncfuture/go/serr"
	"github.com/syncfuture/go/u"
)

const (
	_KEY = "account:Logs"
)

type RedisDAL struct {
	Client r.UniversalClient
}

func (self *RedisDAL) InsertClient(in *model.LogClient) error {
	return self.UpdateClient(in)
}
func (self *RedisDAL) GetClient(id string) (*model.LogClient, error) {
	rs, err := self.Client.HGet(context.Background(), _KEY, id).Result()
	if err != nil {
		return nil, serr.WithStack(err)
	}

	var r *model.LogClient
	err = json.Unmarshal(u.StrToBytes(rs), &r)
	if err != nil {
		return nil, serr.WithStack(err)
	}

	return r, nil
}
func (self *RedisDAL) UpdateClient(in *model.LogClient) error {
	jsonStr, err := json.Marshal(in)
	if err != nil {
		return serr.WithStack(err)
	}
	_, err = self.Client.HSet(context.Background(), _KEY, in.ID, jsonStr).Result()
	if err != nil {
		return serr.WithStack(err)
	}

	return nil
}
func (self *RedisDAL) DeleteClient(id string) error {
	_, err := self.Client.HDel(context.Background(), _KEY, id).Result()
	if err != nil {
		return serr.WithStack(err)
	}

	return nil
}
func (self *RedisDAL) GetClients(in *model.LogClientsQuery) ([]*model.LogClient, error) {
	rs, err := self.Client.HGetAll(context.Background(), _KEY).Result()
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
