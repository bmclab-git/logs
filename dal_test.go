package main

import (
	"testing"
	"time"

	"github.com/Lukiya/logs/dal"
	"github.com/Lukiya/logs/model"
	"github.com/stretchr/testify/assert"
	"github.com/syncfuture/host"
)

func Test_GetDatabases(t *testing.T) {
	dal := dal.NewLogDAL()
	names, _ := dal.GetDatabases("")
	t.Log(names)
	assert.NotEmpty(t, names)
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
	const id = "DL"
	dal := dal.NewClientDAL()
	err := dal.InsertClient(&model.LogClient{
		ID:       id,
		DBPolicy: 1,
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

	list, totalCount, err := dal.GetClients(&model.LogClientsQuery{
		PageSize:  10,
		PageIndex: 1,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, list)
	assert.Positive(t, totalCount)

	err = dal.DeleteClient(id)
	assert.NoError(t, err)

	client, err = dal.GetClient(id)
	assert.NoError(t, err)
	assert.Nil(t, client)
}
