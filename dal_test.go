package main

import (
	"testing"
	"time"

	"github.com/Lukiya/logs/core"
	"github.com/Lukiya/logs/dal"
	"github.com/Lukiya/logs/model"
	"github.com/stretchr/testify/assert"
	"github.com/syncfuture/host"
)

func init() {
	core.Init()
}

func Test_GetArchives(t *testing.T) {
	clientDAL := dal.NewClientDAL()
	logDAL := dal.NewLogDAL()
	clients, err := clientDAL.GetClients(new(model.LogClientsQuery))
	assert.NoError(t, err)
	assert.NotEmpty(t, clients)
	t.Log(clients)

	dbs, err := logDAL.GetDatabases(clients[0].ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, dbs)
	t.Log(dbs)

	tables, err := logDAL.GetTables(dbs[0])
	assert.NoError(t, err)
	assert.NotEmpty(t, tables)
	t.Log(tables)
}

func Test_Log(t *testing.T) {
	dal := dal.NewLogDAL()
	id := host.GenerateID()
	db := "LOG_DL"
	table := "2022"

	err := dal.InsertLogEntry(db, table, &model.LogEntry{
		ID:           id,
		TraceNo:      "AAAA",
		User:         "BBBB",
		Message:      "CCCC",
		Error:        "DDDD",
		StackTrace:   "EEEE",
		Payload:      "FFFF",
		Level:        model.LogLevel_Debug,
		CreatedOnUtc: time.Now().UTC().UnixMilli(),
	})
	assert.NoError(t, err)

	logEntry, err := dal.GetLogEntry(&model.LogEntryQuery{
		DBName:    db,
		TableName: table,
		ID:        id,
	})
	assert.NoError(t, err)
	assert.NotNil(t, logEntry)

	list, totalCount, err := dal.GetLogEntries(&model.LogEntriesQuery{
		DBName:    db,
		TableName: table,
		PageSize:  10,
		PageIndex: 1,
	})
	assert.NoError(t, err)
	assert.Positive(t, totalCount)
	assert.NotEmpty(t, list)
}

func Test_Client(t *testing.T) {
	const id = "OLX"
	dal := dal.NewClientDAL()
	err := dal.InsertClient(&model.LogClient{
		ID:       id,
		DBPolicy: 1,
		Level:    model.LogLevel_Infomation,
	})
	assert.NoError(t, err)

	client, err := dal.GetClient(id)
	assert.NoError(t, err)
	assert.Equal(t, int32(1), client.DBPolicy)

	client.DBPolicy = 3
	err = dal.UpdateClient(client)
	assert.NoError(t, err)

	client, err = dal.GetClient(id)
	assert.NoError(t, err)
	assert.Equal(t, int32(3), client.DBPolicy)

	list, err := dal.GetClients(&model.LogClientsQuery{})
	assert.NoError(t, err)
	assert.NotEmpty(t, list)

	err = dal.DeleteClient(id)
	assert.NoError(t, err)

	client, err = dal.GetClient(id)
	assert.NoError(t, err)
	assert.Nil(t, client)
}

func TestGetLogEntries(t *testing.T) {
	const id = "DL"
	db := "LOG_DL"
	table := "2022"

	dal := dal.NewLogDAL()

	list, totalCount, err := dal.GetLogEntries(&model.LogEntriesQuery{
		DBName:    db,
		TableName: table,
		PageSize:  10,
		PageIndex: 1,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, list)
	assert.Positive(t, totalCount)
}
