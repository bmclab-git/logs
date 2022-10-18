package main

import (
	"testing"

	"github.com/Lukiya/logs/dal"
	"github.com/Lukiya/logs/model"
	"github.com/stretchr/testify/assert"
)

func Test_GetDatabases(t *testing.T) {
	dal := dal.NewLogDAL()
	names, _ := dal.GetDatabases("")
	t.Log(names)
	assert.NotEmpty(t, names)
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

	err = dal.DeleteClient(id)
	assert.NoError(t, err)

	client, err = dal.GetClient(id)
	assert.NoError(t, err)
	assert.Nil(t, client)
}
