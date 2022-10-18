package dal

import (
	"github.com/Lukiya/logs/dal/mongodb"
	"github.com/Lukiya/logs/model"
)

type ILogDAL interface {
	GetDatabases(clientID string) ([]string, error)
	InsertLogEntry(dbName, tableName string, logEntry *model.LogEntry) error
}

func NewLogDAL() ILogDAL {
	return new(mongodb.MongoDAL)
}

type IClientDAL interface {
	InsertClient(*model.LogClient) error
	GetClient(id string) (*model.LogClient, error)
	UpdateClient(*model.LogClient) error
	DeleteClient(id string) error
	GetClients(in *model.LogClientsQuery) (*model.LogClientsResult, error)
}

func NewClientDAL() IClientDAL {
	return new(mongodb.MongoDAL)
}
