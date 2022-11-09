package svc

import (
	"github.com/Lukiya/logs/core"
	"github.com/Lukiya/logs/dal"
)

var (
	_logDAL       = dal.NewLogDAL()
	_clientDAL    = dal.NewClientDAL()
	_asyncWriting bool
)

func init() {
	_asyncWriting = core.GrpcCP.GetBool("AsyncWriting")
}
