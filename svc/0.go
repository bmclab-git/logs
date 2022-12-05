package svc

import (
	"github.com/Lukiya/logs/core"
	"github.com/Lukiya/logs/dal"
)

var (
	_logDAL       dal.ILogDAL
	_clientDAL    dal.IClientDAL
	_asyncWriting bool
)

func Init() {
	_asyncWriting = core.GrpcCP.GetBool("AsyncWriting")
	_logDAL = dal.NewLogDAL()
	_clientDAL = dal.NewClientDAL()
}
