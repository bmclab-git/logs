package core

import (
	"github.com/syncfuture/go/sconfig"
	"github.com/syncfuture/go/slog"
)

var (
	MainCP sconfig.IConfigProvider
	GrpcCP sconfig.IConfigProvider
)

func init() {
	MainCP = sconfig.NewJsonConfigProvider()
	slog.Init(MainCP)

	GrpcCP = sconfig.NewJsonConfigProvider("grpc.json")
}
