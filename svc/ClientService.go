package svc

import (
	"context"

	"github.com/Lukiya/logs/model"
	"github.com/syncfuture/go/u"
)

type LogClientService struct{}

func (self *LogClientService) GetClients(ctx context.Context, in *model.LogClientsQuery) (*model.LogClientsResult, error) {
	r := new(model.LogClientsResult)

	list, err := _clientDAL.GetClients(in)
	if u.LogError(err) {
		r.Message = err.Error()
	}
	r.LogClients = make([]string, 0, len(list))
	for _, x := range list {
		r.LogClients = append(r.LogClients, x.ID)
	}
	return r, nil
}
func (self *LogClientService) GetDatabases(ctx context.Context, in *model.DatabasesQuery) (*model.DatabasesResult, error) {
	r := new(model.DatabasesResult)

	list, err := _logDAL.GetDatabases(in.ClientID)
	if u.LogError(err) {
		r.Message = err.Error()
	}

	r.Databases = list
	return r, nil
}
func (self *LogClientService) GetTables(ctx context.Context, in *model.TablesQuery) (*model.TablesResult, error) {
	r := new(model.TablesResult)
	list, err := _logDAL.GetTables(in.Database)
	if u.LogError(err) {
		r.Message = err.Error()
	}

	r.Tables = list
	return r, nil
}
