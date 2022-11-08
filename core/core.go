package core

import (
	"github.com/syncfuture/go/sconfig"
	"github.com/syncfuture/go/slog"
)

const (
	LOG_DB_PREFIX = "LOG_"
)

var (
	WebCP  sconfig.IConfigProvider
	GrpcCP sconfig.IConfigProvider
)

func init() {
	GrpcCP = sconfig.NewJsonConfigProvider("grpc.json")
	slog.Init(GrpcCP)

	WebCP = sconfig.NewJsonConfigProvider("web.json")
}
