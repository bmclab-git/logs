package dal

import (
	"github.com/Lukiya/logs/dal/mysql"
	"github.com/Lukiya/logs/model"
)

type ILogDAL interface {
	InsertLogEntry(dbName, tableName string, logEntry *model.LogEntry) error
	GetDatabases(clientID string) ([]string, error)
	GetLogEntry(query *model.LogEntryQuery) (*model.LogEntry, error)
	GetLogEntries(query *model.LogEntriesQuery) ([]*model.LogEntry, int64, error)
}

func NewLogDAL() ILogDAL {
	return new(mysql.MySqlDAL)
}

type IClientDAL interface {
	InsertClient(*model.LogClient) error
	GetClient(id string) (*model.LogClient, error)
	UpdateClient(*model.LogClient) error
	DeleteClient(id string) error
	GetClients(in *model.LogClientsQuery) ([]*model.LogClient, int64, error)
}

func NewClientDAL() IClientDAL {
	return new(mysql.MySqlDAL)
}
